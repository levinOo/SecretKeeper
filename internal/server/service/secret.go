package service

import (
	"context"
	"io"
	pb "secretKeeper/internal/server/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Функция создания новой секретной записи и загрузки данных
func (s *Server) CreateSecret(stream pb.ProjectService_CreateSecretServer) error {
	ctx := stream.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return status.Error(codes.Unauthenticated, "user not found in context")
	}
	transactionID, ok := ctx.Value("idempotency-key").(string)
	if !ok {
		return status.Error(codes.InvalidArgument, "idempotency-key not found in context")
	}

	var (
		meta *pb.SecretMetadata

		// For MinIO Storage (Unified)
		pipeReader *io.PipeReader
		pipeWriter *io.PipeWriter
		uploadErr  chan error
		fileSize   int64
	)

	// Ensure writer is closed if we exit early, to unblock reader
	defer func() {
		if pipeWriter != nil {
			// Close with error to stop the reader if it's still reading
			_ = pipeWriter.CloseWithError(io.ErrUnexpectedEOF)
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// --- Stream Finished ---
			if meta == nil {
				return status.Error(codes.InvalidArgument, "no metadata received")
			}

			if pipeWriter == nil {
				return status.Error(codes.InvalidArgument, "no chunks data received")
			}

			// Close writer to signal EOF to UploadFile
			_ = pipeWriter.Close()
			pipeWriter = nil // Avoid defer close with error

			// Wait for upload result
			if err := <-uploadErr; err != nil {
				s.logger.Errorf("minio upload failed: %v", err)
				return status.Errorf(codes.Internal, "storage error")
			}

			// Save to DB (Storage Path = MinIO Bucket/Key = transactionID)
			storageKey := transactionID
			err = s.db.SaveSecret(ctx, userID, transactionID, int(meta.Type), meta.Name, meta.PublicMeta, storageKey)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to save secret metadata: %v", err)
			}

			return stream.SendAndClose(&pb.CreateSecretResponse{
				Id:     transactionID,
				Status: "OK",
			})
		}

		if err != nil {
			return status.Errorf(codes.Unknown, "stream receive error: %v", err)
		}

		// --- Handle Payload ---
		switch payload := req.Data.(type) {
		case *pb.CreateSecretRequest_Metadata:
			if meta != nil {
				return status.Error(codes.InvalidArgument, "metadata sent twice")
			}
			meta = payload.Metadata
			s.logger.Infof("Receiving secret type=%s from user=%s (tx=%s)", meta.Type, userID, transactionID)

			// Idempotency Check
			exists, err := s.db.CheckSecretExists(transactionID)
			if err != nil {
				return status.Errorf(codes.Internal, "idempotency check failed: %v", err)
			}
			if exists {
				s.logger.Infof("Secret %s already exists, skipping upload", transactionID)
				return stream.SendAndClose(&pb.CreateSecretResponse{
					Id:     transactionID,
					Status: "ALREADY_EXISTS",
				})
			}

			// Init MinIO Stream (ALWAYS, for all types)
			pipeReader, pipeWriter = io.Pipe()
			uploadErr = make(chan error, 1)

			go func() {
				defer close(uploadErr)
				// ContentType can be detected or passed in meta.
				// For File -> application/octet-stream or meta detection
				// For Text/JSON -> maybe text/plain or application/json?
				// For now, robust default is application/octet-stream
				err := s.storage.UploadFile(context.Background(), transactionID, pipeReader, -1, "application/octet-stream") // c типом разобраться
				uploadErr <- err
			}()

		case *pb.CreateSecretRequest_ChunkData:
			if meta == nil {
				// According to gRPC best practices, metadata should be first.
				// But we handle it if client sends chunks out of order? No, strict order is better.
				return status.Error(codes.InvalidArgument, "chunk received before metadata")
			}

			// Write to Pipe -> MinIO
			if pipeWriter == nil {
				return status.Error(codes.Internal, "pipe not initialized")
			}

			n, err := pipeWriter.Write(payload.ChunkData)
			if err != nil {
				return status.Errorf(codes.Internal, "write to pipe failed: %v", err)
			}
			fileSize += int64(n)
			if fileSize > s.config.Storage.MaxFileSize {
				return status.Errorf(codes.ResourceExhausted, "file size limit exceeded")
			}

		default:
			return status.Errorf(codes.InvalidArgument, "unknown payload type")
		}
	}
}

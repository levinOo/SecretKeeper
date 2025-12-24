package service

import (
	"context"
	"io"
	pb "secretKeeper/internal/server/transport/grpc/proto"
	"secretKeeper/pkg/server/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SecretRepository interface {
	CheckSecretExists(transactionID string) (bool, error)
	SaveSecret(ctx context.Context, userID, secretID string, secretType int32, title string, meta *pb.SecretMetadata, storagePath string) error
	GetSecret(ctx context.Context, userID, secretID string) (string, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
}

type SecretService struct {
	repo    SecretRepository
	storage *storage.MinioStorage
}

func NewSecretService(repo SecretRepository, stor *storage.MinioStorage) *SecretService {
	return &SecretService{
		repo:    repo,
		storage: stor,
	}
}

// Методы реализации интерфейса Secrets
func (s *SecretService) CreateSecret(stream pb.ProjectService_CreateSecretServer) (string, string, error) {
	ctx := stream.Context()

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", "", status.Error(codes.Unauthenticated, "user not found in context")
	}

	transactionID, ok := ctx.Value("idempotency-key").(string)
	if !ok {
		return "", "", status.Error(codes.InvalidArgument, "idempotency-key not found in context")
	}

	var (
		meta *pb.SecretMetadata

		pipeReader *io.PipeReader
		pipeWriter *io.PipeWriter
		uploadErr  chan error
		fileSize   int64
	)

	defer func() {
		if pipeWriter != nil {
			_ = pipeWriter.CloseWithError(io.ErrUnexpectedEOF)
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			if meta == nil {
				return "", "", status.Error(codes.InvalidArgument, "no metadata received")
			}

			if pipeWriter == nil {
				return "", "", status.Error(codes.InvalidArgument, "no chunks data received")
			}

			_ = pipeWriter.Close()
			pipeWriter = nil

			if err := <-uploadErr; err != nil {
				return "", "", status.Errorf(codes.Internal, "storage error")
			}

			// Save to DB (Storage Path = MinIO Bucket/Key = transactionID)
			storageKey := transactionID
			err := s.repo.SaveSecret(ctx, userID, transactionID, meta.Type, meta.Title, meta, storageKey)
			if err != nil {
				return "", "", status.Errorf(codes.Internal, "failed to save secret metadata: %v", err)
			}

			return transactionID, "OK", nil
		}

		if err != nil {
			return "", "", status.Errorf(codes.Unknown, "stream receive error: %v", err)
		}

		// --- Handle Payload ---
		switch payload := req.Data.(type) {
		case *pb.CreateSecretRequest_Metadata:
			if meta != nil {
				return "", "", status.Error(codes.InvalidArgument, "metadata sent twice")
			}
			meta = payload.Metadata

			// Idempotency Check
			exists, err := s.repo.CheckSecretExists(transactionID)
			if err != nil {
				return "", "", status.Errorf(codes.Internal, "idempotency check failed: %v", err)
			}
			if !exists {
				return transactionID, "ALREADY_EXISTS", nil
			}

			// Init MinIO Stream (ALWAYS, for all types)
			pipeReader, pipeWriter = io.Pipe()
			uploadErr = make(chan error, 1)

			go func() {
				defer close(uploadErr)
				err := s.storage.UploadFile(context.Background(), transactionID, pipeReader, -1, "application/octet-stream") // c типом разобраться
				uploadErr <- err
			}()

		case *pb.CreateSecretRequest_ChunkData:
			if meta == nil {
				return "", "", status.Error(codes.InvalidArgument, "chunk received before metadata")
			}

			// Write to Pipe -> MinIO
			if pipeWriter == nil {
				return "", "", status.Error(codes.Internal, "pipe not initialized")
			}

			n, err := pipeWriter.Write(payload.ChunkData)
			if err != nil {
				return "", "", status.Errorf(codes.Internal, "write to pipe failed: %v", err)
			}
			fileSize += int64(n)
			if fileSize > s.config.Storage.MaxFileSize {
				return "", "", status.Errorf(codes.ResourceExhausted, "file size limit exceeded")
			}

		default:
			return "", "", status.Errorf(codes.InvalidArgument, "unknown payload type")
		}
	}
}

func (s *SecretService) GetSecret(ctx context.Context, userID, secretID string) (string, error) {
	return s.repo.GetSecret(ctx, userID, secretID)
}

func (s *SecretService) DeleteSecret(ctx context.Context, userID, secretID string) error {
	return s.repo.DeleteSecret(ctx, userID, secretID)
}

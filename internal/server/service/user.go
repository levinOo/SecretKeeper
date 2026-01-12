package service

import (
	"context"
	"io"
	pb "secretKeeper/internal/proto"
	"secretKeeper/internal/server/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxFileSize = 1024 * 1024 * 5 // 5MB
	bufferSize  = 64 * 1024       // 64KB
)

// UserRepository интерфейс репозитория пользователя
type UserRepository interface {
	SaveSecret(ctx context.Context, userID, secretID, idempotencyKey string, meta *pb.SecretMetadata) error
	CheckSecretExists(transactionID, userID string) (bool, error)
	ListUserFiles(ctx context.Context, userID string) ([]domain.File, error)
	FileExists(ctx context.Context, userID, secretID string) (bool, error)
	DeleteSecret(ctx context.Context, userID, secretID string) error
	GetSecretMeta(ctx context.Context, userID, secretID string) (*pb.SecretMetadata, error)
}

// FileRepository интерфейс файлового репозитория
type FileRepository interface {
	UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error
	DeleteSecret(ctx context.Context, secretID string) error
	GetSecret(ctx context.Context, secretID string) (io.Reader, error)
}

// UserService сервис пользователя
type UserService struct {
	repo   UserRepository
	file   FileRepository
	logger *zap.SugaredLogger
}

// NewUserService создает новый сервис пользователя
func NewUserService(repo UserRepository, file FileRepository, logger *zap.SugaredLogger) *UserService {
	return &UserService{repo: repo, file: file, logger: logger}
}

// CreateSecret создает секрет
func (s *UserService) CreateSecret(stream pb.SecretService_CreateSecretServer) (string, string, error) {
	ctx := stream.Context()

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return "", "", status.Error(codes.Unauthenticated, "пользователь не найден в контексте")
	}

	transactionID, ok := ctx.Value("idempotency-key").(string)
	if !ok || transactionID == "" {
		return "", "", status.Error(codes.InvalidArgument, "ключ идемпотентности не найден в контексте")
	}

	exist, err := s.repo.CheckSecretExists(transactionID, userID)
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "не удалось проверить существование секрета: %v", err)
	}
	if exist {
		return transactionID, "ALREADY_EXISTS", nil
	}

	secretID := uuid.New().String()

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
				return "", "", status.Error(codes.InvalidArgument, "метаданные не получены")
			}

			if pipeWriter == nil {
				return "", "", status.Error(codes.InvalidArgument, "данные секрета не получены")
			}

			_ = pipeWriter.Close()
			pipeWriter = nil

			if err := <-uploadErr; err != nil {
				return "", "", status.Errorf(codes.Internal, "ошибка хранилища: %v", err)
			}

			err := s.repo.SaveSecret(context.Background(), userID, secretID, transactionID, meta)
			if err != nil {
				return "", "", status.Errorf(codes.Internal, "не удалось сохранить метаданные секрета: %v", err)
			}

			return secretID, "OK", nil
		}

		if err != nil {
			return "", "", status.Errorf(codes.Unknown, "ошибка получения потока: %v", err)
		}

		switch payload := req.Data.(type) {
		case *pb.CreateSecretRequest_Meta:
			if meta != nil {
				return "", "", status.Error(codes.InvalidArgument, "метаданные отправлены дважды")
			}
			meta = payload.Meta

			pipeReader, pipeWriter = io.Pipe()
			uploadErr = make(chan error, 1)

			go func() {
				defer close(uploadErr)
				err := s.file.UploadFile(context.Background(), secretID, pipeReader, -1, "application/octet-stream")
				uploadErr <- err
			}()

		case *pb.CreateSecretRequest_ChunkData:
			if meta == nil {
				return "", "", status.Error(codes.InvalidArgument, "фрагмент получен до метаданных")
			}

			if pipeWriter == nil {
				return "", "", status.Error(codes.Internal, "канал не инициализирован")
			}

			n, err := pipeWriter.Write(payload.ChunkData)
			if err != nil {
				return "", "", status.Errorf(codes.Internal, "ошибка записи в канал: %v", err)
			}
			fileSize += int64(n)
			if fileSize > maxFileSize {
				return "", "", status.Errorf(codes.ResourceExhausted, "превышен лимит размера файла")
			}

		default:
			return "", "", status.Errorf(codes.InvalidArgument, "неизвестный тип нагрузки")
		}
	}
}

// ListSecrets возвращает список секретов
func (s *UserService) ListSecrets(ctx context.Context, req *pb.ListSecretsRequest) ([]*pb.SecretMetadata, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "пользователь не найден в контексте")
	}

	files, err := s.repo.ListUserFiles(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить список секретов: %v", err)
	}

	pbSecrets := make([]*pb.SecretMetadata, 0, len(files))

	for _, f := range files {

		meta := &pb.SecretMetadata{
			Id:    f.ID,
			Type:  f.Meta.SecretType,
			Extra: f.Meta.Extra,
		}

		pbSecrets = append(pbSecrets, meta)
	}

	return pbSecrets, nil
}

// DeleteSecret удаляет секрет
func (s *UserService) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) error {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "пользователь не найден в контексте")
	}

	exist, err := s.repo.FileExists(ctx, userID, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось проверить существование секрета: %v", err)
	}
	if !exist {
		return status.Error(codes.NotFound, "секрет не найден")
	}

	err = s.file.DeleteSecret(ctx, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось удалить секрет из хранилища: %v", err)
	}

	err = s.repo.DeleteSecret(ctx, userID, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось удалить секрет: %v", err)
	}

	return nil
}

// GetSecret получает секрет
func (s *UserService) GetSecret(req *pb.GetSecretRequest, stream pb.SecretService_GetSecretServer) error {
	ctx := stream.Context()

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return status.Error(codes.Unauthenticated, "пользователь не найден в контексте")
	}

	exist, err := s.repo.FileExists(ctx, userID, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось проверить существование секрета: %v", err)
	}
	if !exist {
		return status.Error(codes.NotFound, "секрет не найден")
	}

	fileMeta, err := s.repo.GetSecretMeta(ctx, userID, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось получить метаданные файла: %v", err)
	}

	err = stream.Send(&pb.GetSecretResponse{
		Meta: fileMeta,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось отправить метаданные: %v", err)
	}

	contentReader, err := s.file.GetSecret(ctx, req.Id)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось получить файл из хранилища: %v", err)
	}
	defer func() {
		if closer, ok := contentReader.(io.Closer); ok {
			_ = closer.Close()
		}
	}()

	buffer := make([]byte, bufferSize)
	for {
		n, err := contentReader.Read(buffer)
		if err != nil && err != io.EOF {
			return status.Errorf(codes.Internal, "не удалось прочитать содержимое файла: %v", err)
		}

		if n > 0 {
			err = stream.Send(&pb.GetSecretResponse{
				ChunkData: buffer[:n],
			})
			if err != nil {
				return status.Errorf(codes.Internal, "не удалось отправить фрагмент данных: %v", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

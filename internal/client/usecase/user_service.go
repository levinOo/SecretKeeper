package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"secretKeeper/internal/client/domain"
	"secretKeeper/pkg/client/fileutil"
	storage "secretKeeper/pkg/client/token_storage"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserProvider интерфейс для работы с секретами
type UserProvider interface {
	CreateSecret(ctx context.Context, secret *domain.Secret) (id, status string, err error)
	GetSecret(ctx context.Context, id string) (*domain.Secret, error)
	DeleteSecret(ctx context.Context, id string) (status string, err error)
	GetList(ctx context.Context) ([]domain.SecretMeta, error)
}

// CryptoProvider интерфейс для шифрования
type CryptoProvider interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
	EncryptStream(in io.Reader, out io.Writer) error
}

// CacheProvider интерфейс для кеширования
type CacheProvider interface {
	Has(id string) bool
	Get(id string) (*os.File, *domain.SecretMeta, error)
	Put(id string, meta domain.SecretMeta) (*os.File, error)
}

// UserInteractor интерфейс для взаимодействия с пользователем
type UserInteractor interface {
	ConfirmCacheUsage(secretID string) (bool, error)
}

// UserService сервис для работы с пользователем и секретами
type UserService struct {
	tokenStorage storage.TokenRepository
	client       UserProvider
	crypto       CryptoProvider
	cache        CacheProvider
	interactor   UserInteractor
}

// NewUserService создает новый сервис пользователя
func NewUserService(tokenStorage storage.TokenRepository, client UserProvider, crypto CryptoProvider, cache CacheProvider) *UserService {
	return &UserService{
		tokenStorage: tokenStorage,
		client:       client,
		crypto:       crypto,
		cache:        cache,
		interactor:   nil, // Будет установлен позже через SetInteractor
	}
}

// SetInteractor устанавливает интерактор для взаимодействия с пользователем
func (s *UserService) SetInteractor(interactor UserInteractor) {
	s.interactor = interactor
}

// CreateCredentialsSecret создает секрет с логином и паролем
func (s *UserService) CreateCredentialsSecret(ctx context.Context, name, description, login, password, url string) error {
	content := fmt.Sprintf(`{"login":"%s","password":"%s","url":"%s"}`, login, password, url)

	encryptedData, err := s.crypto.Encrypt([]byte(content))
	if err != nil {
		return fmt.Errorf("не удалось зашифровывать секрет: %w", err)
	}

	secret := &domain.Secret{
		Meta: domain.SecretMeta{
			Type: domain.SecretTypeCredentials,
			Extra: map[string]string{
				"name":        name,
				"description": description,
			},
		},
		Data: io.NopCloser(bytes.NewReader(encryptedData)),
	}

	idempotencyKey := uuid.New().String()
	md := metadata.Pairs(
		"idempotency-key", idempotencyKey,
		"authorization", "Bearer "+s.tokenStorage.GetToken(),
	)

	ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

	if _, _, err = s.client.CreateSecret(ctxWithMeta, secret); err != nil {
		return fmt.Errorf("не удалось создать секрет: %w", err)
	}

	return nil
}

// CreateFileSecret создает секрет из файла
func (s *UserService) CreateFileSecret(ctx context.Context, name, description string, file *os.File) error {
	// 1. Validate size
	if err := fileutil.CheckSize(file); err != nil {
		return err
	}

	if err := fileutil.DetectAndValidate(file); err != nil {
		return err
	}

	pr, pw := io.Pipe()

	go func() {
		if err := s.crypto.EncryptStream(file, pw); err != nil {
			pw.CloseWithError(fmt.Errorf("шифрование не удалась: %w", err))
		} else {
			pw.Close()
		}
	}()

	secret := &domain.Secret{
		Meta: domain.SecretMeta{
			Type: domain.SecretTypeFile,
			Extra: map[string]string{
				"name":        name,
				"filename":    file.Name(),
				"size":        fmt.Sprintf("%d", fileutil.GetFileSize(file)),
				"description": description,
			},
		},
		Data: pr,
	}

	idempotencyKey := uuid.New().String()
	md := metadata.Pairs(
		"idempotency-key", idempotencyKey,
		"authorization", "Bearer "+s.tokenStorage.GetToken(),
	)
	ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

	if _, _, err := s.client.CreateSecret(ctxWithMeta, secret); err != nil {
		pr.Close()
		return fmt.Errorf("не удалось создать секрет: %w", err)
	}

	return nil
}

// GetSecret получает секрет по ID
func (s *UserService) GetSecret(ctx context.Context, id string) (*domain.Secret, error) {
	var file *os.File
	var meta *domain.SecretMeta
	var err error

	if s.cache.Has(id) {
		file, meta, err = s.cache.Get(id)
	} else {
		var secret *domain.Secret
		secret, err = s.downloadAndCacheSecret(ctx, id)
		if err == nil {
			if f, ok := secret.Data.(*os.File); ok {
				file = f
				meta = &secret.Meta
			} else {
				err = fmt.Errorf("внутренняя ошибка: данные секрета не являются файлом")
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("ошибка поиска в файле: %w", err)
	}

	encryptedData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать зашифрованные данные: %w", err)
	}

	decryptedData, err := s.crypto.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать секрет: %w", err)
	}

	return &domain.Secret{
		Meta: *meta,
		Data: io.NopCloser(bytes.NewReader(decryptedData)),
	}, nil
}

// GetSecretWithOfflineFallback получает секрет с поддержкой оффлайн режима
// Если сервер недоступен и секрет есть в кеше, предлагает пользователю использовать кешированную версию
func (s *UserService) GetSecretWithOfflineFallback(ctx context.Context, id string) (*domain.Secret, error) {
	// Сначала пытаемся получить с сервера
	var file *os.File
	var meta *domain.SecretMeta
	var err error

	// Проверяем кеш
	hasCache := s.cache.Has(id)

	// Пытаемся загрузить с сервера
	var secret *domain.Secret
	secret, err = s.downloadAndCacheSecret(ctx, id)

	if err == nil {
		// Успешно загрузили с сервера
		if f, ok := secret.Data.(*os.File); ok {
			file = f
			meta = &secret.Meta
		} else {
			return nil, fmt.Errorf("внутренняя ошибка: данные секрета не являются файлом")
		}
	} else {
		// Ошибка при загрузке с сервера
		// Проверяем, является ли это сетевой ошибкой
		if isNetworkError(err) && hasCache {
			// Сервер недоступен, но есть кеш
			// Спрашиваем пользователя, хочет ли он использовать кеш
			if s.interactor != nil {
				useCache, confirmErr := s.interactor.ConfirmCacheUsage(id)
				if confirmErr != nil {
					return nil, fmt.Errorf("ошибка при запросе подтверждения: %w", confirmErr)
				}

				if !useCache {
					return nil, errors.New("операция отменена пользователем")
				}

				// Пользователь согласился использовать кеш
				file, meta, err = s.cache.Get(id)
				if err != nil {
					return nil, fmt.Errorf("не удалось получить секрет из кеша: %w", err)
				}
			} else {
				// Нет интерактора, используем кеш автоматически
				file, meta, err = s.cache.Get(id)
				if err != nil {
					return nil, fmt.Errorf("не удалось получить секрет из кеша: %w", err)
				}
			}
		} else {
			// Либо не сетевая ошибка, либо нет кеша
			return nil, fmt.Errorf("не удалось получить секрет: %w", err)
		}
	}

	// Read encrypted data
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("ошибка поиска в файле: %w", err)
	}

	encryptedData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать зашифрованные данные: %w", err)
	}

	decryptedData, err := s.crypto.Decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать секрет: %w", err)
	}

	return &domain.Secret{
		Meta: *meta,
		Data: io.NopCloser(bytes.NewReader(decryptedData)),
	}, nil
}

// isNetworkError проверяет, является ли ошибка сетевой (недоступность сервера, timeout и т.д.)
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Проверяем gRPC статус коды
	if st, ok := status.FromError(err); ok {
		code := st.Code()
		return code == codes.Unavailable ||
			code == codes.DeadlineExceeded ||
			code == codes.Canceled ||
			code == codes.Unknown
	}

	// Проверяем текст ошибки на наличие ключевых слов
	errMsg := strings.ToLower(err.Error())
	networkKeywords := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"timeout",
		"deadline exceeded",
		"unavailable",
		"context canceled",
		"context deadline exceeded",
	}

	for _, keyword := range networkKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

// DeleteSecret удаляет секрет
func (s *UserService) DeleteSecret(ctx context.Context, id string) (string, error) {
	md := metadata.Pairs(
		"authorization", "Bearer "+s.tokenStorage.GetToken(),
	)

	ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

	status, err := s.client.DeleteSecret(ctxWithMeta, id)
	if err != nil {
		return "", err
	}

	return status, nil
}

// GetList возвращает список секретов
func (s *UserService) GetList(ctx context.Context) ([]domain.SecretMeta, error) {
	md := metadata.Pairs(
		"authorization", "Bearer "+s.tokenStorage.GetToken(),
	)

	ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

	meta, err := s.client.GetList(ctxWithMeta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (s *UserService) downloadAndCacheSecret(ctx context.Context, id string) (*domain.Secret, error) {
	md := metadata.Pairs(
		"authorization", "Bearer "+s.tokenStorage.GetToken(),
	)
	ctxWithMeta := metadata.NewOutgoingContext(ctx, md)

	secret, err := s.client.GetSecret(ctxWithMeta, id)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить секрет: %w", err)
	}

	tmpFile, err := fileutil.NewTempFile("secret-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("не удалось создать временный файл: %w", err)
	}

	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}
	}()

	if _, err := io.Copy(tmpFile, secret.Data); err != nil {
		return nil, fmt.Errorf("не удалось сохранить данные секрета: %w", err)
	}

	secret.Data.Close()

	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("ошибка поиска в файле: %w", err)
	}

	cachedSecret := &domain.Secret{
		Meta: secret.Meta,
		Data: tmpFile,
	}

	success = true
	return cachedSecret, nil
}

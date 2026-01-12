package myGrpc

import (
	"context"
	"fmt"
	"io"
	"secretKeeper/internal/client/domain"
	pb "secretKeeper/internal/proto"
)

// UserClient клиент пользователя
type UserClient struct {
	client pb.SecretServiceClient
}

// NewUserClient создает новый UserClient
func NewUserClient(client pb.SecretServiceClient) *UserClient {
	return &UserClient{
		client: client,
	}
}

// CreateSecret создает секрет
func (c *UserClient) CreateSecret(ctx context.Context, secret *domain.Secret) (string, string, error) {
	stream, err := c.client.CreateSecret(ctx)
	if err != nil {
		return "", "", fmt.Errorf("не удалось создать поток: %w", err)
	}

	// Отправляем метаданные первым сообщением
	metadata := &pb.SecretMetadata{
		Type:  string(secret.Meta.Type),
		Extra: secret.Meta.Extra,
	}

	metadataReq := &pb.CreateSecretRequest{
		Data: &pb.CreateSecretRequest_Meta{
			Meta: metadata,
		},
	}

	if err := stream.Send(metadataReq); err != nil {
		return "", "", fmt.Errorf("не удалось отправить метаданные: %w", err)
	}

	// Создаем буфер фиксированного размера (64KB)
	chunk := make([]byte, 64*1024)

	for {
		n, err := secret.Data.Read(chunk)
		if n > 0 {
			chunkReq := &pb.CreateSecretRequest{
				Data: &pb.CreateSecretRequest_ChunkData{
					ChunkData: chunk[:n],
				},
			}

			if err := stream.Send(chunkReq); err != nil {
				return "", "", fmt.Errorf("не удалось отправить фрагмент: %w", err)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return "", "", fmt.Errorf("не удалось прочитать фрагмент: %w", err)
		}
	}

	// Закрываем отправку и получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return "", "", fmt.Errorf("не удалось получить ответ: %w", err)
	}

	if resp.Status != "OK" && resp.Status != "success" {
		// Server returns "OK" or "ALREADY_EXISTS"
		// Check for success condition?
		// Just pass it back?
	}

	return resp.Id, resp.Status, nil
}

// GetList получает список секретов
func (c *UserClient) GetList(ctx context.Context) ([]domain.SecretMeta, error) {
	resp, err := c.client.ListSecrets(ctx, &pb.ListSecretsRequest{})
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список секретов: %w", err)
	}

	if resp.Status != "success" && resp.Status != "" {
		return nil, fmt.Errorf("сервер вернул ошибку: %s", resp.Status)
	}

	var secrets []domain.SecretMeta
	for _, meta := range resp.Secrets {
		secrets = append(secrets, domain.SecretMeta{
			ID:    meta.Id,
			Type:  domain.SecretType(meta.Type),
			Extra: meta.Extra,
		})
	}

	return secrets, nil
}

// GetSecret получает секрет по ID
func (c *UserClient) GetSecret(ctx context.Context, id string) (*domain.Secret, error) {
	stream, err := c.client.GetSecret(ctx, &pb.GetSecretRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("не удалось запустить поток: %w", err)
	}

	// We need to read the first message to get metadata
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("не удалось найти метаданные: %w", err)
	}

	secret := &domain.Secret{
		Meta: domain.SecretMeta{
			ID: id,
		},
	}

	if resp.Meta != nil {
		secret.Meta.Type = domain.SecretType(resp.Meta.Type)
		secret.Meta.Extra = resp.Meta.Extra
	}

	// The data reader will be a wrapper around the stream
	secret.Data = &streamReader{stream: stream, firstChunk: resp.ChunkData}
	return secret, nil
}

type streamReader struct {
	stream     pb.SecretService_GetSecretClient
	firstChunk []byte
	buf        []byte
	err        error
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	if len(r.firstChunk) > 0 {
		n = copy(p, r.firstChunk)
		r.firstChunk = r.firstChunk[n:]
		return n, nil
	}

	if len(r.buf) > 0 {
		n = copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	if r.err != nil {
		return 0, r.err
	}

	resp, err := r.stream.Recv()
	if err != nil {
		r.err = err
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, err
	}

	n = copy(p, resp.ChunkData)
	if n < len(resp.ChunkData) {
		r.buf = resp.ChunkData[n:]
	}
	return n, nil
}

func (r *streamReader) Close() error {
	// gRPC receive stream doesn't strictly need closing from client side
	// other than context cancel, but we satisfy the interface.
	return nil
}

// DeleteSecret удаляет секрет
func (c *UserClient) DeleteSecret(ctx context.Context, id string) (string, error) {
	resp, err := c.client.DeleteSecret(ctx, &pb.DeleteSecretRequest{Id: id})
	if err != nil {
		return "", fmt.Errorf("не удалось удалить секрет: %w", err)
	}

	if resp.Status != "success" && resp.Status != "" {
		return "", fmt.Errorf("сервер вернул ошибку: %s", resp.Status)
	}

	return resp.Status, nil
}

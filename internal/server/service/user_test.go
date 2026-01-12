package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	proto "secretKeeper/internal/proto"
	"secretKeeper/internal/server/domain"
	"secretKeeper/internal/server/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// -- Mocks for Streams --

type mockCreateSecretServer struct {
	grpc.ServerStream
	ctx       context.Context
	recvData  []*proto.CreateSecretRequest // Data to be returned by Recv()
	recvIndex int
	sentId    string
	sentCode  string
}

func (m *mockCreateSecretServer) Context() context.Context {
	return m.ctx
}

func (m *mockCreateSecretServer) Recv() (*proto.CreateSecretRequest, error) {
	if m.recvIndex >= len(m.recvData) {
		return nil, io.EOF
	}
	data := m.recvData[m.recvIndex]
	m.recvIndex++
	return data, nil
}

func (m *mockCreateSecretServer) SendAndClose(resp *proto.CreateSecretResponse) error {
	m.sentId = resp.Id
	m.sentCode = resp.Status
	return nil

}

type mockGetSecretServer struct {
	grpc.ServerStream
	ctx      context.Context
	sentData []*proto.GetSecretResponse
}

func (m *mockGetSecretServer) Context() context.Context {
	return m.ctx
}

func (m *mockGetSecretServer) Send(resp *proto.GetSecretResponse) error {
	m.sentData = append(m.sentData, resp)
	return nil
}

// -- Tests for CreateSecret --

func TestUserService_CreateSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUser(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)
	logger := zap.NewNop().Sugar()

	service := NewUserService(mockRepo, mockFile, logger)
	userID := "user123"
	idempotencyKey := "key1"

	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		ctx = context.WithValue(ctx, "idempotency-key", idempotencyKey)

		// Mock check existing
		mockRepo.EXPECT().CheckSecretExists(idempotencyKey, userID).Return(false, nil)

		// Prepare stream data
		meta := &proto.SecretMetadata{Type: "text"}
		reqMeta := &proto.CreateSecretRequest{Data: &proto.CreateSecretRequest_Meta{Meta: meta}}

		chunk := []byte("content")
		reqChunk := &proto.CreateSecretRequest{Data: &proto.CreateSecretRequest_ChunkData{ChunkData: chunk}}

		stream := &mockCreateSecretServer{
			ctx:      ctx,
			recvData: []*proto.CreateSecretRequest{reqMeta, reqChunk},
		}

		// Mock UploadFile
		mockFile.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
				// Read from reader to drain pipe
				buf := make([]byte, 1024)
				n, _ := reader.Read(buf)
				assert.Equal(t, "content", string(buf[:n]))
				return nil
			},
		)

		// Mock SaveSecret
		mockRepo.EXPECT().SaveSecret(gomock.Any(), userID, gomock.Any(), idempotencyKey, meta).Return(nil)

		secretID, code, err := service.CreateSecret(stream)
		require.NoError(t, err)
		assert.Equal(t, "OK", code)
		assert.NotEmpty(t, secretID)
	})

	t.Run("Already Exists", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		ctx = context.WithValue(ctx, "idempotency-key", idempotencyKey)

		mockRepo.EXPECT().CheckSecretExists(idempotencyKey, userID).Return(true, nil)

		stream := &mockCreateSecretServer{ctx: ctx}
		id, code, err := service.CreateSecret(stream)

		assert.NoError(t, err)
		assert.Equal(t, "ALREADY_EXISTS", code)
		assert.Equal(t, idempotencyKey, id)
	})
}

func TestUserService_GetSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUser(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)
	logger := zap.NewNop().Sugar()

	service := NewUserService(mockRepo, mockFile, logger)
	userID := "user123"
	secretID := "file1"

	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		meta := &proto.SecretMetadata{Id: secretID, Type: "text"}
		// Fix: FileMetaData vs SecretMetadata mismatch?
		// UserService code: GetSecretMeta returns *pb.SecretMetadata. OK.

		mockRepo.EXPECT().FileExists(ctx, userID, secretID).Return(true, nil)
		mockRepo.EXPECT().GetSecretMeta(ctx, userID, secretID).Return(meta, nil)

		// Mock GetSecret file
		content := "secret content"
		fileReader := io.NopCloser(strings.NewReader(content))
		mockFile.EXPECT().GetSecret(ctx, secretID).Return(fileReader, nil)

		stream := &mockGetSecretServer{ctx: ctx}

		err := service.GetSecret(&proto.GetSecretRequest{Id: secretID}, stream)
		assert.NoError(t, err)

		// Verify sent data
		require.Len(t, stream.sentData, 2)
		assert.Equal(t, meta, stream.sentData[0].Meta)
		assert.Equal(t, []byte(content), stream.sentData[1].ChunkData)
	})

	t.Run("Not Found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		mockRepo.EXPECT().FileExists(ctx, userID, secretID).Return(false, nil)

		stream := &mockGetSecretServer{ctx: ctx}
		err := service.GetSecret(&proto.GetSecretRequest{Id: secretID}, stream)
		assert.Error(t, err)
	})
}

func TestUserService_ListSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUser(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)
	logger := zap.NewNop().Sugar()

	service := NewUserService(mockRepo, mockFile, logger)

	userID := "user123"

	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		files := []domain.File{
			{ID: "f1", Meta: domain.FileMetaData{SecretType: "text"}},
			{ID: "f2", Meta: domain.FileMetaData{SecretType: "binary"}},
		}

		mockRepo.EXPECT().ListUserFiles(ctx, userID).Return(files, nil)

		res, err := service.ListSecrets(ctx, &proto.ListSecretsRequest{})
		require.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "f1", res[0].Id)
		assert.Equal(t, "text", res[0].Type)
	})

	t.Run("No Context User", func(t *testing.T) {
		ctx := context.Background()
		_, err := service.ListSecrets(ctx, &proto.ListSecretsRequest{})
		assert.Error(t, err)
	})

	t.Run("Repo Error", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		mockRepo.EXPECT().ListUserFiles(ctx, userID).Return(nil, errors.New("db error"))

		_, err := service.ListSecrets(ctx, &proto.ListSecretsRequest{})
		assert.Error(t, err)
	})
}

func TestUserService_DeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUser(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)
	logger := zap.NewNop().Sugar()

	service := NewUserService(mockRepo, mockFile, logger)
	userID := "user123"
	secretID := "file1"

	t.Run("Success", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)

		mockRepo.EXPECT().FileExists(ctx, userID, secretID).Return(true, nil)
		mockFile.EXPECT().DeleteSecret(ctx, secretID).Return(nil)
		mockRepo.EXPECT().DeleteSecret(ctx, userID, secretID).Return(nil)

		err := service.DeleteSecret(ctx, &proto.DeleteSecretRequest{Id: secretID})
		assert.NoError(t, err)
	})

	t.Run("Not Found", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "user_id", userID)
		mockRepo.EXPECT().FileExists(ctx, userID, secretID).Return(false, nil)

		err := service.DeleteSecret(ctx, &proto.DeleteSecretRequest{Id: secretID})
		assert.Error(t, err)
	})
}

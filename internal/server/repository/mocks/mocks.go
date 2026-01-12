package mocks

import (
	context "context"
	io "io"
	reflect "reflect"
	api "secretKeeper/internal/proto"
	domain "secretKeeper/internal/server/domain"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockAuthorization is a mock of Authorization interface.
type MockAuthorization struct {
	ctrl     *gomock.Controller
	recorder *MockAuthorizationMockRecorder
	isgomock struct{}
}

// MockAuthorizationMockRecorder is the mock recorder for MockAuthorization.
type MockAuthorizationMockRecorder struct {
	mock *MockAuthorization
}

// NewMockAuthorization creates a new mock instance.
func NewMockAuthorization(ctrl *gomock.Controller) *MockAuthorization {
	mock := &MockAuthorization{ctrl: ctrl}
	mock.recorder = &MockAuthorizationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthorization) EXPECT() *MockAuthorizationMockRecorder {
	return m.recorder
}

// CreateSession mocks base method.
func (m *MockAuthorization) CreateSession(ctx context.Context, id, refreshTokenHash string, expiresIn time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, id, refreshTokenHash, expiresIn)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockAuthorizationMockRecorder) CreateSession(ctx, id, refreshTokenHash, expiresIn any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockAuthorization)(nil).CreateSession), ctx, id, refreshTokenHash, expiresIn)
}

// CreateUser mocks base method.
func (m *MockAuthorization) CreateUser(ctx context.Context, username, hashedPassword string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, username, hashedPassword)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockAuthorizationMockRecorder) CreateUser(ctx, username, hashedPassword any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockAuthorization)(nil).CreateUser), ctx, username, hashedPassword)
}

// LoginUser mocks base method.
func (m *MockAuthorization) LoginUser(ctx context.Context, username string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoginUser", ctx, username)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// LoginUser indicates an expected call of LoginUser.
func (mr *MockAuthorizationMockRecorder) LoginUser(ctx, username any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoginUser", reflect.TypeOf((*MockAuthorization)(nil).LoginUser), ctx, username)
}

// MockUser is a mock of User interface.
type MockUser struct {
	ctrl     *gomock.Controller
	recorder *MockUserMockRecorder
	isgomock struct{}
}

// MockUserMockRecorder is the mock recorder for MockUser.
type MockUserMockRecorder struct {
	mock *MockUser
}

// NewMockUser creates a new mock instance.
func NewMockUser(ctrl *gomock.Controller) *MockUser {
	mock := &MockUser{ctrl: ctrl}
	mock.recorder = &MockUserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUser) EXPECT() *MockUserMockRecorder {
	return m.recorder
}

// CheckSecretExists mocks base method.
func (m *MockUser) CheckSecretExists(transactionID, userID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckSecretExists", transactionID, userID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckSecretExists indicates an expected call of CheckSecretExists.
func (mr *MockUserMockRecorder) CheckSecretExists(transactionID, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckSecretExists", reflect.TypeOf((*MockUser)(nil).CheckSecretExists), transactionID, userID)
}

// DeleteSecret mocks base method.
func (m *MockUser) DeleteSecret(ctx context.Context, userID, secretID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", ctx, userID, secretID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSecret indicates an expected call of DeleteSecret.
func (mr *MockUserMockRecorder) DeleteSecret(ctx, userID, secretID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MockUser)(nil).DeleteSecret), ctx, userID, secretID)
}

// FileExists mocks base method.
func (m *MockUser) FileExists(ctx context.Context, userID, secretID string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FileExists", ctx, userID, secretID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FileExists indicates an expected call of FileExists.
func (mr *MockUserMockRecorder) FileExists(ctx, userID, secretID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FileExists", reflect.TypeOf((*MockUser)(nil).FileExists), ctx, userID, secretID)
}

// GetSecretMeta mocks base method.
func (m *MockUser) GetSecretMeta(ctx context.Context, userID, secretID string) (*api.SecretMetadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecretMeta", ctx, userID, secretID)
	ret0, _ := ret[0].(*api.SecretMetadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecretMeta indicates an expected call of GetSecretMeta.
func (mr *MockUserMockRecorder) GetSecretMeta(ctx, userID, secretID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecretMeta", reflect.TypeOf((*MockUser)(nil).GetSecretMeta), ctx, userID, secretID)
}

// ListUserFiles mocks base method.
func (m *MockUser) ListUserFiles(ctx context.Context, userID string) ([]domain.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUserFiles", ctx, userID)
	ret0, _ := ret[0].([]domain.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUserFiles indicates an expected call of ListUserFiles.
func (mr *MockUserMockRecorder) ListUserFiles(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUserFiles", reflect.TypeOf((*MockUser)(nil).ListUserFiles), ctx, userID)
}

// SaveSecret mocks base method.
func (m *MockUser) SaveSecret(ctx context.Context, userID, secretID, idempotencyKey string, meta *api.SecretMetadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveSecret", ctx, userID, secretID, idempotencyKey, meta)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveSecret indicates an expected call of SaveSecret.
func (mr *MockUserMockRecorder) SaveSecret(ctx, userID, secretID, idempotencyKey, meta any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveSecret", reflect.TypeOf((*MockUser)(nil).SaveSecret), ctx, userID, secretID, idempotencyKey, meta)
}

// MockFileStorage is a mock of FileStorage interface.
type MockFileStorage struct {
	ctrl     *gomock.Controller
	recorder *MockFileStorageMockRecorder
	isgomock struct{}
}

// MockFileStorageMockRecorder is the mock recorder for MockFileStorage.
type MockFileStorageMockRecorder struct {
	mock *MockFileStorage
}

// NewMockFileStorage creates a new mock instance.
func NewMockFileStorage(ctrl *gomock.Controller) *MockFileStorage {
	mock := &MockFileStorage{ctrl: ctrl}
	mock.recorder = &MockFileStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileStorage) EXPECT() *MockFileStorageMockRecorder {
	return m.recorder
}

// DeleteSecret mocks base method.
func (m *MockFileStorage) DeleteSecret(ctx context.Context, secretID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", ctx, secretID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSecret indicates an expected call of DeleteSecret.
func (mr *MockFileStorageMockRecorder) DeleteSecret(ctx, secretID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MockFileStorage)(nil).DeleteSecret), ctx, secretID)
}

// GetSecret mocks base method.
func (m *MockFileStorage) GetSecret(ctx context.Context, secretID string) (io.Reader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", ctx, secretID)
	ret0, _ := ret[0].(io.Reader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockFileStorageMockRecorder) GetSecret(ctx, secretID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockFileStorage)(nil).GetSecret), ctx, secretID)
}

// UploadFile mocks base method.
func (m *MockFileStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadFile", ctx, objectName, reader, objectSize, contentType)
	ret0, _ := ret[0].(error)
	return ret0
}

// UploadFile indicates an expected call of UploadFile.
func (mr *MockFileStorageMockRecorder) UploadFile(ctx, objectName, reader, objectSize, contentType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadFile", reflect.TypeOf((*MockFileStorage)(nil).UploadFile), ctx, objectName, reader, objectSize, contentType)
}

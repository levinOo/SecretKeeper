-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    minio_object_id TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_meta JSONB NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_user_files_user_id_idempotency_key
    ON user_files(user_id, idempotency_key);

CREATE INDEX IF NOT EXISTS idx_user_files_user_id ON user_files(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_files_user_id;
DROP INDEX IF EXISTS ux_user_files_user_id_idempotency_key;
DROP TABLE IF EXISTS user_files;
-- +goose StatementEnd

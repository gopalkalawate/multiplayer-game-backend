-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS matches (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS matches;
-- +goose StatementEnd

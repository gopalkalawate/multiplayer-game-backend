-- +goose Up
CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    mmr INTEGER NOT NULL,
    ping INTEGER NOT NULL,
    region TEXT NOT NULL,
    tier TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementBegin
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS players;
-- +goose StatementBegin
-- +goose StatementEnd

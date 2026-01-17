-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS matches_players (
    match_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    PRIMARY KEY (match_id, player_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS matches_players;
-- +goose StatementEnd

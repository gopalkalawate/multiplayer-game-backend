-- +goose Up
-- +goose StatementBegin
ALTER TABLE players ADD COLUMN status TEXT DEFAULT 'waiting';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE players DROP COLUMN status;
-- +goose StatementEnd

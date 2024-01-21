-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE user
    ADD COLUMN salt BLOB NOT NULL default '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE user
    DROP COLUMN salt;
-- +goose StatementEnd

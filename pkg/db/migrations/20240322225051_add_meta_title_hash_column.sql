-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE reference
    ADD COLUMN meta_title_hash TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE reference
    DROP COLUMN meta_title_hash;
-- +goose StatementEnd

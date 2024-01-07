-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE gallery
    ADD COLUMN deleted NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE gallery
    DROP COLUMN deleted;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE gallery
    ADD COLUMN page_thumbnails int;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE gallery
    DROP COLUMN page_thumbnails;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE gallery RENAME COLUMN title_short TO title_translated;
ALTER TABLE gallery DROP COLUMN artists;
ALTER TABLE gallery DROP COLUMN circle;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE gallery RENAME COLUMN title_translated TO title_short;
ALTER TABLE gallery ADD COLUMN artists text;
ALTER TABLE gallery ADD COLUMN circle text;
-- +goose StatementEnd

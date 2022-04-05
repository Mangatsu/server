-- +goose Up
-- +goose StatementBegin
ALTER TABLE gallery RENAME COLUMN title_short TO project_id;
ALTER TABLE gallery DROP COLUMN artists;
ALTER TABLE gallery DROP COLUMN circle;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE gallery RENAME COLUMN project_id TO title_short;
ALTER TABLE gallery ADD COLUMN artists text;
ALTER TABLE gallery ADD COLUMN circle text;
-- +goose StatementEnd

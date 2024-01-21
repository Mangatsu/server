-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE gallery2
(
    uuid             text UNIQUE NOT NULL,
    library_id       integer     NOT NULL,
    archive_path     text UNIQUE NOT NULL,
    title            text        NOT NULL,
    title_native     text,
    title_translated text,
    category         text,
    series           text,
    released         text,
    language         text,
    translated       boolean,
    nsfw             boolean     NOT NULL DEFAULT false,
    hidden           boolean     NOT NULL DEFAULT false,
    image_count      int,
    archive_size     int,
    archive_hash     text,
    thumbnail        text,
    created_at       datetime    NOT NULL,
    updated_at       datetime    NOT NULL,
    PRIMARY KEY (uuid),
    FOREIGN KEY (library_id)
        REFERENCES library (id)
);
INSERT INTO gallery2
SELECT uuid,
       library_id,
       archive_path,
       title,
       title_native,
       title_short AS title_translated,
       category,
       series,
       released, language, translated, nsfw, hidden, image_count, archive_size, archive_hash, thumbnail, created_at, updated_at
FROM gallery;
DROP TABLE gallery;
ALTER TABLE gallery2 RENAME TO gallery;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
CREATE TABLE gallery2
(
    uuid         text UNIQUE NOT NULL,
    library_id   integer     NOT NULL,
    archive_path text UNIQUE NOT NULL,
    title        text        NOT NULL,
    title_native text,
    title_short  text,
    released     text,
    circle       text,
    artists      text,
    series       text,
    category     text,
    language     text,
    translated   boolean,
    image_count  int,
    archive_size int,
    archive_hash text,
    thumbnail    text,
    nsfw         boolean     NOT NULL DEFAULT false,
    hidden       boolean     NOT NULL DEFAULT false,
    created_at   datetime    NOT NULL,
    updated_at   datetime    NOT NULL,
    PRIMARY KEY (uuid),
    FOREIGN KEY (library_id)
        REFERENCES library (id)
);
INSERT INTO gallery2
SELECT uuid,
       library_id,
       archive_path,
       title,
       title_native,
       title_translated AS title_short,
       released,
       null             AS circle,
       null             AS artists,
       series,
       category, language, translated, image_count, archive_size, archive_hash, thumbnail, nsfw, hidden, created_at, updated_at
FROM gallery;
DROP TABLE gallery;
ALTER TABLE gallery2 RENAME TO gallery;
-- +goose StatementEnd

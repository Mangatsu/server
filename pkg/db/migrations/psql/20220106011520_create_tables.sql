-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS library (
    id integer UNIQUE NOT NULL,
    path text UNIQUE NOT NULL,
    layout text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS gallery (
    uuid UUID UNIQUE NOT NULL,
    library_id integer NOT NULL,
    archive_path text UNIQUE NOT NULL,
    title text NOT NULL,
    title_native text,
    title_short text,
    released text,
    circle text,
    artists text,
    series text,
    category text,
    language text,
    translated boolean,
    image_count int,
    archive_size int,
    archive_hash text,
    thumbnail text,
    nsfw boolean NOT NULL DEFAULT false,
    hidden boolean NOT NULL DEFAULT false,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY(uuid),
    FOREIGN KEY(library_id)
        REFERENCES library(id)
);

CREATE TABLE IF NOT EXISTS tag (
    id SERIAL PRIMARY KEY,
    namespace text NOT NULL,
    name text NOT NULL,
    CONSTRAINT unique_tag UNIQUE (namespace, name)
);

CREATE TABLE IF NOT EXISTS gallery_tag (
    gallery_uuid UUID NOT NULL,
    tag_id integer NOT NULL,
    CONSTRAINT unique_gallery_tag UNIQUE (gallery_uuid, tag_id),
    CONSTRAINT gallery
       FOREIGN KEY(gallery_uuid)
           REFERENCES gallery(uuid)
           ON DELETE CASCADE,
    CONSTRAINT tag
       FOREIGN KEY(tag_id)
           REFERENCES tag(id)
           ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reference (
    gallery_uuid UUID UNIQUE NOT NULL,
    meta_internal boolean NOT NULL DEFAULT false,
    meta_path text,
    meta_match integer,
    urls text,
    exh_gid int,
    exh_token text,
    anilist_id int,
    PRIMARY KEY(gallery_uuid),
    CONSTRAINT gallery
        FOREIGN KEY(gallery_uuid)
            REFERENCES gallery(uuid)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "user" (
    uuid UUID UNIQUE NOT NULL,
    username text UNIQUE NOT NULL,
    password text NOT NULL,
    role integer NOT NULL DEFAULT 10,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY(uuid)
);

CREATE TABLE IF NOT EXISTS session (
    id text NOT NULL,
    user_uuid UUID NOT NULL,
    name text,
    expires_at timestamp,
    PRIMARY KEY(id, user_uuid),
    CONSTRAINT "user"
        FOREIGN KEY(user_uuid)
            REFERENCES "user" (uuid)
            ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS gallery_pref (
    user_uuid UUID NOT NULL,
    gallery_uuid UUID NOT NULL,
    progress integer NOT NULL DEFAULT 0,
    favorite_group text,
    updated_at timestamp NOT NULL,
    PRIMARY KEY(user_uuid, gallery_uuid),
    FOREIGN KEY(gallery_uuid)
        REFERENCES gallery(uuid),
    CONSTRAINT "user"
        FOREIGN KEY(user_uuid)
            REFERENCES "user" (uuid)
            ON DELETE CASCADE
);

CREATE INDEX idx_title ON gallery (title);
CREATE INDEX idx_title_native ON gallery (title_native);
CREATE INDEX idx_series ON gallery (series);
CREATE INDEX idx_category ON gallery (category);
CREATE INDEX idx_archive_path ON gallery (archive_path);
CREATE INDEX idx_updated_at ON gallery (updated_at);
CREATE INDEX idx_tag ON tag (namespace, name);
CREATE INDEX idx_gallery_pref ON gallery_pref (user_uuid, gallery_uuid);
CREATE INDEX idx_favorite ON gallery_pref (favorite_group);
CREATE INDEX idx_session ON session (id, user_uuid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reference;
DROP TABLE IF EXISTS tag;
DROP TABLE IF EXISTS gallery_pref;
DROP TABLE IF EXISTS gallery;
DROP TABLE IF EXISTS library;
DROP TABLE IF EXISTS session;
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd

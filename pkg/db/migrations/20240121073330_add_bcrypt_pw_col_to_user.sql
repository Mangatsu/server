-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE user2
(
    uuid       text UNIQUE NOT NULL,
    username   text UNIQUE NOT NULL,
    password   blob        NOT NULL,
    salt       blob        NOT NULL,
    role       integer     NOT NULL DEFAULT 10,
    bcrypt_pw  text,
    created_at datetime    NOT NULL,
    updated_at datetime    NOT NULL
);
INSERT INTO user2
SELECT uuid,
       username,
       ''       AS password,
       ''       AS salt,
       role,
       password AS bcrypt_pw,
       created_at,
       updated_at
FROM user;

DROP TABLE user;

ALTER TABLE user2
    RENAME TO user;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
CREATE TABLE user2
(
    uuid       text UNIQUE NOT NULL,
    username   text UNIQUE NOT NULL,
    password   blob        NOT NULL,
    role       integer     NOT NULL DEFAULT 10,
    created_at datetime    NOT NULL,
    updated_at datetime    NOT NULL
);

INSERT INTO user2
SELECT uuid,
       username,
       bcrypt_pw AS password,
       role,
       created_at,
       updated_at
FROM user;

DROP TABLE user;

ALTER TABLE user2
    RENAME TO user;
-- +goose StatementEnd

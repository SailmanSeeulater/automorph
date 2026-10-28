-- +goose Up

-- Times are stored as Unix seconds (INTEGER). unixepoch() returns the current one.

CREATE TABLE users (
    id            INTEGER PRIMARY KEY,
    email         TEXT    NOT NULL UNIQUE COLLATE NOCASE,
    password_hash TEXT    NOT NULL,
    display_name  TEXT    NOT NULL DEFAULT '',
    role          TEXT    NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    created_at    INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at    INTEGER NOT NULL DEFAULT (unixepoch())
);

-- id is the SHA-256 of the cookie token, so a leaked DB can't be used to log in.
CREATE TABLE sessions (
    id         TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER NOT NULL
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

-- One row per uploaded image. Resized variants live on disk next to the original.
CREATE TABLE media (
    id          INTEGER PRIMARY KEY,
    uploader_id INTEGER REFERENCES users (id) ON DELETE SET NULL,
    storage_key TEXT    NOT NULL UNIQUE,
    mime        TEXT    NOT NULL,
    width       INTEGER NOT NULL,
    height      INTEGER NOT NULL,
    bytes       INTEGER NOT NULL,
    alt         TEXT    NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE posts (
    id             INTEGER PRIMARY KEY,
    author_id      INTEGER NOT NULL REFERENCES users (id),
    slug           TEXT    NOT NULL UNIQUE,
    title          TEXT    NOT NULL,
    excerpt        TEXT    NOT NULL DEFAULT '',
    body_md        TEXT    NOT NULL DEFAULT '',
    status         TEXT    NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    cover_media_id INTEGER REFERENCES media (id) ON DELETE SET NULL,
    published_at   INTEGER,
    created_at     INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at     INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX posts_status_published_at_idx ON posts (status, published_at DESC);

-- +goose Down
DROP TABLE posts;
DROP TABLE media;
DROP TABLE sessions;
DROP TABLE users;
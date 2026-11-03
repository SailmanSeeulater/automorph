-- name: ListPublishedPosts :many
SELECT * FROM posts
WHERE status = 'published'
ORDER BY published_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetPublishedPostBySlug :one
SELECT * FROM posts
WHERE slug = ? AND status = 'published';

-- name: ListAllPosts :many
SELECT * FROM posts
ORDER BY updated_at DESC, id DESC;

-- name: GetPost :one
SELECT * FROM posts
WHERE id = ?;

-- name: SlugTaken :one
SELECT EXISTS (SELECT 1 FROM posts WHERE slug = ?);

-- name: CreatePost :one
-- published_at is set the first time a post is published and kept after that.
INSERT INTO posts (author_id, slug, title, excerpt, body_md, cover_media_id, status, published_at)
VALUES (
    sqlc.arg(author_id), sqlc.arg(slug), sqlc.arg(title), sqlc.arg(excerpt), sqlc.arg(body_md),
    sqlc.narg(cover_media_id), sqlc.arg(status),
    CASE WHEN sqlc.arg(status) = 'published' THEN unixepoch() END
)
RETURNING *;

-- name: UpdatePost :one
UPDATE posts SET
    slug           = sqlc.arg(slug),
    title          = sqlc.arg(title),
    excerpt        = sqlc.arg(excerpt),
    body_md        = sqlc.arg(body_md),
    cover_media_id = sqlc.narg(cover_media_id),
    status         = sqlc.arg(status),
    published_at   = CASE
        WHEN sqlc.arg(status) = 'published' AND published_at IS NULL THEN unixepoch()
        ELSE published_at
    END,
    updated_at     = unixepoch()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePost :execrows
DELETE FROM posts
WHERE id = ?; 
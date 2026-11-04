-- name: CreateImage :one
INSERT INTO media (uploader_id, storage_key, mime, width, height, bytes, alt)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetImage :one
SELECT * FROM media
WHERE id = ?;

-- name: ListImages :many
SELECT * FROM media
ORDER BY id DESC;

-- name: DeleteImage :one
-- Returns the row so the caller knows which files to remove.
DELETE FROM media
WHERE id = ?
RETURNING *;
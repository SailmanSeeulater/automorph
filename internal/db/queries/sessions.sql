-- name: CreateSession :exec
INSERT INTO sessions (id, user_id, expires_at)
VALUES (?, ?, ?);

-- name: GetSessionUser :one
SELECT sqlc.embed(users)
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.id = ? AND sessions.expires_at > unixepoch();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= unixepoch();
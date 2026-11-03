-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?
LIMIT 1;

-- name: UpsertAdmin :one
-- Creates the admin, or resets the password (and promotes to admin) if the email exists.
INSERT INTO users (email, password_hash, display_name, role)
VALUES (?, ?, ?, 'admin')
ON CONFLICT (email) DO UPDATE SET
    password_hash = excluded.password_hash,
    display_name  = CASE WHEN excluded.display_name != '' THEN excluded.display_name ELSE users.display_name END,
    role          = 'admin',
    updated_at    = unixepoch()
RETURNING *;
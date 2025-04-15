-- name: FindPlayerByID :one
SELECT * FROM users WHERE id = $1;

-- name: FindPlayerByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: FindPlayerByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: Create :one
INSERT INTO users (id, username, email, passHash, created_at, updated_at)
VALUES (uuid_generate_v4(), $1, $2, $3, now(), now())
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users SET passHash = $1 WHERE id = $2;

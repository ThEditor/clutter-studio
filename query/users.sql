-- name: FindUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: FindUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, username, email, passHash, created_at, updated_at)
VALUES (uuid_generate_v4(), $1, $2, $3, now(), now())
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET passHash = $1 WHERE id = $2;

-- name: UpdateEmailVerificationStatus :exec
UPDATE users SET email_verified = $1 WHERE id = $2;

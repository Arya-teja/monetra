-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetTokenVersion :one
SELECT token_version FROM users
WHERE id = $1;

-- name: IncrementTokenVersion :execrows
UPDATE users
SET token_version = token_version + 1, updated_at = now()
WHERE id = $1;

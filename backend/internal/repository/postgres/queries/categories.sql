-- name: CreateCategory :exec
INSERT INTO categories (user_id, name, type, is_default)
VALUES ($1, $2, $3, $4);

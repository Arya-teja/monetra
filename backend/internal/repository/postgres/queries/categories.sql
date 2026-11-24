-- name: CreateCategory :one
INSERT INTO categories (user_id, name, type, is_default)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListCategories :many
SELECT * FROM categories
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.narg(type)::cashflow_type IS NULL OR type = sqlc.narg(type))
ORDER BY type, name;

-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND user_id = $2;

-- name: RenameCategory :one
UPDATE categories
SET name = $3, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteCategory :execrows
DELETE FROM categories
WHERE id = $1 AND user_id = $2;

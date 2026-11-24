-- name: ListAccounts :many
SELECT * FROM accounts
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.arg(include_archived)::boolean OR NOT is_archived)
ORDER BY created_at;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 AND user_id = $2;

-- name: CreateAccount :one
INSERT INTO accounts (user_id, name, type, institution_name, account_number, balance)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $3,
    type = $4,
    institution_name = $5,
    account_number = $6,
    is_archived = $7,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteAccount :execrows
DELETE FROM accounts
WHERE id = $1 AND user_id = $2;

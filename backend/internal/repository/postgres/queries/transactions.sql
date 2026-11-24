-- name: ListTransactions :many
SELECT t.*, a.name AS account_name, c.name AS category_name
FROM transactions t
JOIN accounts a ON a.id = t.account_id
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(from_date)::date IS NULL OR t.occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::date IS NULL OR t.occurred_at < sqlc.narg(to_date))
  AND (sqlc.narg(account_id)::uuid IS NULL OR t.account_id = sqlc.narg(account_id))
  AND (sqlc.narg(category_id)::uuid IS NULL OR t.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(type)::cashflow_type IS NULL OR t.type = sqlc.narg(type))
ORDER BY t.occurred_at DESC, t.created_at DESC
LIMIT sqlc.arg(row_limit) OFFSET sqlc.arg(row_offset);

-- name: CountTransactions :one
SELECT count(*) FROM transactions t
WHERE t.user_id = sqlc.arg(user_id)
  AND (sqlc.narg(from_date)::date IS NULL OR t.occurred_at >= sqlc.narg(from_date))
  AND (sqlc.narg(to_date)::date IS NULL OR t.occurred_at < sqlc.narg(to_date))
  AND (sqlc.narg(account_id)::uuid IS NULL OR t.account_id = sqlc.narg(account_id))
  AND (sqlc.narg(category_id)::uuid IS NULL OR t.category_id = sqlc.narg(category_id))
  AND (sqlc.narg(type)::cashflow_type IS NULL OR t.type = sqlc.narg(type));

-- name: GetTransaction :one
SELECT t.*, a.name AS account_name, c.name AS category_name
FROM transactions t
JOIN accounts a ON a.id = t.account_id
JOIN categories c ON c.id = t.category_id
WHERE t.id = $1 AND t.user_id = $2;

-- name: CreateTransaction :one
INSERT INTO transactions (user_id, account_id, category_id, type, amount, occurred_at, note)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- Optimistic lock: only succeeds while updated_at still equals the value the caller read.
-- name: UpdateTransaction :execrows
UPDATE transactions
SET account_id = sqlc.arg(account_id),
    category_id = sqlc.arg(category_id),
    type = sqlc.arg(type),
    amount = sqlc.arg(amount),
    occurred_at = sqlc.arg(occurred_at),
    note = sqlc.narg(note),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND updated_at = sqlc.arg(expected_updated_at);

-- name: DeleteTransaction :execrows
DELETE FROM transactions
WHERE id = sqlc.arg(id)
  AND user_id = sqlc.arg(user_id)
  AND updated_at = sqlc.arg(expected_updated_at);

-- name: AdjustAccountBalance :execrows
UPDATE accounts
SET balance = balance + sqlc.arg(delta), updated_at = now()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id);

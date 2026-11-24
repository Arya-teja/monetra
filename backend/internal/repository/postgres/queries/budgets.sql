-- spent = expense transactions of the budget's category within the budget's calendar month.

-- name: ListBudgets :many
SELECT b.id, b.user_id, b.category_id, c.name AS category_name, b.month, b.amount,
       COALESCE(SUM(t.amount), 0)::bigint AS spent
FROM budgets b
JOIN categories c ON c.id = b.category_id
LEFT JOIN transactions t
       ON t.user_id = b.user_id
      AND t.category_id = b.category_id
      AND t.type = 'expense'
      AND t.occurred_at >= b.month
      AND t.occurred_at < (b.month + INTERVAL '1 month')::date
WHERE b.user_id = $1 AND b.month = $2
GROUP BY b.id, c.name
ORDER BY c.name;

-- name: GetBudget :one
SELECT b.id, b.user_id, b.category_id, c.name AS category_name, b.month, b.amount,
       COALESCE(SUM(t.amount), 0)::bigint AS spent
FROM budgets b
JOIN categories c ON c.id = b.category_id
LEFT JOIN transactions t
       ON t.user_id = b.user_id
      AND t.category_id = b.category_id
      AND t.type = 'expense'
      AND t.occurred_at >= b.month
      AND t.occurred_at < (b.month + INTERVAL '1 month')::date
WHERE b.id = $1 AND b.user_id = $2
GROUP BY b.id, c.name;

-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, month, amount)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: UpdateBudgetAmount :execrows
UPDATE budgets
SET amount = $3, updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: DeleteBudget :execrows
DELETE FROM budgets
WHERE id = $1 AND user_id = $2;

-- Categories that already have a budget in the target month are skipped, not overwritten.
-- name: CopyBudgets :execrows
INSERT INTO budgets (user_id, category_id, month, amount)
SELECT src.user_id, src.category_id, sqlc.arg(to_month)::date, src.amount
FROM budgets src
WHERE src.user_id = sqlc.arg(user_id) AND src.month = sqlc.arg(from_month)
ON CONFLICT (user_id, month, category_id) DO NOTHING;

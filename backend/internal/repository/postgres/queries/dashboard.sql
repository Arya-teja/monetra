-- name: TotalActiveBalance :one
SELECT COALESCE(SUM(balance), 0)::bigint
FROM accounts
WHERE user_id = $1 AND NOT is_archived;

-- name: MonthlyTotals :many
-- Income and expense per calendar month within [from_date, to_date).
SELECT date_trunc('month', occurred_at)::date AS month,
       COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0)::bigint AS income,
       COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)::bigint AS expense
FROM transactions
WHERE user_id = sqlc.arg(user_id)
  AND occurred_at >= sqlc.arg(from_date)
  AND occurred_at < sqlc.arg(to_date)
GROUP BY 1
ORDER BY 1;

-- name: TopExpenseCategories :many
SELECT c.id, c.name, SUM(t.amount)::bigint AS total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.user_id = sqlc.arg(user_id)
  AND t.type = 'expense'
  AND t.occurred_at >= sqlc.arg(from_date)
  AND t.occurred_at < sqlc.arg(to_date)
GROUP BY c.id, c.name
ORDER BY total DESC, c.name
LIMIT sqlc.arg(row_limit);

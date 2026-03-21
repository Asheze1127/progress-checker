SELECT
    u.id,
    u.name,
    u.email,
    COUNT(o.id) AS order_count,
    COALESCE(SUM(o.total_amount), 0) AS total_spent
FROM
    users u
    LEFT JOIN orders o ON o.user_id = u.id
WHERE
    u.deleted_at IS NULL
GROUP BY
    u.id,
    u.name,
    u.email
HAVING
    COUNT(o.id) >= 1
ORDER BY
    total_spent DESC,
    u.id ASC
LIMIT
    10;
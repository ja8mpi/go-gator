-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (user_id, feed_id, created_at, updated_at)
    VALUES ($1, $2, NOW(), NOW())
    RETURNING *
)

SELECT
    iff.*,
    f.name AS feed_name,
    u.name AS user_name
FROM inserted_feed_follow AS iff
INNER JOIN feeds AS f ON iff.feed_id = f.id
INNER JOIN users AS u ON iff.user_id = u.id;

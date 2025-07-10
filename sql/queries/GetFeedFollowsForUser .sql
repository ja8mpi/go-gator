-- name: GetFeedFollowsForUser :many

SELECT DISTINCT feeds.name as feed_name, users.name as user_name
FROM feeds
INNER JOIN feed_follows ON feeds.id = feed_follows.feed_id
INNER JOIN users ON users.id = feed_follows.user_id
WHERE users.name = $1;

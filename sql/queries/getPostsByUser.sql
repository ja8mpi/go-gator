-- name: GetPostsForUser :many

SELECT posts.*
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
JOIN feed_follows ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC;
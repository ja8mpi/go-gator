-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE feeds.url = $1;

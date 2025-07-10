-- name: DeleteFeedFollow :exec

DELETE FROM feed_follows WHERE feed_id = $1 AND user_id = $2;
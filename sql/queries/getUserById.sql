-- name: GetUserByID :one
SELECT * FROM users WHERE users.id = $1;

-- name: GetUsers :many
SELECT * FROM user;

-- name: GetUser :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO user (email, hashed_password, avatar_url) VALUES (
  ?, ?, ?
);

-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?;

-- name: DeleteUserByEmail :exec
DELETE FROM user
WHERE email = ?;

-- name: GetClients :many
SELECT * FROM client;

-- name: GetClient :one
SELECT * FROM client
WHERE id = ?;

-- name: GetClientByName :one
SELECT * FROM client
WHERE name = ?;

-- name: GetUserAndClientCount :one
SELECT
	(SELECT COUNT(*) FROM user) as user_count,
	(SELECT COUNT(*) FROM client) as client_count;

-- name: CreateClient :exec
INSERT INTO client (
	id,
	secret_hash,
	name,
	callback_urls,
	picture_url,
	token_expiration
) VALUES (
	?, ?, ?, ?, ?, ?
);

-- name: GetClientByNameForUnmatchingID :one
SELECT * FROM client
WHERE name = ? AND id != ?;

-- name: UpdateClient :exec
UPDATE client
SET name = ?,
	  callback_urls = ?,
		picture_url = ?,
		token_expiration = ?
WHERE id = ?;

-- name: DeleteClient :exec
DELETE FROM client
WHERE id = ?;
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

-- name: UpdateUserPasswordHash :exec
UPDATE user
SET hashed_password = ?
WHERE id = ?;

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

-- name: CreateSession :execresult
INSERT INTO session (id, user_id, client_id, expires_at, os, browser) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: GetSession :one
SELECT * FROM session
WHERE id = ? LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM session
WHERE id = ? OR expires_at <= NOW();

-- name: GetSessionForUserID :many
SELECT
    session.id,
    session.created_at,
    session.expires_at,
    session.os,
    session.browser,
    client.name as client_name
FROM
    session
LEFT JOIN client
ON session.client_id = client.id
WHERE session.user_id = ?;

-- name: GetSessionWithUser :one
SELECT
    session.id,
    session.expires_at,
    user.id as user_id,
    user.email,
    user.avatar_url
FROM
    session
INNER JOIN user
ON session.user_id = user.id
WHERE session.id = ?;

-- name: GetAuthzHistory :one
SELECT * FROM authorization_history
WHERE user_id = ? AND client_id = ?;

-- name: CreateAuthzHistory :execresult
INSERT INTO authorization_history (user_id, client_id) VALUES (
    ?, ?
);
-- name: GetUsers :many
SELECT * FROM user;

-- name: GetUser :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: GetClients :many
SELECT * FROM client;

-- name: GetClient :one
SELECT * FROM client
WHERE id = ?;

-- name: GetClientByName :one
SELECT * FROM client
WHERE name = ?;

-- name: GetClientByNameForUnmatchingID :one
SELECT * FROM client
WHERE name = ? AND id != ?;

-- name: GetUserAndClientCount :one
SELECT
    (SELECT COUNT(*) FROM user) as user_count,
    (SELECT COUNT(*) FROM client) as client_count;

-- name: GetSession :one
SELECT * FROM session
WHERE id = ? LIMIT 1;

-- name: GetSessionWithUser :one
SELECT
    session.id,
    session.expires_at,
    user.id as user_id,
    user.email,
    user.avatar_url,
    user.is_admin
FROM
    session
INNER JOIN
    user
ON
    session.user_id = user.id
WHERE
    session.id = ?;

-- name: GetSessionWithClient :one
SELECT
    session.id,
    client.id as client_id,
    client.name as client_name,
    client.logout_callback_urls,
    client.backchannel_logout_url
FROM
    session
INNER JOIN
    client
ON
    session.client_id = client.id
WHERE
    session.id = ?;

-- name: GetSessionWithOptionalClient :one
SELECT
    session.id,
    client.id as client_id,
    client.name as client_name,
    client.logout_callback_urls,
    client.backchannel_logout_url
FROM
    session
LEFT JOIN
    client
ON
    session.client_id = client.id
WHERE
    session.id = ?;

-- name: GetSessionWithClientForUserID :many
SELECT
    session.id,
    session.created_at,
    session.expires_at,
    session.os,
    session.browser,
    client.name as client_name
FROM
    session
LEFT JOIN
    client
ON
    session.client_id = client.id
WHERE
    session.user_id = ?;

-- name: GetAuthzHistory :one
SELECT * FROM authorization_history
WHERE user_id = ? AND client_id = ?;

-- name: GetAuthzCode :one
SELECT * FROM authorization_code
WHERE id = ?;


-- name: CreateUser :execresult
INSERT INTO user (id, email, hashed_password, avatar_url) VALUES (
    ?, ?, ?, ?
);

-- name: CreateClient :execresult
INSERT INTO client (
    id,
    secret_hash,
    name,
    auth_callback_urls,
    logout_callback_urls,
    picture_url,
    backchannel_logout_url,
    token_expiration
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: CreateSession :execresult
INSERT INTO session (
    id,
    user_id,
    client_id,
    expires_at,
    os,
    browser
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: CreateAuthzHistory :execresult
INSERT INTO authorization_history (user_id, client_id) VALUES (?, ?);

-- name: CreateAuthzCode :execresult
INSERT INTO authorization_code (
    id,
    user_id,
    client_id,
    scopes,
    os,
    browser
) VALUES (
    ?, ?, ?, ?, ?, ?
);


-- name: UpdateUserPasswordHash :exec
UPDATE user
SET hashed_password = ?
WHERE id = ?;

-- name: UpdateUserAvatar :exec
UPDATE user
SET avatar_url = ?
WHERE id = ?;

-- name: UpdateClient :exec
UPDATE client
SET name = ?,
    auth_callback_urls = ?,
    logout_callback_urls = ?,
    picture_url = ?,
    backchannel_logout_url = ?,
    token_expiration = ?
WHERE id = ?;


-- name: DeleteClient :exec
DELETE FROM client
WHERE id = ?;

-- name: DeleteSession :exec
DELETE FROM session
WHERE id = ? OR expires_at <= NOW();

-- name: DeleteAuthzCode :exec
DELETE FROM authorization_code
WHERE id = ?;

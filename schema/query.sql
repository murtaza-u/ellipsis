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

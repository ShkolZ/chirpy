-- name: CreateRefreshTokenForUser :one
 INSERT INTO refresh_tokens(token, expires_at, revoked_at, created_at, updated_at, user_id)
 VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
 ) RETURNING *;

-- name: GetTokenbyToken :one 
 SELECT * FROM refresh_tokens
 WHERE token = $1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = $2,
    updated_at = $3
WHERE token = $1;
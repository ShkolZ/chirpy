-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: UpdateCredentials :one
UPDATE users
SET email = $2,
    password = $3
WHERE id = $1
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one 
SELECT * FROM users
WHERE email = $1;

-- name: SetChirpyRedTrue :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;
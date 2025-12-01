-- name: CreateChirp :one
INSERT INTO chirps(id, body, created_at, updated_at, user_id )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;
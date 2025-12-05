-- name: CreateChirp :one
INSERT INTO chirps(id, body, created_at, updated_at, user_id )
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
) RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpById :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirpById :exec 
DELETE FROM chirps
WHERE id = $1 and user_id = $2;


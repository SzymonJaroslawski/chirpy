-- name: CreateChirp :one
INSERT INTO chirps (created_at, updated_at, body, user_id)
VALUES (
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: GetChirpWithId :one
SELECT * FROM chirps 
WHERE id = $1;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

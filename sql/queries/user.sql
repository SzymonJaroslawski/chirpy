-- name: CreateUser :one
INSERT INTO users (created_at, updated_at, email, hashed_passowrd)
VALUES (
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserWithEmail :one
SELECT * FROM users 
WHERE email = $1;

-- name: UpdateUserEmailAndPassword :one 
UPDATE users 
SET email = $1, hashed_passowrd = $2
WHERE id = $3
RETURNING *;

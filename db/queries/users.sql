-- name: CreateUser :exec
INSERT INTO users (id, hospital_id, first_name, last_name, email, password, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetUser :one
SELECT *
FROM users
WHERE id = ?
  AND deleted_at IS NULL;

-- name: GetHospitalUsers :many
SELECT *
FROM users
WHERE hospital_id = ?
  AND deleted_at IS NULL;


-- name: GetHospitalUsersByRole :many
SELECT *
FROM users
WHERE hospital_id = ?
  AND role = ?
  AND deleted_at IS NULL;

-- name: UpdateUser :execresult
UPDATE users
SET hospital_id = ?,
    first_name  = ?,
    last_name   = ?,
    email       = ?,
    password    = ?,
    role        = ?,
    updated_at  = ?
WHERE id = ?
  AND deleted_at IS NULL;
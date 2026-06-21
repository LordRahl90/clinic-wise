-- name: CreateHospital :exec
INSERT INTO hospitals (id, name, created_at, updated_at)
VALUES (?, ?, ?, ?);

-- name: GetHospital :one
SELECT *
FROM hospitals
WHERE id = ?
  AND deleted_at IS NULL;


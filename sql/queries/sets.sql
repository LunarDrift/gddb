-- name: CreateSet :one
INSERT INTO sets (
show_id,
set_name,
position
)
VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT (show_id, position) DO UPDATE
SET set_name = EXCLUDED.set_name
RETURNING id;

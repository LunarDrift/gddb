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
RETURNING id;

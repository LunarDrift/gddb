-- name: CreateSetEntry :exec
INSERT INTO set_entries (
set_id,
raw_entry,
position
)
VALUES (
  $1,
  $2,
  $3
  )
ON CONFLICT (set_id, position) DO NOTHING;

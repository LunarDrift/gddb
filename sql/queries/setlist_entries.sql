-- name: InsertSetlistEntry :exec
INSERT INTO setlist_entries (set_id, raw_entry, position)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

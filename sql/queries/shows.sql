-- name: InsertShow :exec
INSERT INTO shows (show_id, date, day, venue, location, notes)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (show_id) DO NOTHING;

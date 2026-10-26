-- name: InsertEntrySong :exec
INSERT INTO set_entry_songs (set_entry_id, song_name)
VALUES ($1, $2);

-- name: DeleteEntrySongs :exec
DELETE FROM set_entry_songs WHERE set_entry_id = $1;

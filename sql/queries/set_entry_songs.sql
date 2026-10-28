-- name: InsertEntrySong :exec
INSERT INTO set_entry_songs (set_entry_id, song_name)
VALUES ($1, $2);

-- name: DeleteEntrySongs :exec
DELETE FROM set_entry_songs WHERE set_entry_id = $1;

-- name: ListSongStats :many
SELECT
    ses.song_name AS title,
    COUNT(*) AS times_played,
    MIN(sh.show_date)::date AS first_played,
    MAX(sh.show_date)::date AS last_played,
    COUNT(*) FILTER (WHERE st.set_name = 'set_1') AS set_1,
    COUNT(*) FILTER (WHERE st.set_name IN ('set_2', 'set_3')) AS set_2_3,
    COUNT(*) FILTER (WHERE st.set_name = 'encore') AS encore,
    COUNT(*) FILTER (WHERE st.set_name IN ('acoustic', 'acoustic_1', 'acoustic_2')) AS acoustic,
    COUNT(*) FILTER (WHERE st.set_name = 'electric') AS electric
FROM set_entry_songs ses
JOIN set_entries se ON se.id = ses.set_entry_id
JOIN sets st ON st.id = se.set_id
JOIN shows sh ON sh.show_id = st.show_id
GROUP BY ses.song_name
ORDER BY ses.song_name;

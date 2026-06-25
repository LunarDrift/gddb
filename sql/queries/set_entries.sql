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

-- name: MostPlayedSongs :many
SELECT se.song_name AS song, count(*) AS times_played
FROM set_entries se
GROUP BY se.song_name
ORDER BY times_played DESC;

-- name: SongsPlayedLessThan :many
SELECT se.song_name AS song, count(*) AS times_played
FROM set_entries se
GROUP BY se.song_name
HAVING count(*) < $1
ORDER BY times_played DESC;

-- name: MostCommonEncore :many
SELECT se.song_name AS song, count(*) AS times_played
FROM set_entries se 
JOIN "sets" s ON se.set_id = s.id
WHERE s.set_name = 'encore'
GROUP BY se.song_name 
ORDER BY times_played  DESC;

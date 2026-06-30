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

-- name: MostCommonSongsBySetName :many
SELECT se.song_name AS song, count(*) AS times_played
FROM set_entries se 
JOIN "sets" s ON se.set_id = s.id
WHERE s.set_name = $1
GROUP BY se.song_name 
ORDER BY times_played  DESC;

-- name: UniqueSongsPerCity :many
SELECT sh.city, sh.state AS state_or_country, count(DISTINCT se.song_name) AS unique_song_count
FROM set_entries se
JOIN "sets" s ON se.set_id = s.id
JOIN shows sh ON s.show_id = sh.show_id
GROUP BY sh.city, sh.state
ORDER BY unique_song_count DESC;

-- name: AllSongsPlayedAtVenue :many
SELECT DISTINCT se.song_name, sh.venue, sh.city, sh.state
FROM set_entries se
JOIN "sets" s ON se.set_id = s.id
JOIN shows sh ON s.show_id = sh.show_id
WHERE sh.venue ILIKE $1
ORDER BY sh.venue, se.song_name;

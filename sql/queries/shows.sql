-- name: CreateShow :exec
INSERT INTO shows (
show_id,
show_date,
day,
city,
state,
venue,
notes
)
VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
)
ON CONFLICT (show_id) DO NOTHING;

-- name: GetShowDetailsFromDate :many
SELECT 
  sh.show_id,
	sh.show_date,
	sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
	st.set_name,
	st.position AS set_position,
	se.raw_entry,
	se.position AS song_position
FROM
	shows sh
LEFT JOIN "sets" st ON
	st.show_id = sh.show_id
LEFT JOIN set_entries se ON
	se.set_id = st.id
WHERE
	sh.show_date = $1
ORDER BY
  sh.show_id,
	st."position",
	se."position";

-- name: GetShowDetailsFromID :many
SELECT
  sh.show_id,
	sh.show_date,
	sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
	st.set_name,
	se.raw_entry
FROM
	shows sh
LEFT JOIN SETS st
	ON
	sh.show_id = st.show_id
LEFT JOIN set_entries se
	ON
	st.id = se.set_id
WHERE
	sh.show_id = $1
ORDER BY
	st.position,
	se.position;

-- name: GetAllShows :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM
  shows sh
ORDER BY
  sh.show_date,
  sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetShowsFromTodaysDate :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM
  shows sh
WHERE EXTRACT(MONTH FROM sh.show_date) = EXTRACT(MONTH FROM CURRENT_DATE)
  AND EXTRACT(DAY FROM sh.show_date) = EXTRACT(DAY FROM CURRENT_DATE)
ORDER BY sh.show_date
LIMIT @page_limit OFFSET @page_offset;

-- name: GetAllShowIDs :many
SELECT show_id FROM shows ORDER BY show_id;

-- name: GetShowsBetweenDates :many
SELECT
  sh.show_id,
	sh.show_date,
	sh.venue,
	sh.city,
	sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM
	shows sh
WHERE
sh.show_date BETWEEN @start_date AND @end_date
GROUP BY
	sh.show_date, sh.venue, sh.show_id 
ORDER BY
	sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: SongStats :one
SELECT
  count(*) AS times_played,
  min(sh.show_date)::date AS first_played,
  max(sh.show_date)::date AS last_played
FROM shows sh 
JOIN "sets" st ON st.show_id = sh.show_id
JOIN set_entries se ON se.set_id = st.id
WHERE se.song_name ILIKE @song_name;

-- name: GetShowsFromSetName :many
SELECT
	sh.show_id,
	sh.show_date,
	sh.venue,
	sh.city,
	sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
JOIN "sets" s ON s.show_id = sh.show_id 
WHERE s.set_name = $1
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetAcousticShows :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
JOIN "sets" st ON st.show_id = sh.show_id
WHERE st.set_name IN ('acoustic', 'acoustic_1', 'acoustic_2')
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: ShowsWithShowNotes :many
SELECT 
	sh.show_id,
	sh.show_date,
	sh.venue,
	sh.city,
	sh.state AS location,
	sh.notes,
  COUNT(*) OVER ()
FROM shows sh
WHERE sh.notes IS NOT NULL AND sh.notes != ''
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: ShowsWithoutNotes :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  COUNT(*) OVER ()
FROM shows sh
WHERE sh.notes IS NULL OR sh.notes = ''
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetValidLocations :many
SELECT DISTINCT LOWER(state) AS location FROM shows;

-- name: GetAllVenues :many
SELECT DISTINCT venue FROM shows ORDER BY venue;

-- name: GetAllLocations :many
SELECT state AS "location"
FROM shows
GROUP BY "location"
ORDER BY
  -- states first, then countries
  CASE WHEN length(shows.state) = 2 THEN 0 ELSE 1 END,
  "location";

-- name: GetAllCities :many
SELECT DISTINCT city FROM shows ORDER BY city;

-- name: ListVenueStats :many
SELECT
  s.city,
  s.state AS location,
  s.venue,
  COUNT(s.show_id) AS shows_played
FROM shows s
GROUP BY s.city, s.state, s.venue
ORDER BY
  CASE WHEN length(s.state) = 2 THEN 0 ELSE 1 END,
  s.state, s.city, s.venue;

-- name: GetSongDebuts :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
JOIN show_footnotes sf ON sf.show_id = sh.show_id
WHERE sf.note_text ILIKE '%first%time%played%'
  AND sf.note_text NOT ILIKE '%since%'
  AND sf.note_text NOT ILIKE '%electric%'
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: SearchShowsWithFilters :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
JOIN "sets" st ON st.show_id = sh.show_id
JOIN set_entries se ON se.set_id = st.id
WHERE (sqlc.narg('year')::int IS NULL OR EXTRACT(YEAR FROM sh.show_date) = sqlc.narg('year'))
  AND (sqlc.narg('month')::int IS NULL OR EXTRACT(MONTH FROM sh.show_date) = sqlc.narg('month'))
  AND (sqlc.narg('day')::int IS NULL OR EXTRACT(DAY FROM sh.show_date) = sqlc.narg('day'))
  AND (sqlc.narg('location')::text IS NULL OR LOWER(sh.state) = LOWER(sqlc.narg('location')))
  AND (sqlc.narg('city')::text IS NULL OR sh.city ILIKE sqlc.narg('city'))
  AND (sqlc.narg('song')::text IS NULL OR se.song_name ILIKE sqlc.narg('song'))
  AND (sqlc.narg('venue')::text IS NULL OR sh.venue ILIKE sqlc.narg('venue'))
  AND (sqlc.narg('notes')::text IS NULL OR sh.notes ILIKE sqlc.narg('notes'))
GROUP BY sh.show_id
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

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

-- name: GetShowFromDate :many
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

-- name: GetShowFromID :many
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

-- name: SearchByVenue :many
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
WHERE venue ILIKE @venue_name
ORDER BY
  sh.show_id,
	sh.venue,
	sh.show_date
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

-- name: GetShowsFromSongName :many
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
WHERE se.raw_entry ILIKE @song_name
ORDER BY sh.show_date, sh.show_id
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

-- name: GetShowsFromYearAndLocation :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
WHERE EXTRACT(YEAR FROM sh.show_date) = @year::int
AND LOWER(sh.state) = LOWER(@location)
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetShowsFromYear :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
WHERE EXTRACT(YEAR FROM sh.show_date) = @year::int
ORDER BY sh.show_date, show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetShowsFromLocation :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
WHERE LOWER(sh.state) = LOWER(@location)
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetShowsFromCity :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes,
  COUNT(*) OVER ()
FROM shows sh
WHERE sh.city ILIKE @city
ORDER BY sh.show_date, sh.show_id
LIMIT @page_limit OFFSET @page_offset;

-- name: GetValidLocations :many
SELECT DISTINCT LOWER(state) AS location FROM shows;

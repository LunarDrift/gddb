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
  shows.show_id,
	shows.show_date,
	shows.venue,
  shows.city,
  shows.state AS location,
  shows.notes,
	sets.set_name,
	sets.position AS set_position,
	set_entries.raw_entry,
	set_entries.position AS song_position
FROM
	shows
LEFT JOIN SETS ON
	sets.show_id = shows.show_id
LEFT JOIN set_entries ON
	set_entries.set_id = sets.id
WHERE
	shows.show_date = $1
ORDER BY
  shows.show_id,
	sets.position,
	set_entries."position";

-- name: GetShowFromID :many
SELECT
  s.show_id,
	s.show_date,
	s.venue,
  s.city,
  s.state AS location,
  s.notes,
	st.set_name,
	se.raw_entry
FROM
	shows s
LEFT JOIN SETS st
	ON
	s.show_id = st.show_id
LEFT JOIN set_entries se
	ON
	st.id = se.set_id
WHERE
	s.show_id = $1
ORDER BY
	st.position,
	se.position;

-- name: SearchByVenue :many
SELECT
  shows.show_id,
	shows.show_date,
	shows.venue,
  shows.city,
  shows.state AS location,
  shows.notes
FROM
	shows
WHERE venue ILIKE $1
ORDER BY
  shows.show_id,
	shows.venue,
	shows.show_date;

-- name: GetAllShowIDs :many
SELECT show_id FROM shows ORDER BY show_id;

-- name: GetShowsBetweenDates :many
SELECT
  s.show_id,
	s.show_date,
	s.venue,
	s.city,
	s.state AS location,
  s.notes
FROM
	shows s
JOIN "sets" st ON 
	s.show_id = st.show_id 
JOIN set_entries se ON
	st.id = se.set_id 
WHERE
	s.show_date BETWEEN $1 AND $2
GROUP BY
	s.show_date, s.venue, s.show_id 
ORDER BY
	s.show_date;

-- name: GetShowsFromSongName :many
SELECT
  s.show_id,
  s.show_date,
  s.venue,
  s.city,
  s.state AS location,
  s.notes
FROM shows s
JOIN "sets" st ON st.show_id = s.show_id
JOIN set_entries se ON se.set_id = st.id
WHERE se.raw_entry ILIKE $1
ORDER BY show_date;

-- name: SongStats :one
SELECT
  count(*) AS times_played,
  min(sh.show_date)::date AS first_played,
  max(sh.show_date)::date AS last_played
FROM shows sh 
JOIN "sets" s ON s.show_id = sh.show_id 
JOIN set_entries se ON se.set_id = s.id 
WHERE se.song_name ILIKE $1;

-- name: GetShowsFromSetName :many
SELECT
	sh.show_id,
	sh.show_date,
	sh.venue,
	sh.city,
	sh.state AS location,
  sh.notes
FROM shows sh
JOIN "sets" s ON s.show_id = sh.show_id 
WHERE s.set_name = $1;

-- name: ShowsWithShowNotes :many
SELECT 
	sh.show_id,
	sh.show_date,
	sh.venue,
	sh.city,
	sh.state AS location,
	sh.notes
FROM shows sh
WHERE sh.notes IS NOT NULL AND sh.notes != '';

-- name: ShowsWithoutNotes :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location
FROM shows sh
WHERE sh.notes IS NULL OR sh.notes = '';

-- name: GetShowsFromYearAndLocation :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes
FROM shows sh
WHERE EXTRACT(YEAR FROM sh.show_date) = @year::int
AND LOWER(sh.state) = LOWER(@location)
ORDER BY sh.show_date;

-- name: GetShowsFromYear :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes
FROM shows sh
WHERE EXTRACT(YEAR FROM sh.show_date) = @year::int
ORDER BY sh.show_date;

-- name: GetShowsFromLocation :many
SELECT
  sh.show_id,
  sh.show_date,
  sh.venue,
  sh.city,
  sh.state AS location,
  sh.notes
FROM shows sh
WHERE LOWER(sh.state) = LOWER(@location)
ORDER BY sh.show_date;

-- name: GetValidLocations :many
SELECT DISTINCT LOWER(state) AS location FROM shows;

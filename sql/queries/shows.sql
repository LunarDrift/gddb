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
	shows.show_date,
	shows.venue,
	sets.set_name,
	sets.position AS set_position,
	set_entries.raw_entry,
	set_entries.position AS song_position
FROM
	shows
JOIN SETS ON
	sets.show_id = shows.show_id
JOIN set_entries ON
	set_entries.set_id = sets.id
WHERE
	shows.show_date = $1
ORDER BY
  shows.show_id,
	sets.position,
	set_entries."position";

-- name: GetShowFromID :many
SELECT
	s.show_date,
	s.venue,
	st.set_name,
	se.raw_entry
FROM
	shows s
JOIN SETS st
	ON
	s.show_id = st.show_id
JOIN set_entries se
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
	shows.show_date AS "date",
	shows.venue,
  shows.city,
  shows.state
FROM
	shows
WHERE venue ILIKE $1
ORDER BY
  shows.show_id,
	shows.venue,
	shows.show_date;

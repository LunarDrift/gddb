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
	sets.position,
	set_entries."position";

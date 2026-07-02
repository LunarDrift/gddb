-- name: CreateFootnote :exec
INSERT INTO show_footnotes (show_id, marker, note_text)
VALUES ($1, $2, $3)
ON CONFLICT (show_id, marker) DO NOTHING
RETURNING *;

-- name: GetFootnotesFromShowID :many
SELECT marker, note_text
FROM show_footnotes
WHERE show_id = $1;

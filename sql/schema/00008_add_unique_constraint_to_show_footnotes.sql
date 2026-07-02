-- +goose Up
ALTER TABLE show_footnotes
ADD constraint show_footnotes_show_id_marker_key UNIQUE (show_id, marker);

-- +goose Down
ALTER TABLE show_footnotes
DROP constraint show_footnotes_show_id_marker_key ;

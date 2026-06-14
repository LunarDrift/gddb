-- +goose Up
CREATE INDEX idx_shows_show_date
ON shows (show_date);

CREATE INDEX idx_sets_show_id
ON sets (show_id);

CREATE INDEX idx_entries_set_id
ON set_entries (set_id);

-- +goose Down
DROP INDEX idx_entries_set_id;
DROP INDEX idx_sets_show_id;
DROP INDEX idx_shows_show_date;

-- +goose Up
ALTER TABLE set_entries
   ADD COLUMN song_name TEXT GENERATED ALWAYS AS (trim(trailing ' *' from raw_entry)) STORED;

CREATE INDEX idx_set_entries_song_name ON set_entries(song_name);

-- +goose Down
DROP INDEX idx_set_entries_song_name;
ALTER TABLE set_entries DROP COLUMN song_name;

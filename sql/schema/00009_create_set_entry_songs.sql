-- +goose Up
CREATE TABLE set_entry_songs (
  id BIGSERIAL PRIMARY KEY,
  set_entry_id BIGINT NOT NULL REFERENCES set_entries(id) ON DELETE CASCADE,
  song_name TEXT NOT NULL
);
CREATE INDEX idx_set_entry_songs_song_name ON set_entry_songs (song_name);

-- +goose Down
DROP TABLE set_entry_songs;

-- +goose Up
CREATE TABLE show_footnotes (
    id SERIAL PRIMARY KEY,
    show_id INTEGER NOT NULL REFERENCES shows (show_id),
    marker TEXT NOT NULL,
    note_text TEXT NOT NULL
);

-- +goose Down
DROP TABLE show_footnotes;

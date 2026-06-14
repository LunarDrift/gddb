-- +goose Up
CREATE TABLE shows (
    show_id INTEGER PRIMARY KEY,
    date DATE NOT NULL,
    day TEXT,
    venue TEXT NOT NULL,
    location TEXT NOT NULL,
    notes TEXT
);

-- +goose Down
DROP TABLE shows;

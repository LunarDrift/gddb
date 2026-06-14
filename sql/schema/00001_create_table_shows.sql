-- +goose Up
CREATE TABLE shows (
    show_id INTEGER PRIMARY KEY,
    show_date DATE NOT NULL,
    day TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    venue TEXT NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE shows;

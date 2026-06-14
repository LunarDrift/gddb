-- +goose Up
CREATE TABLE sets (
id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
show_id INTEGER NOT NULL REFERENCES shows (show_id) ON DELETE CASCADE,
set_name TEXT NOT NULL,   -- 'set_1', 'set_2', 'encore' etc.
position INTEGER NOT NULL, -- order among sets for this show
created_at TIMESTAMP NOT NULL DEFAULT now (),

UNIQUE (show_id, position)
) ;

-- +goose Down
DROP TABLE sets ;

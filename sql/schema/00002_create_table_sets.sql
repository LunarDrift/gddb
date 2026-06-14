-- +goose Up
CREATE TABLE sets (
id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
show_id INTEGER NOT NULL REFERENCES shows (show_id),
set_name TEXT NOT NULL,   -- 'set_1', 'set_2', 'encore' etc.
position INTEGER NOT NULL -- order among sets for this show
) ;

-- +goose Down
DROP TABLE sets ;

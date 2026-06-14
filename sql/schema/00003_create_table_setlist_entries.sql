-- +goose Up
CREATE TABLE set_entries (
id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
set_id INTEGER NOT NULL REFERENCES sets (id) ON DELETE CASCADE,
raw_entry TEXT NOT NULL,   -- "Help On The Way > Slipknot! > Franklin's Tower"
position INTEGER NOT NULL,  -- order within the set
crated_at TIMESTAMP NOT NULL DEFAULT now (),

UNIQUE (set_id, position)
) ;

-- +goose Down
DROP TABLE set_entries ;

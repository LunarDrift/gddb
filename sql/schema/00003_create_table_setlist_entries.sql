-- +goose Up
CREATE TABLE setlist_entries (
id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
set_id INTEGER NOT NULL REFERENCES sets (id),
raw_entry TEXT NOT NULL,   -- "Help On The Way > Slipknot! > Franklin's Tower"
position INTEGER NOT NULL  -- order within the set
) ;

-- +goose Down
DROP TABLE setlist_entries ;

-- +goose Up
ALTER TABLE set_entries
RENAME column crated_at TO created_at ;

-- +goose Down
ALTER TABLE set_entries
RENAME COLUMN created_at TO crated_at ;

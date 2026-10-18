-- +goose Up
ALTER TABLE campaigns ADD COLUMN scenario_type varchar(32) NOT NULL DEFAULT 'standard';
ALTER TABLE campaigns ADD COLUMN file_type varchar(16) NOT NULL DEFAULT '';

-- +goose Down
-- Backward migration deliberately not implemented for SQLite compatibility.

-- +goose Up
ALTER TABLE campaigns ADD COLUMN file_name varchar(80) NOT NULL DEFAULT '';
ALTER TABLE campaigns ADD COLUMN preview_mode varchar(32) NOT NULL DEFAULT 'standard';

-- +goose Down
ALTER TABLE campaigns DROP COLUMN preview_mode;
ALTER TABLE campaigns DROP COLUMN file_name;

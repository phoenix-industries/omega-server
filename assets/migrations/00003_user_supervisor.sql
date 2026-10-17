-- +goose Up
ALTER TABLE users ALTER COLUMN gender DROP NOT NULL;
ALTER TABLE users ADD COLUMN is_supervisor boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE users ALTER COLUMN gender SET NOT NULL;
ALTER TABLE users DROP COLUMN is_supervisor;

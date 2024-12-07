-- +goose Up 
ALTER TABLE users
ADD hashed_passowrd TEXT DEFAULT 'unset' NOT NULL;

-- +goose Down 
ALTER TABLE users
DROP COLUMN hashed_passowrd;

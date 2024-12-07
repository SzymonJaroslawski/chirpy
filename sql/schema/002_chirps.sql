-- +goose Up 
CREATE TABLE chirps (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  body VARCHAR(140) NOT NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;

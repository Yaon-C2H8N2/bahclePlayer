ALTER TABLE users ADD COLUMN IF NOT EXISTS token VARCHAR(255);

ALTER TABLE users ADD COLUMN IF NOT EXISTS token_created_at TIMESTAMP;

ALTER TYPE config_type ADD VALUE IF NOT EXISTS 'QUEUE_METHOD';

ALTER TYPE config_type ADD VALUE IF NOT EXISTS 'PLAYLIST_METHOD';
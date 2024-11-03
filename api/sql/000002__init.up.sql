ALTER TABLE users_videos ADD COLUMN IF NOT EXISTS thumbnail_url varchar;

ALTER TABLE users_videos ADD COLUMN IF NOT EXISTS added_by varchar;
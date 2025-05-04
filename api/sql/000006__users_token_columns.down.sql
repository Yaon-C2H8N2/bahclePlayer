ALTER TABLE users DROP COLUMN IF EXISTS token_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS refresh_token;
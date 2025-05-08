CREATE TABLE IF NOT EXISTS users_overlays
(
    overlay_type_id integer NOT NULL REFERENCES overlay_types (overlay_type_id) ON DELETE CASCADE,
    user_id         integer NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
    settings        jsonb   NOT NULL,
    PRIMARY KEY (overlay_type_id, user_id)
);

ALTER TABLE overlay_types
    ADD COLUMN IF NOT EXISTS overlay_code VARCHAR NOT NULL UNIQUE DEFAULT '';

INSERT INTO overlay_types (name, description, schema, overlay_code)
VALUES ('Currently playing', 'Displays the currently playing song', '{}', 'currently_playing');
CREATE TABLE overlay_types (
    overlay_type_id serial primary key,
    name varchar not null,
    description varchar,
    schema jsonb not null
);
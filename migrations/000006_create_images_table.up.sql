CREATE TABLE IF NOT EXISTS images (
    id bigserial PRIMARY KEY,
    url text NOT NULL UNIQUE
);
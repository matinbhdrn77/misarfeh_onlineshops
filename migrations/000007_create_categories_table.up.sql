CREATE TABLE IF NOT EXISTS categories (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    img_url text
);
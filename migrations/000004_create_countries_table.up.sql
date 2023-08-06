CREATE TABLE IF NOT EXISTS countries (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS shops (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    year integer NOT NULL,
    description text,
    follower_count integer,
    telegram text,
    instagram text,
    phone text,
    logo_url text, 
    rating real,
    rating_count integer NOT NULL DEFAULT 0,
    verified boolean NOT NULL DEFAULT FALSE,
    delivery_time integer NOT NULL DEFAULT 1  
);
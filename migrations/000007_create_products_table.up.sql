CREATE TABLE IF NOT EXISTS products (
    id bigserial,
    shop_id INTEGER REFERENCES shops(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    country_id INTEGER REFERENCES countries(id) ON DELETE SET NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    description text NOT NULL,
    price real,
    sale_price INTEGER NOT NULL,
    off INTEGER NOT NULL DEFAULT 0,
    brand text NOT NULL,
    PRIMARY KEY(id,shop_id)
);
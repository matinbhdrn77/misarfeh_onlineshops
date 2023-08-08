CREATE TABLE IF NOT EXISTS images (
    id bigserial PRIMARY KEY,
    url text NOT NULL UNIQUE,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    shop_id INTEGER REFERENCES shops(id) ON DELETE CASCADE
);

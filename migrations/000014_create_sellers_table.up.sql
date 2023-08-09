CREATE TABLE IF NOT EXISTS sellers (
    id bigserial PRIMARY KEY,
    meli_code text,
    meli_cart_url text,
    FOREIGN KEY(id) REFERENCES users(id) 
);
ALTER TABLE products ADD CONSTRAINT products_off_check CHECK (off BETWEEN 0 AND 100);

ALTER TABLE products ADD CONSTRAINT products_price_check CHECK (price > 0);

ALTER TABLE products ADD CONSTRAINT products_sale_price_check CHECK(sale_price > 0);
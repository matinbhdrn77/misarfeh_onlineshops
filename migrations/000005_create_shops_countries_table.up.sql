CREATE TABLE IF NOT EXISTS shops_countries(
    shop_id INTEGER REFERENCES shops(id) ON DELETE CASCADE,
    country_id INTEGER REFERENCES countries(id) ON DELETE CASCADE,
    CONSTRAINT shops_countries_pk PRIMARY KEY(shop_id,country_id) 
);
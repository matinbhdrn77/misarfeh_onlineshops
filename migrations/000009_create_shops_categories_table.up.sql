CREATE TABLE IF NOT EXISTS shops_categories(
    shop_id INTEGER REFERENCES shops(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES categories(id) ON DELETE CASCADE,
    CONSTRAINT shops_categories_pk PRIMARY KEY(shop_id,category_id) 
);
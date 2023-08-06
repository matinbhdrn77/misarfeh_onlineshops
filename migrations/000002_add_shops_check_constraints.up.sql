ALTER TABLE shops ADD CONSTRAINT shops_year_check CHECK (year BETWEEN 1888 AND date_part('year', now()));

ALTER TABLE shops ADD CONSTRAINT shops_delivery_time_check CHECK (delivery_time BETWEEN 1 AND 100);

ALTER TABLE shops ADD CONSTRAINT shops_rating_check CHECK(rating BETWEEN 0 AND 5)
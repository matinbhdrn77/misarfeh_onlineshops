CREATE EXTENSION pg_trgm;

CREATE INDEX IF NOT EXISTS shops_title_tgrm_idx ON shops USING gin (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS shops_instagram_tgrm_idx ON shops USING gin (instagram gin_trgm_ops);
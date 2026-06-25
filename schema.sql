-- Schema for the URL shortener database.
--
-- 1. Create the database (run once):
--      CREATE DATABASE url_shortener;
-- 2. Apply this file against it:
--      psql -d url_shortener -f schema.sql

CREATE TABLE IF NOT EXISTS url_mappings (
    id         SERIAL PRIMARY KEY,
    long_url   TEXT        NOT NULL,
    short_url  VARCHAR(16) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The UNIQUE constraint on short_url also creates the index used by the
-- redirect lookup (SELECT ... WHERE short_url = $1).

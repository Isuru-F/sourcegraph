BEGIN;
-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

-- Create pgcrypto exension to be able to use utility functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Alter auto generated column
ALTER TABLE repo
ADD COLUMN IF NOT EXISTS hashed_name TEXT GENERATED ALWAYS AS (ENCODE(SHA256(LOWER(name)::BYTEA), 'HEX')) STORED;

-- Create index on auto-generated column
CREATE INDEX IF NOT EXISTS repo_hashed_names_idx ON repo USING btree (hashed_name);

COMMIT;

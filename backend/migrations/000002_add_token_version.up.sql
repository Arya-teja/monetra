-- Every access token carries the version it was issued with. Logging out increments
-- the version, which makes all earlier tokens of the user invalid.
ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0;

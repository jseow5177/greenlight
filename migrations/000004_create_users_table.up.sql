CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  -- Stores text data exactly as inputted, but comparisons are case-insensitive.
  -- Combined with citext type, the UNIQUE constraint means that no two rows can have the same value - even if different cases.
  email citext UNIQUE NOT NULL,
  password_hash bytea NOT NULL, -- binary string type column
  activated bool NOT NULL, -- To denote whether a user account is active
  version integer NOT NULL DEFAULT 1
);
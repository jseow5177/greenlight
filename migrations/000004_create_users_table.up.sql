CREATE TABLE IF NOT EXISTS users (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  name text NOT NULL,
  email citext UNIQUE NOT NULL, -- case-insensitive text
  password_hash bytea NOT NULL, -- binary string type column
  activated bool NOT NULL, -- To denote whether a user account is active
  version integer NOT NULL DEFAULT 1
);
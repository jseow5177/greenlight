CREATE TABLE IF NOT EXISTS movies (
  id bigserial PRIMARY KEY, -- 64-bit auto incrementing integer starting at 1
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(), -- Timestamp with time zone at precision 0
  title text NOT NULL,
  year integer NOT NULL,
  runtime integer NOT NULL,
  genres text[] NOT NULL, -- An array of zero or more text values
  version integer NOT NULL DEFAULT 1
);
-- Start psql terminal session as superuser postgres
-- psql -U postgres

-- Start psql as postgress user at greenlight database in localhost
-- psql --host=localhost --dbname=greenlight --username=postgres

-- meta commands
-- \c <database_name> : Connect to database_name
-- \l : List all databases
-- \dt : List tables
-- \du : List users

-- See current user
SELECT current_user;

-- See current datase
SELECT current_database();

-- Create a database called greenlight
CREATE DATABASE greenlight;

-- Create a new greenlight user
-- Can only be done by a superuser
CREATE ROLE greenlight WITH LOGIN PASSWORD 'pa55word';

-- Add citext (case-insensitive text) extension to greenlight database
-- Note that extension can only be added by superuser to a specific database
CREATE EXTENSION IF NOT EXISTS citext;
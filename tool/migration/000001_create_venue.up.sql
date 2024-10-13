-- Create the venues table
CREATE TABLE IF NOT EXISTS venues(
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    city text NOT NULL,
    country text NOT NULL
);
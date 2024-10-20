-- Create the sections table
CREATE TABLE IF NOT EXISTS sections(
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    venue_id bigint NOT NULL REFERENCES venues (id) ON DELETE CASCADE
);

-- Create the rows table
CREATE TABLE IF NOT EXISTS rows(
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    section_id bigint NOT NULL REFERENCES sections (id) ON DELETE CASCADE
);

-- Create the seats table
CREATE TABLE IF NOT EXISTS seats(
    id bigserial PRIMARY KEY,
    seat_number int NOT NULL,
    row_id bigint NOT NULL REFERENCES rows (id) ON DELETE CASCADE
);
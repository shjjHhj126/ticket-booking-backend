CREATE TABLE IF NOT EXISTS artists (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text
);

CREATE TABLE IF NOT EXISTS events (
    id bigserial PRIMARY KEY,
    name text NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status text NOT NULL,
    venue_id bigint REFERENCES venues (id) ON DELETE CASCADE,
    artist_id bigint REFERENCES artists (id) ON DELETE CASCADE,
    description text
);

CREATE TABLE IF NOT EXISTS event_seat (
    id bigserial PRIMARY KEY,
    seat_id bigint REFERENCES seats(id) ON DELETE CASCADE,
    event_id bigint REFERENCES events(id) ON DELETE CASCADE,
    price int NOT NULL,
    UNIQUE (seat_id, event_id)
);

CREATE TABLE IF NOT EXISTS bookings (       
    id bigserial PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    event_seat_id bigint REFERENCES event_seat(id) ON DELETE CASCADE,   
    booked_by bigint REFERENCES users(id) ON DELETE SET NULL
);

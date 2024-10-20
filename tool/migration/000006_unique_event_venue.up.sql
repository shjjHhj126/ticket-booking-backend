ALTER TABLE events
ADD CONSTRAINT unique_event_venue UNIQUE (id, venue_id);
ALTER TABLE seats
ADD CONSTRAINT unique_row_seat
UNIQUE (row_id, seat_number);
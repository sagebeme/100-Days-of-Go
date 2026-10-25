CREATE TABLE bookings (
    id        bigserial PRIMARY KEY,
    trip_id   bigint NOT NULL REFERENCES trips (id) ON DELETE CASCADE,
    seat      integer NOT NULL CHECK (seat > 0),
    passenger text NOT NULL,
    phone     text NOT NULL,
    UNIQUE (trip_id, seat)
);

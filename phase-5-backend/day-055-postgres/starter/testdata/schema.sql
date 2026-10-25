-- The booking API's tables. Day 56 turns this into numbered migrations.
-- "from" and "to" are reserved words in SQL, so the columns are from_stage and to_stage.

CREATE TABLE IF NOT EXISTS routes (
    id         bigserial PRIMARY KEY,
    name       text NOT NULL,
    from_stage text NOT NULL,
    to_stage   text NOT NULL
);

CREATE TABLE IF NOT EXISTS trips (
    id       bigserial PRIMARY KEY,
    route_id bigint NOT NULL REFERENCES routes (id),
    departs  timestamptz NOT NULL,
    seats    integer NOT NULL CHECK (seats > 0)
);

CREATE TABLE IF NOT EXISTS bookings (
    id        bigserial PRIMARY KEY,
    trip_id   bigint NOT NULL REFERENCES trips (id) ON DELETE CASCADE,
    seat      integer NOT NULL CHECK (seat > 0),
    passenger text NOT NULL,
    phone     text NOT NULL,
    UNIQUE (trip_id, seat)
);

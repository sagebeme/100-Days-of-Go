CREATE TABLE trips (
    id       bigserial PRIMARY KEY,
    route_id bigint NOT NULL REFERENCES routes (id),
    departs  timestamptz NOT NULL,
    seats    integer NOT NULL CHECK (seats > 0)
);

CREATE INDEX trips_route_id_departs_idx ON trips (route_id, departs);

CREATE TABLE routes (
    id         bigserial PRIMARY KEY,
    name       text NOT NULL,
    from_stage text NOT NULL,
    to_stage   text NOT NULL
);

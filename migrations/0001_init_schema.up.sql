CREATE TABLE incidents (
    id serial primary key,
    title text not null unique,
    description text,
    lat double precision not null,
    long double precision not null,
    radius_m integer not null check (radius_m > 0),
    active boolean not null default true,
    created_at timestamp,
    updated_at timestamp
);
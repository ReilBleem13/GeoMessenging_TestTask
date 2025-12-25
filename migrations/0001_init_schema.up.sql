CREATE TABLE incidents (
    id serial primary key,
    title text not null unique,
    description text,
    lat double precision not null,
    long double precision not null,
    radius_m integer not null,
    active boolean not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);
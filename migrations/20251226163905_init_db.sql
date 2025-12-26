-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE incidents (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL UNIQUE,
    description TEXT,
    lat         DOUBLE PRECISION NOT NULL,
    long        DOUBLE PRECISION NOT NULL,
    radius_m    INTEGER NOT NULL,
    active      BOOLEAN NOT NULL,
    geom        GEOGRAPHY(Point, 4326) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX incidents_geom_gist_idx
ON incidents
USING GIST (geom);

CREATE TABLE location_checks (
    id              SERIAL PRIMARY KEY,
    user_id         TEXT NOT NULL,
    lat             DOUBLE PRECISION NOT NULL,
    long            DOUBLE PRECISION NOT NULL,
    in_danger_zone  BOOLEAN NOT NULL,
    nearest_id      INTEGER REFERENCES incidents(id) ON DELETE SET NULL,
    checked_at      TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE incidents;
DROP TABLE location_checks;
-- +goose StatementEnd

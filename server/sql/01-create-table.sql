CREATE TABLE IF NOT EXISTS continent (
    index bigserial PRIMARY KEY,
    uuid uuid NOT NULL UNIQUE,
    name text NOT NULL,
    type smallint,
    area_by_km2 float,
    created timestamp DEFAULT NOW(),
    updated timestamp,
    creator jsonb,
    deleted_state smallint default 0
);

CREATE TABLE IF NOT EXISTS country (
    index bigserial PRIMARY KEY,
    continent_index bigint REFERENCES continent(index),
    uuid uuid NOT NULL UNIQUE,
    name text,
    created timestamp DEFAULT NOW(),
    updated timestamp,
    deleted_state smallint default 0
);

CREATE TABLE IF NOT EXISTS city (
    index bigserial PRIMARY KEY,
    continent_index bigint REFERENCES continent(index),
    country_uuid uuid NOT NULL,
    uuid uuid NOT NULL UNIQUE,
    name text,
    created timestamp DEFAULT NOW(),
    updated timestamp,
    deleted_state smallint default 0
);
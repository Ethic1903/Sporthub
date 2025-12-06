CREATE TABLE IF NOT EXISTS facilities (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    city TEXT NOT NULL,
    address TEXT NOT NULL,
    type TEXT NOT NULL,
    amenities JSONB NOT NULL DEFAULT '[]'::jsonb,
    description TEXT NOT NULL DEFAULT '',
    maintenance JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS facilities_city_idx ON facilities (LOWER(city));
CREATE INDEX IF NOT EXISTS facilities_type_idx ON facilities (LOWER(type));

CREATE TABLE IF NOT EXISTS facility_availability (
    id UUID PRIMARY KEY,
    facility_id UUID NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    CHECK (end_at > start_at)
);

CREATE INDEX IF NOT EXISTS facility_availability_facility_idx
    ON facility_availability (facility_id, start_at);

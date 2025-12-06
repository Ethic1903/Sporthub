CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    facility_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    participants INT NOT NULL DEFAULT 0,
    price DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS bookings_facility_idx
    ON bookings (facility_id, starts_at);

CREATE INDEX IF NOT EXISTS bookings_user_idx
    ON bookings (user_id);

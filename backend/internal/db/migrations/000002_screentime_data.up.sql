CREATE TYPE data_status AS ENUM ('raw', 'validated', 'anonymized', 'listed', 'sold', 'withdrawn');

CREATE TABLE data_batches (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    record_count        INTEGER NOT NULL DEFAULT 0,
    date_range_start    TIMESTAMPTZ,
    date_range_end      TIMESTAMPTZ,
    status              data_status NOT NULL DEFAULT 'raw',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE screentime_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    batch_id        UUID REFERENCES data_batches(id) ON DELETE SET NULL,
    app_name        TEXT NOT NULL,
    app_category    TEXT,
    duration_secs   INTEGER NOT NULL CHECK (duration_secs > 0),
    started_at      TIMESTAMPTZ NOT NULL,
    ended_at        TIMESTAMPTZ NOT NULL,
    device_type     TEXT,
    os              TEXT,
    status          data_status NOT NULL DEFAULT 'raw',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_screentime_user     ON screentime_records(user_id);
CREATE INDEX idx_screentime_batch    ON screentime_records(batch_id);
CREATE INDEX idx_screentime_category ON screentime_records(app_category);
CREATE INDEX idx_screentime_started  ON screentime_records(started_at);

CREATE TABLE sharing_preferences (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    app_category        TEXT NOT NULL,
    share_app_name      BOOLEAN NOT NULL DEFAULT FALSE,
    share_duration      BOOLEAN NOT NULL DEFAULT TRUE,
    share_time_of_day   BOOLEAN NOT NULL DEFAULT TRUE,
    share_device        BOOLEAN NOT NULL DEFAULT TRUE,
    min_price_credits   BIGINT NOT NULL DEFAULT 100,
    UNIQUE(user_id, app_category)
);

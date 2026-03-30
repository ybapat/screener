CREATE TYPE dataset_status AS ENUM ('draft', 'active', 'paused', 'exhausted', 'withdrawn');
CREATE TYPE purchase_status AS ENUM ('pending', 'completed', 'refunded', 'failed');

CREATE TABLE datasets (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title                   TEXT NOT NULL,
    description             TEXT,
    category_filter         TEXT[],
    contributor_count       INTEGER NOT NULL DEFAULT 0,
    record_count            INTEGER NOT NULL DEFAULT 0,
    date_range_start        TIMESTAMPTZ,
    date_range_end          TIMESTAMPTZ,
    k_anonymity_k           INTEGER NOT NULL DEFAULT 5,
    epsilon_per_query       DOUBLE PRECISION NOT NULL,
    noise_mechanism         TEXT NOT NULL DEFAULT 'laplace',
    base_price_credits      BIGINT NOT NULL,
    current_price_credits   BIGINT NOT NULL,
    age_ranges              TEXT[],
    countries               TEXT[],
    status                  dataset_status NOT NULL DEFAULT 'draft',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_datasets_status   ON datasets(status);
CREATE INDEX idx_datasets_category ON datasets USING GIN(category_filter);

CREATE TABLE dataset_contributors (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_id          UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    epsilon_charged     DOUBLE PRECISION NOT NULL,
    records_included    INTEGER NOT NULL,
    earning_credits     BIGINT NOT NULL DEFAULT 0,
    UNIQUE(dataset_id, user_id)
);

CREATE TABLE dataset_samples (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_id              UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    app_category            TEXT NOT NULL,
    duration_range          TEXT NOT NULL,
    time_of_day             TEXT NOT NULL,
    device_type             TEXT,
    contributor_age_range   TEXT,
    contributor_country     TEXT
);

CREATE TABLE purchases (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id        UUID NOT NULL REFERENCES users(id),
    dataset_id      UUID NOT NULL REFERENCES datasets(id),
    price_credits   BIGINT NOT NULL,
    status          purchase_status NOT NULL DEFAULT 'pending',
    download_url    TEXT,
    download_count  INTEGER NOT NULL DEFAULT 0,
    purchased_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_purchases_buyer   ON purchases(buyer_id);
CREATE INDEX idx_purchases_dataset ON purchases(dataset_id);

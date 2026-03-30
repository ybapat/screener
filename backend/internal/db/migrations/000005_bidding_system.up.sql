CREATE TYPE bid_status AS ENUM ('active', 'accepted', 'rejected', 'expired', 'cancelled');

CREATE TABLE data_segments (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id            UUID NOT NULL REFERENCES users(id),
    app_categories      TEXT[] NOT NULL,
    date_range_start    TIMESTAMPTZ,
    date_range_end      TIMESTAMPTZ,
    age_ranges          TEXT[],
    countries           TEXT[],
    device_types        TEXT[],
    min_contributors    INTEGER NOT NULL DEFAULT 10,
    min_records         INTEGER NOT NULL DEFAULT 100,
    desired_k_anonymity INTEGER NOT NULL DEFAULT 5,
    max_epsilon         DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bids (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    segment_id      UUID NOT NULL REFERENCES data_segments(id) ON DELETE CASCADE,
    buyer_id        UUID NOT NULL REFERENCES users(id),
    bid_credits     BIGINT NOT NULL,
    status          bid_status NOT NULL DEFAULT 'active',
    expires_at      TIMESTAMPTZ NOT NULL,
    dataset_id      UUID REFERENCES datasets(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bids_segment ON bids(segment_id);
CREATE INDEX idx_bids_buyer   ON bids(buyer_id);
CREATE INDEX idx_bids_status  ON bids(status);

CREATE TABLE price_history (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    dataset_id      UUID REFERENCES datasets(id),
    app_category    TEXT NOT NULL,
    price_credits   BIGINT NOT NULL,
    demand_score    DOUBLE PRECISION NOT NULL,
    rarity_score    DOUBLE PRECISION NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credit_transactions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id),
    amount          BIGINT NOT NULL,
    balance_after   BIGINT NOT NULL,
    tx_type         TEXT NOT NULL,
    reference_id    UUID,
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_credit_tx_user ON credit_transactions(user_id);

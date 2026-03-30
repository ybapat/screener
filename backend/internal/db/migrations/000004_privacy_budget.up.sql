CREATE TYPE budget_event_type AS ENUM ('dataset_sale', 'query_response', 'sample_generation', 'budget_refund');

CREATE TABLE epsilon_ledger (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type          budget_event_type NOT NULL,
    epsilon_spent       DOUBLE PRECISION NOT NULL,
    epsilon_remaining   DOUBLE PRECISION NOT NULL,
    dataset_id          UUID REFERENCES datasets(id),
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_epsilon_ledger_user    ON epsilon_ledger(user_id);
CREATE INDEX idx_epsilon_ledger_dataset ON epsilon_ledger(dataset_id);
CREATE INDEX idx_epsilon_ledger_created ON epsilon_ledger(created_at);

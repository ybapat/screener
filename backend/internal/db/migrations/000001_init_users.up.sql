CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('seller', 'buyer', 'admin');

CREATE TABLE users (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email                   TEXT NOT NULL UNIQUE,
    password_hash           TEXT NOT NULL,
    display_name            TEXT NOT NULL,
    role                    user_role NOT NULL DEFAULT 'seller',
    age_range               TEXT,
    country                 TEXT,
    timezone                TEXT,
    credit_balance          BIGINT NOT NULL DEFAULT 0,
    global_epsilon_budget   DOUBLE PRECISION NOT NULL DEFAULT 10.0,
    epsilon_spent           DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);

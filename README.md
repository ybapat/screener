# Screener

A privacy-preserving screen time data marketplace built on Solana. Users sell their anonymized app usage data to researchers and buyers, with mathematical privacy guarantees ensuring no individual can be re-identified. Payments are handled via SOL on Solana devnet through a custom Anchor escrow program.

The system applies **k-anonymity** (grouping records so no individual stands out) and **differential privacy** (adding calibrated Laplace noise to aggregations) before any data leaves the platform. Each user has an **epsilon budget** — a hard cap on how much information about them can ever be extracted — tracked atomically with a full audit ledger.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Next.js    │────▶│   Go API    │────▶│  PostgreSQL  │     │ Solana Devnet│
│   Frontend   │     │  (Chi)      │────▶│    Redis     │     │              │
│  :3000       │     │  :8080      │     │  :5432/:6379 │     │ Escrow Prog. │
└──────┬──────┘     └──────┬──────┘     └──────────────┘     └──────┬───────┘
       │                   │                                        │
       │            ┌──────┴──────┐                                 │
       │            │   Privacy   │                                 │
       │            │   Engine    │                                 │
       │            ├─────────────┤                                 │
       │            │ Generalizer │  app → category, time → bucket  │
       │            │ k-Anonymity │  suppress groups with < k users │
       │            │ DP Noise    │  Laplace mechanism (ε-DP)       │
       │            │ Budget Mgr  │  atomic ε debit + audit ledger  │
       │            └─────────────┘                                 │
       │                                                            │
       └──── Phantom/Solflare wallet ── deposit/release/refund ─────┘
```

**Stack:** Go (Chi) / Next.js 15 / React 19 / Tailwind v4 / PostgreSQL 16 / Redis 7 / Solana (Anchor/Rust) — all Dockerized, zero global installs.

## Quick Start

```bash
git clone https://github.com/ybapat/screener.git
cd screener
make up
```

That's it. Docker Compose builds and starts everything:
- **Frontend** at [http://localhost:3000](http://localhost:3000)
- **API** at [http://localhost:8080](http://localhost:8080)
- **PostgreSQL** on port 5432, **Redis** on port 6379

Migrations run automatically on startup.

### Seed Data

To populate the database with test users and 30 days of realistic screen time data:

```bash
make seed
```

This creates 10 seller accounts, 3 buyer accounts, and uploads screen time data for each seller.

## How It Works

### 1. Sellers Upload Data

Sellers submit batches of screen time records (app name, duration, timestamps, device type). The API validates each record (duration bounds, timestamp sanity, rate limiting) and stores the raw data.

### 2. Anonymization Pipeline

When a dataset is assembled, the pipeline runs four stages:

| Stage | What it does | Why |
|-------|-------------|-----|
| **Generalize** | Maps apps to categories (e.g., Instagram → Social), timestamps to time-of-day buckets, durations to ranges | Reduces quasi-identifier uniqueness |
| **k-Anonymize** | Groups records by quasi-identifiers, suppresses groups with fewer than *k* distinct contributors | Prevents singling out individuals |
| **DP Noise** | Adds Laplace noise to count, mean, and sum per group (ε split three ways) | Bounds information leakage mathematically |
| **Budget Debit** | Atomically decrements each contributor's epsilon budget and writes to the audit ledger | Enforces lifetime privacy limits via sequential composition |

### 3. Buyers Browse and Purchase

Buyers browse anonymized datasets with metadata (contributor count, categories, date range, k value, epsilon used). They can preview sample data before purchasing. Credits transfer from buyer to sellers proportionally.

### 4. Dynamic Pricing

Dataset prices are computed as:

```
price = base × rarity × demand × quality
```

- **Rarity:** inverse of supply (log-scaled)
- **Demand:** number of active bids for similar data
- **Quality:** ratio of k-anonymity threshold to epsilon (higher k and lower ε = better privacy = higher quality)

### 5. Solana Payments

The platform supports SOL payments on Solana devnet via a custom **Anchor escrow program** (`contracts/escrow/`). Two payment flows:

**Credit Top-Up (Direct Transfer):**
Users connect a Phantom or Solflare wallet, then send SOL directly to the server wallet. The backend verifies the on-chain transfer and credits their account at a configurable exchange rate.

**Dataset Purchase (Escrow):**
1. Buyer clicks "Pay with SOL" — frontend calls `POST /solana/purchase/init`
2. Backend returns escrow PDA address, amount in lamports, program ID
3. Frontend builds a `deposit` instruction targeting the escrow program; Phantom signs and submits
4. Frontend sends the tx signature to `POST /solana/purchase/confirm`
5. Backend verifies the deposit on-chain (escrow PDA now holds the SOL)
6. Backend creates the purchase, then calls `release` on the escrow program for each seller with a linked wallet (server keypair is the authority)
7. Sellers without wallets receive mock credits instead

#### Escrow Program

The Anchor program (`contracts/escrow/programs/escrow/src/lib.rs`) has three instructions:

| Instruction | Signer | What it does |
|-------------|--------|-------------|
| `deposit` | Buyer | Transfers SOL into a PDA vault; creates `EscrowState` account |
| `release` | Authority (server) | Sends SOL from vault to a seller; can be called per-seller until drained |
| `refund` | Authority (server) | Returns remaining SOL to the buyer if the purchase is cancelled |

PDA seeds: `["escrow", buyer_pubkey, dataset_id]` for state, `["vault", buyer_pubkey, dataset_id]` for the vault.

The entire Solana subsystem is **optional** — if `SOLANA_RPC_URL` is unset, the platform runs with mock credits only.

## API Routes

### Public
| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/register` | Create account (seller/buyer) |
| POST | `/auth/login` | Get access + refresh tokens |
| POST | `/auth/refresh` | Rotate refresh token |
| GET | `/api/v1/marketplace/datasets` | Browse available datasets |
| GET | `/api/v1/marketplace/datasets/:id` | Dataset detail |
| GET | `/api/v1/marketplace/datasets/:id/samples` | Preview anonymized samples |

### Seller (authenticated)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/data/upload` | Upload screen time batch |
| GET | `/api/v1/data/batches` | List upload batches |
| DELETE | `/api/v1/data/batches/:id` | Withdraw a batch |
| GET | `/api/v1/privacy/budget` | Check remaining epsilon |
| GET | `/api/v1/privacy/ledger` | Audit trail of budget usage |
| GET | `/api/v1/dashboard/seller` | Earnings, batches, budget overview |

### Buyer (authenticated)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/marketplace/datasets/:id/purchase` | Buy a dataset |
| GET | `/api/v1/buyer/purchases` | Purchase history |
| POST | `/api/v1/marketplace/segments` | Define a data segment |
| POST | `/api/v1/marketplace/segments/:id/bids` | Place a bid |
| GET | `/api/v1/marketplace/bids` | Active bids |
| POST | `/api/v1/credits/topup` | Add credits (mock) |
| GET | `/api/v1/dashboard/buyer` | Spend history, active bids |

### Solana (authenticated)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/solana/info` | Server wallet, balance, exchange rate, program ID |
| POST | `/api/v1/solana/wallet/link` | Link Phantom/Solflare wallet (Ed25519 sig verify) |
| GET | `/api/v1/solana/transactions` | SOL transaction history |
| POST | `/api/v1/solana/topup/init` | Get server wallet address for direct transfer |
| POST | `/api/v1/solana/topup/confirm` | Verify on-chain transfer, credit account |
| POST | `/api/v1/solana/purchase/init` | Get escrow PDA + amount for deposit |
| POST | `/api/v1/solana/purchase/confirm` | Verify escrow deposit, release to sellers |

### Admin
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/datasets/assemble` | Trigger anonymization pipeline |

## Project Structure

```
screener/
├── docker-compose.yml
├── Makefile
├── contracts/
│   ├── escrow/
│   │   ├── Anchor.toml              # Devnet config
│   │   ├── Cargo.toml               # Workspace manifest
│   │   └── programs/escrow/
│   │       └── src/lib.rs           # Anchor escrow program (deposit/release/refund)
│   ├── Dockerfile.anchor            # Rust + Anchor build environment
│   └── deploy.sh                    # Build + deploy to devnet
├── backend/
│   ├── cmd/server/main.go           # Entrypoint — wires everything
│   ├── internal/
│   │   ├── config/                   # Env-based configuration
│   │   ├── db/                       # Postgres/Redis setup + SQL migrations
│   │   ├── models/                   # Domain models (User, ScreenTime, Dataset, Solana)
│   │   ├── repository/               # Database access layer (pgx)
│   │   ├── service/                  # Business logic (auth, ingestion, anonymization, solana)
│   │   ├── handler/                  # HTTP handlers
│   │   ├── middleware/               # Auth, RBAC, CORS, rate limiting, logging
│   │   ├── router/                   # Chi route assembly
│   │   ├── privacy/                  # k-anonymity, differential privacy, budget tracker
│   │   ├── pricing/                  # Dynamic pricing engine
│   │   └── solana/                   # Solana RPC client, keypair mgmt, escrow instructions
│   └── pkg/                          # Shared utilities (errors, validation, response)
├── frontend/
│   └── src/
│       ├── app/                      # Next.js App Router pages (incl. /solana)
│       ├── components/
│       │   └── solana/               # WalletButton, SolTopup
│       ├── lib/                      # API client, auth helpers, escrow instruction builder
│       ├── contexts/                 # Auth + Solana wallet contexts
│       └── types/                    # TypeScript types
└── scripts/
    └── seed.go                       # Test data generator
```

## Database Schema

6 migrations, applied automatically on `docker compose up`:

1. **Users & Auth** — `users` (with roles, epsilon budget, credit balance), `refresh_tokens`
2. **Screen Time** — `screentime_records`, `data_batches`, `sharing_preferences`
3. **Datasets & Marketplace** — `datasets`, `dataset_contributors`, `dataset_samples`, `purchases`
4. **Privacy Ledger** — `epsilon_ledger` (immutable audit log of all budget expenditures)
5. **Bidding** — `data_segments`, `bids`, `price_history`, `credit_transactions`
6. **Solana** — `sol_transactions` (on-chain tx log), `sol_escrows` (escrow state), `sol_config` (exchange rate), `solana_wallet` column on `users`

## Make Commands

| Command | Description |
|---------|-------------|
| `make up` | Build and start all services |
| `make down` | Stop all services |
| `make build` | Rebuild and restart |
| `make logs` | Tail all logs |
| `make logs-backend` | Tail backend logs only |
| `make seed` | Populate database with test data |
| `make migrate-up` | Run migrations manually |
| `make reset` | Destroy volumes and rebuild from scratch |

## Privacy Guarantees

- **k-Anonymity (k=5 default):** Every record in a released dataset belongs to a group of at least *k* distinct contributors. Groups that don't meet the threshold are suppressed entirely.
- **ε-Differential Privacy:** Aggregation queries use the Laplace mechanism. The total epsilon per dataset assembly is split across count, mean, and sum queries. Noise scale = sensitivity / ε.
- **Budget Enforcement:** Each user has a lifetime epsilon budget (default 10.0). The budget is debited atomically in a PostgreSQL transaction — if the budget would go negative, the operation fails and no data is included. Every debit is recorded in an append-only ledger.
- **Sequential Composition:** Total privacy loss is the sum of epsilons across all datasets a user contributes to. The budget tracker enforces this automatically.

## Tech Decisions

| Decision | Rationale |
|----------|-----------|
| Chi over Fiber | Chi uses stdlib `net/http`, compatible with all Go middleware |
| k-anonymity from scratch | No mature Go library; the algorithm is ~200 lines of group-by + suppress |
| Pure-Go Laplace mechanism | Avoids CGo dependency on Google's DP library while providing formal guarantees |
| Credits as int64 cents | No float precision issues, maps to lamports cleanly |
| Datasets stored as JSONL files | Not in Postgres — they can be large. Metadata in DB, data on disk |
| Atomic epsilon debit | PostgreSQL transaction with conditional UPDATE prevents race conditions |
| Anchor escrow program | Demonstrates real Solana program dev: PDAs, CPIs, account validation, authority gating |
| Server keypair as authority | Backend controls release/refund — business logic stays server-side, program stays simple |
| `gagliardetto/solana-go` | Standard Go Solana library, pure Go, no CGo |
| Direct transfer for top-ups | Escrow is overkill for credit purchases — simple `SystemProgram.transfer` is cleaner |
| Optional Solana subsystem | If `SOLANA_RPC_URL` is unset, the entire layer is disabled; mock credits still work |
| Idempotent tx confirms | `tx_signature UNIQUE` constraint prevents double-crediting on retries |

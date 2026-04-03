# Screener

A privacy-preserving screen time data marketplace. Users sell their anonymized app usage data to researchers and buyers, with mathematical privacy guarantees ensuring no individual can be re-identified.

The system applies **k-anonymity** (grouping records so no individual stands out) and **differential privacy** (adding calibrated Laplace noise to aggregations) before any data leaves the platform. Each user has an **epsilon budget** — a hard cap on how much information about them can ever be extracted — tracked atomically with a full audit ledger.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│   Next.js    │────▶│   Go API    │────▶│  PostgreSQL  │
│   Frontend   │     │  (Chi)      │────▶│    Redis     │
│  :3000       │     │  :8080      │     │  :5432/:6379 │
└─────────────┘     └──────┬──────┘     └──────────────┘
                           │
                    ┌──────┴──────┐
                    │   Privacy   │
                    │   Engine    │
                    ├─────────────┤
                    │ Generalizer │  app → category, time → bucket
                    │ k-Anonymity │  suppress groups with < k users
                    │ DP Noise    │  Laplace mechanism (ε-DP)
                    │ Budget Mgr  │  atomic ε debit + audit ledger
                    └─────────────┘
```

**Stack:** Go (Chi) / Next.js 15 / React 19 / Tailwind v4 / PostgreSQL 16 / Redis 7 — all Dockerized, zero global installs.

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

### Admin
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/datasets/assemble` | Trigger anonymization pipeline |

## Project Structure

```
screener/
├── docker-compose.yml
├── Makefile
├── backend/
│   ├── cmd/server/main.go          # Entrypoint — wires everything
│   ├── internal/
│   │   ├── config/                  # Env-based configuration
│   │   ├── db/                      # Postgres/Redis setup + SQL migrations
│   │   ├── models/                  # Domain models (User, ScreenTime, Dataset, etc.)
│   │   ├── repository/              # Database access layer (pgx)
│   │   ├── service/                 # Business logic (auth, ingestion, anonymization, credits)
│   │   ├── handler/                 # HTTP handlers
│   │   ├── middleware/              # Auth, RBAC, CORS, rate limiting, logging
│   │   ├── router/                  # Chi route assembly
│   │   ├── privacy/                 # k-anonymity, differential privacy, budget tracker
│   │   └── pricing/                 # Dynamic pricing engine
│   └── pkg/                         # Shared utilities (errors, validation, response)
├── frontend/
│   └── src/
│       ├── app/                     # Next.js App Router pages
│       ├── components/              # UI components
│       ├── lib/                     # API client, auth helpers
│       ├── contexts/                # Auth context
│       └── types/                   # TypeScript types
└── scripts/
    └── seed.go                      # Test data generator
```

## Database Schema

5 migrations, applied automatically on `docker compose up`:

1. **Users & Auth** — `users` (with roles, epsilon budget, credit balance), `refresh_tokens`
2. **Screen Time** — `screentime_records`, `data_batches`, `sharing_preferences`
3. **Datasets & Marketplace** — `datasets`, `dataset_contributors`, `dataset_samples`, `purchases`
4. **Privacy Ledger** — `epsilon_ledger` (immutable audit log of all budget expenditures)
5. **Bidding** — `data_segments`, `bids`, `price_history`, `credit_transactions`

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
| Credits as int64 cents | No float precision issues, maps to Stripe cents and ETH wei later |
| Datasets stored as JSONL files | Not in Postgres — they can be large. Metadata in DB, data on disk |
| Atomic epsilon debit | PostgreSQL transaction with conditional UPDATE prevents race conditions |

# HashNut Demo Shop (Go)

A demo merchant application demonstrating how to integrate with HashNut Payment API (V4) using the [payment-sdk-go](../payment-sdk-go). Supports multi-chain payment (ERC20 / TRC20).

## Tech Stack

- **Backend**: Go + Gin
- **Database**: PostgreSQL
- **Payment**: HashNut Go SDK (V4)
- **Frontend**: Shared React + TypeScript project (see [demo-web](../demo-web))

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Node.js 18+ (for frontend)
- A HashNut merchant account with API Key
- ngrok (for local development callback)

## Quick Start

### 1. Create Database

```bash
psql -U postgres -c "CREATE DATABASE demo_shop;"
psql -U postgres -d demo_shop -f migrate.sql
```

### 2. Configure

Edit `migrate.sql` seed data before running:

- **t_coin_info**: Supported chain + coin combinations
- **t_hashnut_api_key**: Your splitter address + API credentials per chain

```sql
-- Example: configure ETH and Tron USDT
INSERT INTO t_hashnut_api_key (chain_code, splitter, access_key_id, secret_key) VALUES
    ('erc20', '0x...your-eth-splitter...', 'your-access-key-id', 'your-secret-key'),
    ('trc20', 'T...your-tron-splitter',    'your-access-key-id', 'your-secret-key');
```

Edit `etc/application.yaml`:

```yaml
server:
  port: 1800

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "demo_shop"
  sslmode: "disable"

hashnut:
  testMode: false    # true = testnet, false = production
  baseURL: ""        # Leave empty for default; set custom URL if needed
```

### 3. Run Backend

```bash
go run main.go
```

The server starts on `http://localhost:1800`.

### 4. Run Frontend

```bash
cd ../demo-web
npm install
npm run dev
```

The frontend starts on `http://localhost:5173` and proxies `/api` to `localhost:1800`.

Open `http://localhost:5173` in your browser.

## Local Development with ngrok

For HashNut backend to send payment notifications to your local machine, use ngrok:

### 1. Start ngrok

```bash
ngrok http 1800
```

This gives you a public URL like `https://xxxxx.ngrok-free.dev`.

### 2. Configure HashNut API Key

In your HashNut merchant dashboard, set:

| Field | Local Development | Production |
|-------|-------------------|------------|
| notifyURL | `https://xxxxx.ngrok-free.dev/api/notify` | `https://your-domain.com/api/notify` |
| callbackURL | `http://localhost:5173/payment-result` | `https://your-domain.com/payment-result` |

- **notifyURL**: Backend webhook — must be publicly accessible, use ngrok URL
- **callbackURL**: Frontend redirect after payment — use `localhost:5173` for local dev (browser redirect, no public access needed)

### 3. Payment Flow (Local Dev)

```
Browser (localhost:5173)
  → Click "Pay with Crypto"
  → POST /api/orders (proxied to localhost:1800)
  → Redirect to HashNut payment page (defi.hashnut.io/pay)
  → User pays on-chain
  → HashNut backend sends notification to ngrok URL → localhost:1800/api/notify
  → HashNut frontend redirects to http://localhost:5173/payment-result?state=4&...
  → Frontend shows payment success
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/products` | List all products |
| GET | `/api/chains` | List supported chains + coins (from DB) |
| POST | `/api/orders` | Create order `{productId, chainCode, coinCode}` |
| GET | `/api/orders/:id` | Query order status |
| POST | `/api/orders/:id/confirm` | Submit payment tx hash `{payTxId}` |
| POST | `/api/notify` | HashNut payment result webhook |

### Create Order

```bash
curl -X POST http://localhost:1800/api/orders \
  -H "Content-Type: application/json" \
  -d '{"productId": 1, "chainCode": "erc20", "coinCode": "usdt"}'
```

## Database Schema

| Table | Description |
|-------|-------------|
| `t_coin_info` | Supported chain + coin configurations |
| `t_hashnut_api_key` | Splitter address + API credentials per chain |
| `products` | Demo products (price only, no chain/coin binding) |
| `orders` | Orders with user-selected chain + coin |

## Project Structure

```
demo-go/
├── main.go                  # Entry point, Gin router setup
├── etc/application.yaml     # Runtime configuration
├── migrate.sql              # Database schema + seed data
└── internal/
    ├── config/config.go     # Config loading (testMode + baseURL only)
    ├── model/model.go       # Product, Order, CoinInfo, ApiKeyInfo
    ├── store/store.go       # PostgreSQL operations
    ├── handler/handler.go   # API handlers (SDK client cached per secretKey)
    └── notify/notify.go     # HashNut webhook handler
```

## License

MIT

# Namex - Bot & Anti-DDoS Protection Platform

A high-performance, Cloudflare-like platform with Bot Protection, Rate Limiting, and DDoS Mitigation at both **HTTP (Layer 7)** and **DNS** layers.

## Features

### Core
- 🛡️ **Bot Protection**: IP and User-Agent based blocking
- ⏱️ **Rate Limiting**: Redis-based token bucket (HTTP + DNS)
- 🚫 **Anti-DDoS**: Request throttling and IP blocking
- 📊 **Analytics**: High-performance ClickHouse storage
- 🌐 **DNS Server**: Authoritative DNS with A, CNAME, TXT, MX records
- 🔀 **Reverse Proxy**: Layer 7 HTTP proxy with security filtering

### Tech Stack
- **Language**: Go 1.20+
- **Router**: Chi
- **Database**: PostgreSQL (configuration), ClickHouse (analytics)
- **Cache**: Redis (rate limiting)
- **DNS**: github.com/miekg/dns

## Quick Start

1. **Start infrastructure:**
   ```bash
   docker-compose up -d
   ```

2. **Apply migrations:**
   ```bash
   docker exec -i goflare_postgres psql -U user -d goflare < migrations/001_create_domains_table.sql
   docker exec -i goflare_postgres psql -U user -d goflare < migrations/002_create_bot_rules_table.sql
   docker exec -i goflare_postgres psql -U user -d goflare < migrations/004_create_dns_records_table.sql
   docker exec -i goflare_clickhouse clickhouse-client --password "" --database goflare_analytics < migrations/003_create_requests_table_clickhouse.sql
   ```

3. **Run server:**
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints

### Domain Management
- `POST /api/v1/domains` - Add a domain
- `GET /api/v1/domains` - List domains

### Bot Rules
- `POST /api/v1/bot/rules` - Add bot rule
- `GET /api/v1/bot/rules` - List rules

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Client                           │
└─────┬───────────────────────────────────┬───────────┘
      │                                   │
      │ DNS Query (8053)                  │ HTTP Request (8080)
      │                                   │
┌─────▼──────────────┐            ┌──────▼─────────────┐
│   DNS Server       │            │   HTTP Proxy       │
│  (Authoritative)   │            │  (Reverse Proxy)   │
└─────┬──────────────┘            └──────┬─────────────┘
      │                                   │
      └───────────┬───────────────────────┘
                  │
            ┌─────▼──────────┐
            │  Engine Core   │
            │ • Bot Check    │
            │ • Rate Limit   │
            │ • Analytics    │
            └────────────────┘
```

## Configuration

See `.env.example` for available environment variables.

## License

MIT

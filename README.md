# Microgame Bot

A Telegram bot for playing mini-games with betting system built in Go.

## Overview

Microgame Bot is a competitive multiplayer Telegram bot that allows users to play quick mini-games against each other with optional betting. The bot uses Telegram's inline mode for seamless game interaction directly in chats.

## Features

### Games

- **Rock Paper Scissors (RPS)** - Classic hand game for two players with best-of-N series support
- **Tic Tac Toe (TTT)** - Strategic board game with turn-based gameplay

### Core Features

- **Inline Game Selector** - Start games in any chat using inline mode (`@bot_name`)
- **Betting System** - Place bets on game outcomes with automatic payout
- **User Profiles** - Track your wins, losses, balance, and statistics
- **Daily Bonus** - Claim daily rewards to boost your balance
- **Series Matches** - Play best-of-N game series with configurable rounds
- **Real-time Updates** - Live game state updates via inline keyboard buttons

### Technical Features

- **Distributed Locking** - Prevents race conditions in concurrent gameplay
- **Task Queue System** - Handles async operations (payouts, timeouts, cleanups)
- **Job Scheduler** - Automated maintenance tasks with cron expressions
- **Unit of Work Pattern** - Ensures transactional consistency across repositories
- **Session Management** - Persistent game sessions with state recovery
- **Webhook & Long Polling** - Flexible deployment options
- **Health Check** - Monitor system health via `/health` endpoint

## Architecture

Built with clean architecture principles:
- Domain-driven design with isolated game logic
- Repository pattern for data persistence
- Middleware chain for request processing
- FSM (Finite State Machine) for complex workflows
- GORM for database operations (PostgreSQL)

## Tech Stack

- **Language**: Go 1.25.5
- **Bot Framework**: [Telego](https://github.com/mymmrac/telego)
- **Database**: PostgreSQL (via GORM)
- **Queue**: PostgreSQL (via GORM)
- **Scheduler**: PostgreSQL (via GORM)

## Deployment

### Docker Compose

1. Create `.env` file with required variables:

```env
TELEGRAM__TOKEN=your_bot_token_here
TELEGRAM__WEBHOOK_URL=https://yourdomain.com
POSTGRES__URL=postgres://app:app@postgres:5432/app
```

See `.env.example` for complete list of available options.

2. Run with production compose file:

```bash
docker compose -f docker-compose.production.yaml up -d
```

The bot will be available on port 8080 for webhook connections.

## Health Check

The bot exposes a `/health` endpoint for monitoring system health. The endpoint checks:

- **Database** - PostgreSQL connection and pool status
- **Queue** - Task queue health and stuck tasks detection
- **Scheduler** - Active cron jobs status

### Health Check Response

```json
{
  "status": "ok",
  "timestamp": "2026-01-05T10:30:00Z",
  "components": {
    "database": {
      "status": "ok",
      "latency": "2.5ms"
    },
    "queue": {
      "status": "ok",
      "latency": "1.8ms"
    },
    "scheduler": {
      "status": "ok",
      "latency": "1.2ms"
    }
  }
}
```

### Status Codes

- `200 OK` - All systems operational or degraded
- `503 Service Unavailable` - Critical system failure

### Component Statuses

- `ok` - Component is healthy
- `degraded` - Component is working but with issues (e.g., high stuck tasks)
- `down` - Component is not responding

### Usage

Access the health check endpoint at: `http://your-host:8080/health`

For Kubernetes deployments, liveness and readiness probes are pre-configured in `manifests/deployment.yaml`.

**Note**: Health check is available regardless of whether the bot uses webhook or long polling mode.

### Kubernetes

1. Update `manifests/env.yaml` with your configuration:
   - Set `TELEGRAM__TOKEN` with your bot token
   - Set `TELEGRAM__WEBHOOK_URL` with your domain
   - Configure `POSTGRES__URL` to point to your database

2. Apply manifests:

```bash
kubectl apply -f manifests/
```

Required manifests:
- `env.yaml` - Secret with environment variables
- `deployment.yaml` - Bot deployment with health checks
- `service.yaml` - ClusterIP service
- `ingress.yaml` - Ingress for webhook routing

### Required Environment Variables

- `TELEGRAM__TOKEN` - Telegram bot token from [@BotFather](https://t.me/BotFather)
- `TELEGRAM__WEBHOOK_URL` - Public HTTPS URL for webhook
- `POSTGRES__URL` - PostgreSQL connection string

See `manifests/env.yaml` and `.env.example` for complete list of available options.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

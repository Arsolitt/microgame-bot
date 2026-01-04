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

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

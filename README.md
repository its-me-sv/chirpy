# Chirpy

A small Twitter-style backend written in Go. Users can sign up, log in, and post short "chirps" (140 characters max). Backed by PostgreSQL, with JWT-based auth and a webhook for handling paid upgrades.

Built while working through the [boot.dev](https://boot.dev) Go course.

## Table of Contents

- [About](#about)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Notes & Warnings](#notes--warnings)

## About

Chirpy is a REST API, no frontend beyond a static welcome page. It covers the basics of a social app backend:

- User accounts with email/password auth
- Access tokens (JWT) + refresh tokens
- Creating, listing, filtering, sorting, and deleting chirps
- A profanity filter on chirp bodies
- A "Chirpy Red" paid upgrade flow, triggered by an external webhook
- Admin endpoints for request metrics and resetting dev data

## Tech Stack

| Layer         | Choice                                      |
|---------------|----------------------------------------------|
| Language      | Go 1.26 (standard library `net/http`, no web framework) |
| Database      | PostgreSQL                                    |
| Queries       | [sqlc](https://sqlc.dev) (generated into `internal/database`) |
| Migrations    | [goose](https://github.com/pressly/goose)-style SQL files in `sql/schema` |
| Auth          | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) for access tokens, random hex tokens for refresh |
| Password hashing | [argon2id](https://github.com/alexedwards/argon2id) |
| Env config    | [godotenv](https://github.com/joho/godotenv) |
| IDs           | [google/uuid](https://github.com/google/uuid) |

## Getting Started

### Prerequisites

- Go 1.26+
- A running PostgreSQL instance
- [goose CLI](https://github.com/pressly/goose) for running migrations

### Setup

1. Clone the repo and create a database:

   ```bash
   createdb chirpy
   ```

2. Create a `.env` file in the project root:

   ```
   DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
   PLATFORM=dev
   JWT_SECRET=<a long random string>
   POLKA_KEY=<key shared with the Polka webhook sender>
   ```

3. Run migrations:

   ```bash
   goose -dir sql/schema postgres "$DB_URL" up
   ```

4. Start the server:

   ```bash
   go run .
   ```

   The server listens on `:8080`.

## API Reference

| Method | Path                    | Auth              | Description                              |
|--------|--------------------------|-------------------|-------------------------------------------|
| GET    | `/api/healthz`           | —                 | Health check                              |
| POST   | `/api/users`             | —                 | Create a user                             |
| PUT    | `/api/users`             | Bearer (access)   | Update the logged-in user's email/password |
| POST   | `/api/login`             | —                 | Log in, returns access + refresh tokens   |
| POST   | `/api/refresh`           | Bearer (refresh)  | Exchange a refresh token for a new access token |
| POST   | `/api/revoke`            | Bearer (refresh)  | Revoke a refresh token                    |
| POST   | `/api/chirps`            | Bearer (access)   | Create a chirp                            |
| GET    | `/api/chirps`            | —                 | List chirps (`?author_id=`, `?sort=asc\|desc`) |
| GET    | `/api/chirps/{chirpID}`  | —                 | Get a single chirp                        |
| DELETE | `/api/chirps/{chirpID}`  | Bearer (access)   | Delete your own chirp                     |
| POST   | `/api/polka/webhooks`    | ApiKey header     | Upgrade a user to Chirpy Red              |
| GET    | `/admin/metrics`         | —                 | View fileserver hit count (dev)           |
| POST   | `/admin/reset`           | —                 | Wipe users and reset hit count (dev only) |

## Notes & Warnings

- `/admin/reset` only works when `PLATFORM=dev`. It deletes all users (chirps cascade with them) — never point it at a production database.
- Chirp bodies over 140 characters are rejected.
- Three specific words are auto-censored in chirp bodies (see `getProfanceReplacedString` in [handler_chirps.go](handler_chirps.go)) — this is a toy filter, not real moderation.
- Access tokens expire after 1 hour; refresh tokens after 60 days.
- Passwords are hashed with argon2id, never stored or logged in plain text.
- `.env` is gitignored on purpose — never commit real secrets there.

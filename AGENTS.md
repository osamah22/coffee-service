# AGENTS.md

## Purpose

This project is a portfolio-grade coffee ordering system. The goal is to demonstrate clean architecture, pragmatic service boundaries, and minimal event-driven design without overengineering.

## Current Architecture

Currently implemented:

- Order Service (Go)
- Notification Service (Go)
- Frontend (React)
- SuperTokens (auth/session provider)
- RabbitMQ (event bus)
- PostgreSQL (primary runtime DB)
- SQLite (tests only)
- MailHog (local notification email capture)

Future architecture work:

- Gateway (Go)
- Dedicated Product Service (Go)
- Authentik OIDC/JWKS flow

## Service Boundaries

### Current Order Service

- Owns orders and line items.
- Currently also owns product/menu data for simplicity.
- Looks up product data server-side during checkout.
- Emits order events through an outbox dispatcher.

### Future Product Service

- Will own products once split out.
- Should have no knowledge of orders.

### Notification Service

- Consumes events only.
- Sends notifications through SMTP or logs in dry-run mode.
- Must not write to another service database.

### Future Gateway

- Will handle authentication/JWT validation.
- Will route requests to downstream services.
- Will inject user identity headers.

## Authentication

Current auth is handled by SuperTokens inside order-service.

- Auth routes are mounted at `/auth`.
- Role-aware middleware supports guest, user, barista, and admin flows.
- Admin/barista roles are derived from configured email lists.

Future auth target:

- Authentik OIDC.
- Gateway validates JWTs using JWKS.
- Downstream services trust gateway headers such as `X-User-Sub` and `X-User-Email`.

## Events

RabbitMQ topic exchange:

- `coffee.orders`

Event types:

- `order.created`
- `order.status_updated`

Rules:

- Events are facts, not commands.
- Events must be idempotent for consumers.
- Keep payloads stable and minimal.
- Do not add events unless another service actually needs the fact.

## Code Structure Rules

Each Go service should keep this basic shape:

```text
internal/
  models/
  services/
  handlers/
```

Guidelines:

- `models/`: DB models and validation.
- `services/`: business logic; GORM can be used directly.
- `handlers/`: HTTP layer only.
- Avoid repositories/interfaces unless there is a concrete need.

## Testing

- SQLite is used for tests.
- Each test must use an isolated DB.
- Focus tests on model validation, service behavior, event/outbox behavior, and basic DB operations.
- Notification service still needs coverage for email formatting and consumer ack/nack behavior.

## UI Guidelines

Frontend uses a retro/pixel aesthetic.

- Prefer pixelated assets.
- Keep styling minimal and consistent.
- Build usable app workflows over marketing pages.

## Design Philosophy

- Keep it simple but structured.
- Prefer clarity over cleverness.
- Build only what is needed.
- Make tradeoffs explicit.

## Future Improvements

- Gateway service.
- Product-service split.
- Authentik OIDC/JWKS migration.
- Pagination/filtering.
- Observability.
- CI/CD.
- Rate-limit hardening.

## Summary

This project is structured but intentionally not overengineered, event-driven where useful, and production-inspired without losing demo clarity.

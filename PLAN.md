# Coffee Service Plan

## Goal

Build a clean, working coffee ordering demo that shows practical service boundaries, role-aware workflows, and minimal event-driven behavior without overengineering.

## Current Status

- [x] Product model, service, handlers, tests, and seed data.
- [x] Order model, line items, status workflow, service tests, and SQLite test setup.
- [x] Server-side product lookup during checkout so clients do not send trusted prices.
- [x] Customer order history and staff queue endpoints.
- [x] SuperTokens auth/session integration with role-aware middleware.
- [x] React frontend for menu browsing, cart checkout, customer orders, and barista/admin queue.
- [x] Outbox records for `order.created` and `order.status_updated`.
- [x] RabbitMQ publisher/dispatcher.
- [x] Notification service RabbitMQ consumer.
- [x] SMTP and log email senders with MailHog local configuration.
- [x] Docker Compose for Postgres, RabbitMQ, SuperTokens, order-service, notification-service, MailHog, and frontend.
- [x] README updated to match the current implementation.
- [x] Notification tests for config, email formatting, sender error handling, and consumer ack/nack/idempotency paths.
- [x] Makefile for common checks.
- [x] Handler tests for checkout product lookup, customer order history, staff status transitions, and role access.
- [x] Docker Compose smoke test documented for API, frontend, and MailHog.
- [x] Architecture, API, event contract, runbook, and README diagrams documented.
- [x] README demo flow documented for a portfolio walkthrough.
- [x] Swagger decision recorded: keep Markdown docs for this portfolio demo unless generated OpenAPI becomes a concrete requirement.

## Next Work

No remaining current-scope feature work. Items below are future improvements, not incomplete demo features.

## Architecture Decisions

- Current auth is SuperTokens inside order-service. Authentik plus a Go gateway remains a future replacement, not the active implementation.
- Product and order code currently live in order-service. A standalone product-service remains future work.
- The outbox pattern is already implemented for order events because it keeps event publishing consistent with order writes.
- Notification service consumes events only and does not write to another service database.

## Future Improvements

- Separate gateway that validates JWTs, routes requests, and injects user identity headers.
- Dedicated product-service with order-service reading product data without ownership.
- Authentik OIDC/JWKS flow if the project moves back toward external identity-provider architecture.
- Pagination/filtering for products and orders.
- Observability, CI/CD, and rate-limit hardening.

## Priority Order

1. Add screenshots or a short recorded demo if the README will be used as a portfolio landing page.
2. Gateway/Auth/Product-service separation.
3. Observability, CI/CD, and production hardening.

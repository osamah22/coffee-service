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

## Next Work

- [ ] Add tests for notification email formatting, SMTP sender behavior, and consumer ack/nack/idempotency paths.
- [ ] Add handler-level tests for role access, order ownership/history, and staff status transitions.
- [ ] Run and document a full Docker Compose smoke test.
- [ ] Decide whether Swagger is still needed; if yes, add current annotations, wire `/swagger`, and regenerate docs.
- [ ] Add a Makefile or task runner for common checks.
- [ ] Add screenshots or a short demo flow to the README.

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

1. Notification tests and compose smoke test.
2. Handler/auth coverage.
3. Documentation polish with screenshots/demo.
4. Swagger regeneration only if it is useful for the portfolio demo.
5. Gateway/Auth/Product-service separation.

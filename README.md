# Coffee Service

Portfolio-grade coffee ordering system built with Go microservice-style boundaries, a React frontend, PostgreSQL, RabbitMQ events, and a notification worker.

The current implementation keeps products and orders in the order-service for simplicity, while preserving clear internal boundaries between handlers, services, and models. Authentication is currently handled with SuperTokens directly in the order-service. A separate gateway and Authentik/JWKS flow remain future architecture work.

## Current Features

- Product menu CRUD with seeded coffee products.
- Order checkout that looks up product prices server-side.
- Customer order history by authenticated user email or supplied email.
- Staff order queue with preparing, ready, completed, and cancelled states.
- Transactional outbox records for `order.created` and `order.status_updated`.
- RabbitMQ dispatcher for publishing order events.
- Notification service that consumes order events and sends email through SMTP, with MailHog support for local development.
- React frontend with retro/pixel styling, cart checkout, order history, staff queue, login, signup, and role-aware controls.
- SQLite-backed service/model tests for isolated test databases.

## Stack

- Go, Gin, GORM
- React, Vite
- PostgreSQL for local runtime data
- SQLite for tests
- RabbitMQ for order events
- SuperTokens for current auth/session handling
- MailHog for local notification email capture
- Docker Compose for local infrastructure

## Running Locally

Start the full stack:

```bash
docker compose up --build
```

Default local endpoints:

- Frontend: `http://localhost`
- Order API: `http://localhost:8080`
- Health check: `http://localhost:8080/ping`
- SuperTokens: `http://localhost:3567`
- RabbitMQ management: `http://localhost:15672`
- MailHog UI: `http://localhost:8025`

Default role emails are configured through compose:

- Admin: `admin@example.com`
- Barista: `barista@example.com`
- Other signed-up users receive the `user` role.

## API Overview

Products:

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| GET | `/products` | guest/user/admin | List menu products |
| GET | `/products/:id` | guest/user/admin | Get one product |
| POST | `/products` | admin | Create a product |
| PUT | `/products/:id` | admin | Update a product |
| DELETE | `/products/:id` | admin | Delete a product |

Orders:

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| POST | `/orders` | guest/user/admin | Create an order from product IDs and quantities |
| GET | `/orders/mine` | guest/user/admin | List customer orders |
| GET | `/orders` | barista/admin | List all orders |
| GET | `/orders/:id` | barista/admin | Get one order |
| POST | `/orders/:id/ready` | barista/admin | Mark preparing order ready |
| POST | `/orders/:id/complete` | barista/admin | Mark ready order completed |
| POST | `/orders/:id/cancel` | barista/admin | Cancel preparing or ready order |
| DELETE | `/orders/:id` | admin | Delete an order |

Auth routes are mounted under `/auth` through SuperTokens.

## Events

Order events are published to the `coffee.orders` topic exchange:

- `order.created`
- `order.status_updated`

Events are written to the order-service outbox inside the same database transaction as the order change, then dispatched to RabbitMQ by the background outbox dispatcher. The notification service binds its durable queue to both event types.

## Tests and Checks

Run Go checks by module:

```bash
cd order-service && go test ./...
cd ../notification-service && go test ./...
cd ../shared && go test ./...
```

Run the frontend build:

```bash
cd frontend && npm run build
```

Swagger files were removed because the generated output was stale and not wired into the current server. Regenerate and serve Swagger only after the handler annotations match the current API.

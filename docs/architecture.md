# Architecture

Coffee Service is a compact service-oriented demo with a frontend, two HTTP APIs, RabbitMQ, and an event-only notification worker.

## System Views

### Current Runtime

![Coffee Service system context Excalidraw diagram](diagrams/system-context.svg)

[Edit Excalidraw source](diagrams/system-context.excalidraw)

![Runtime system Excalidraw diagram](diagrams/runtime-system.svg)

[Edit Excalidraw source](diagrams/runtime-system.excalidraw)

## Runtime Containers

| Container | Purpose |
| --- | --- |
| `frontend` | Serves the Vite-built React console through Nginx. |
| `auth-service` | Owns users, password hashes, JWT issuance, role identity, and auth outbox events. |
| `order-service` | Owns products, orders, checkout, status workflow, and order outbox events. |
| `notification-service` | Consumes order/auth facts and sends emails. It does not write to another service database. |
| `postgres` | Shared PostgreSQL instance with service-owned tables. |
| `rabbitmq` | Topic exchange transport for service facts. |
| `mailhog` | Local SMTP sink and email inspection UI. |

## Service Boundaries

- `auth-service`: email/password auth and roles (`user`, `barista`, `admin`).
- `order-service`: products, orders, and barista workflow.
- `notification-service`: event consumer only.

Shared packages remain intentionally small:

- `shared/auth`: JWT issuing/validation, role parsing, and CORS.
- `shared/events`: event names and payload contracts.
- `shared/rabbitmq`: AMQP helpers.

`order-service` validates bearer tokens locally through `shared/auth`. It does not make a runtime HTTP call back to `auth-service` for each request.

## Component Ownership

![Service ownership Excalidraw diagram](diagrams/service-ownership.svg)

[Edit Excalidraw source](diagrams/service-ownership.excalidraw)

## Auth And Roles

The frontend logs in through `auth-service`:

```json
{
  "email": "customer@example.com",
  "password": "customer123"
}
```

The auth API returns a JWT. Order-service trusts that token and enforces route access:

- `user`: menu, checkout, own order history.
- `barista`: menu and queue/status actions.
- `admin`: both user and barista capabilities.

## Checkout Flow

![Checkout sequence Excalidraw diagram](diagrams/architecture-checkout-sequence.svg)

[Edit Excalidraw source](diagrams/architecture-checkout-sequence.excalidraw)

## Events

RabbitMQ carries facts between services:

- `coffee.orders`: `order.created`, `order.status_updated`
- `coffee.auth`: `password_reset.requested`

Both producer services use the transactional outbox pattern so database state and publish intent are recorded atomically.

## Future Target

![Future target shape Excalidraw diagram](diagrams/future-target-shape.svg)

[Edit Excalidraw source](diagrams/future-target-shape.excalidraw)

## Database Diagrams

### Auth Service Schema

- `users`: identity, password hash, and role ownership.
- `outbox_events`: durable auth facts such as `password_reset.requested`.
- Both are written by `auth-service` only, even though they live in the shared local PostgreSQL instance.

### Order Service Schema

![Data model Excalidraw diagram](diagrams/data-model.svg)

[Edit Excalidraw source](diagrams/data-model.excalidraw)

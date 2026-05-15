# Credentials And Demo Notes

This page collects local credentials plus short defense-ready answers for the current split-auth architecture.

## Demo Auth

The frontend signs in through `auth-service` with email/password and then stores a bearer JWT.

| Role | Email | Password |
| --- | --- | --- |
| User | `customer@example.com` | `customer123` |
| Barista | `barista@coffee.local` | `barista123` |
| Admin | `admin@coffee.local` | `admin123` |

Login request:

```json
{
  "email": "customer@example.com",
  "password": "customer123"
}
```

After login, browser requests send:

```text
Authorization: Bearer <token>
```

## Local Infrastructure Credentials

| Component | URL or DSN | Username | Password | Notes |
| --- | --- | --- | --- | --- |
| Frontend | `http://localhost` | None | None | React console served by Nginx. |
| Auth API | `http://localhost:8081` | Demo credentials above | Demo credentials above | Login and token-backed identity. |
| Order API | `http://localhost:8080` | Bearer JWT | Bearer JWT | Product/order API. |
| PostgreSQL | `postgres://postgres:postgres@localhost:5432/coffee` | `postgres` | `postgres` | Shared instance; auth and order own separate tables. |
| RabbitMQ AMQP | `amqp://guest:guest@127.0.0.1:5672/` | `guest` | `guest` | Used by both outbox dispatchers and notification-service. |
| RabbitMQ UI | `http://localhost:15672` | `guest` | `guest` | Management UI. |
| MailHog UI | `http://localhost:8025` | None | None | Local email inspection UI. |
| MailHog SMTP | `localhost:1025` | None | None | Notification-service sends here locally. |

## Quick Defense Answers

| Question | Answer |
| --- | --- |
| How many services? | Three application services: auth-service, order-service, notification-service. |
| How many APIs? | Two HTTP APIs: auth-service and order-service. Notification-service is event-only. |
| Why split auth? | Email/password auth now has its own user table, JWT boundary, and auth events without mixing that logic into order ownership. |
| Which roles exist? | `user`, `barista`, `admin`. |
| Why RabbitMQ? | It decouples both order and auth side effects from email delivery and lets notification handling retry independently. |

## Database Snapshot

See [architecture.md](architecture.md) for the ER diagrams. The runtime PostgreSQL volume now contains:

- `users` and `outbox_events` owned by `auth-service`
- `products`, `orders`, `line_items`, and order outbox rows owned by `order-service`

## Diagram Pack

### Auth And Access

![Auth and role sequence Excalidraw diagram](diagrams/auth-role-sequence.svg)

[Edit Excalidraw source](diagrams/auth-role-sequence.excalidraw)

![Role resolution Excalidraw diagram](diagrams/role-resolution.svg)

[Edit Excalidraw source](diagrams/role-resolution.excalidraw)

![Role access matrix Excalidraw diagram](diagrams/role-access-matrix.svg)

[Edit Excalidraw source](diagrams/role-access-matrix.excalidraw)

### Frontend And Checkout

![Frontend workflow Excalidraw diagram](diagrams/frontend-workflow.svg)

[Edit Excalidraw source](diagrams/frontend-workflow.excalidraw)

![Checkout sequence Excalidraw diagram](diagrams/checkout-sequence.svg)

[Edit Excalidraw source](diagrams/checkout-sequence.excalidraw)

### Events And State

![Outbox and events Excalidraw diagram](diagrams/outbox-events.svg)

[Edit Excalidraw source](diagrams/outbox-events.excalidraw)

![Order event routing Excalidraw diagram](diagrams/order-event-routing.svg)

[Edit Excalidraw source](diagrams/order-event-routing.excalidraw)

![Order state machine Excalidraw diagram](diagrams/order-state-machine.svg)

[Edit Excalidraw source](diagrams/order-state-machine.excalidraw)

![Data model Excalidraw diagram](diagrams/data-model.svg)

[Edit Excalidraw source](diagrams/data-model.excalidraw)

### Future Split

![Future target shape Excalidraw diagram](diagrams/future-target-shape.svg)

[Edit Excalidraw source](diagrams/future-target-shape.excalidraw)

## Reset Local Data

```bash
docker compose down -v
```

That deletes runtime data for users, products, orders, line items, and outbox rows.

# Architecture

Coffee Service is a compact service-oriented demo with a frontend, two HTTP APIs, RabbitMQ, and an event-only notification worker.

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

## Events

RabbitMQ carries facts between services:

- `coffee.orders`: `order.created`, `order.status_updated`
- `coffee.auth`: `password_reset.requested`

Both producer services use the transactional outbox pattern so database state and publish intent are recorded atomically.

## Database Diagrams

### Auth Service Schema

```mermaid
erDiagram
  USERS {
    uuid id PK
    string email UK
    string password_hash
    string role
    timestamp created_at
    timestamp updated_at
  }

  AUTH_OUTBOX_EVENTS {
    uuid id PK
    string event_type
    string aggregate_type
    string aggregate_id
    string routing_key
    text payload
    int attempts
    text last_error
    timestamp occurred_at
    timestamp published_at
    timestamp created_at
  }

  USERS ||--o{ AUTH_OUTBOX_EVENTS : emits
```

### Order Service Schema

```mermaid
erDiagram
  PRODUCTS {
    uuid id PK
    string name
    string category
    int64 price_in_kurus
    string image_path
    bool available
  }

  ORDERS {
    uuid id PK
    string customer_email
    int64 total
    string status
    timestamp created_at
    timestamp updated_at
  }

  LINE_ITEMS {
    uuid id PK
    uuid order_id FK
    uuid product_id
    string product_name
    int quantity
    int64 price_in_kurus
  }

  ORDER_OUTBOX_EVENTS {
    uuid id PK
    string event_type
    string aggregate_type
    string aggregate_id
    string routing_key
    text payload
    int attempts
    text last_error
    timestamp occurred_at
    timestamp published_at
    timestamp created_at
  }

  ORDERS ||--o{ LINE_ITEMS : contains
  ORDERS ||--o{ ORDER_OUTBOX_EVENTS : emits
```

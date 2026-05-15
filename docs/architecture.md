# Architecture

Coffee Service is a compact service-oriented demo with a small frontend, one HTTP API, RabbitMQ, and a second event-consuming service.

## Runtime Containers

| Container | Purpose |
| --- | --- |
| `frontend` | Serves the Vite-built React console through Nginx. |
| `order-service` | Owns products, orders, checkout, custom basic-auth login, JWT validation, staff status workflow, and outbox dispatch. |
| `notification-service` | Consumes order events and sends order lifecycle emails. It does not write to another service database. |
| `postgres` | Primary runtime database for products, orders, line items, and outbox rows. |
| `rabbitmq` | Topic exchange transport for order events. |
| `mailhog` | Local SMTP sink and email inspection UI. |

## Service Boundaries

The application has two app services:

- `order-service`: owns product/menu data and orders for the demo.
- `notification-service`: consumes events only and sends notifications.

The order-service keeps the same simple Go structure:

- `internal/models`: GORM models and validation.
- `internal/services`: business behavior, transactions, and outbox creation.
- `internal/handlers`: HTTP request/response layer.
- `shared/auth`: basic-auth login, JWT issuing/validation, and role checks.
- `shared/events`: event names and JSON payload contracts.
- `shared/rabbitmq`: exchange declaration and publishing helpers.

## Auth And Roles

Auth is intentionally simple for the project defense. The frontend logs in with:

```text
POST /auth/login
Authorization: Basic base64(email:password)
```

The API returns a bearer token. Application requests send `Authorization: Bearer <jwt>`, and handlers enforce role checks:

- `customer`: menu, checkout, own order history.
- `staff`: menu and staff order queue/status changes.
- `admin`: customer and staff routes.

This demonstrates custom authentication and authorization without adding an external identity provider to the demo.

## Events

RabbitMQ carries facts from the order-service to the notification-service:

- `order.created`
- `order.status_updated`

The order-service writes outbox rows in the same transaction as order changes, then a background dispatcher publishes them to the `coffee.orders` topic exchange. This keeps checkout independent from notification delivery.

## Data Ownership

| Data | Owner | Notes |
| --- | --- | --- |
| Products | Order service | Kept local for demo simplicity and server-side price lookup. |
| Orders | Order service | Includes line items and status lifecycle. |
| Outbox events | Order service | Published asynchronously to RabbitMQ. |
| Notification delivery | Notification service process | No cross-service database writes. |

## Intentional Non-Goals

- No separate product service.
- No gateway.
- No production identity provider.
- No extra event types unless another service consumes the fact.

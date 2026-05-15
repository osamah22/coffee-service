# Credentials And Demo Notes

This page collects local credentials and the short explanation points for the simplified Coffee Service demo.

## Demo Auth

The frontend now uses local demo accounts with HTTP Basic login and JWT bearer sessions.

| Role | Email | Password |
| --- | --- | --- |
| Customer | `customer@example.com` | `customer123` |
| Staff | `staff@coffee.local` | `staff123` |
| Admin | `admin@coffee.local` | `admin123` |

Login request:

```text
POST /auth/login
Authorization: Basic base64(email:password)
```

After login, the frontend stores the JWT and sends:

```text
Authorization: Bearer <token>
```

## Local Infrastructure Credentials

These credentials are for local development only.

| Component | URL or DSN | Username | Password | Notes |
| --- | --- | --- | --- | --- |
| Frontend | `http://localhost` | None | None | React console served by Nginx. |
| Order API | `http://localhost:8080` | Basic auth at login, then Bearer JWT | Demo credentials above | Single HTTP API. |
| PostgreSQL | `postgres://postgres:postgres@localhost:5432/coffee` | `postgres` | `postgres` | Runtime DB for order-service. |
| RabbitMQ AMQP | `amqp://guest:guest@127.0.0.1:5672/` | `guest` | `guest` | Used by outbox dispatcher and notification service. |
| RabbitMQ UI | `http://localhost:15672` | `guest` | `guest` | Management UI from the RabbitMQ image. |
| MailHog UI | `http://localhost:8025` | None | None | Local email inspection UI. |
| MailHog SMTP | `localhost:1025` | None | None | Notification service sends here locally. |

## Quick Defense Answers

| Question | Answer |
| --- | --- |
| How many services? | Two application services: order-service and notification-service. |
| How many APIs? | One HTTP API: order-service. Notification-service is event-only. |
| How many endpoints? | Eight demo endpoints including `/ping`. |
| Why RabbitMQ? | It decouples order creation from notification delivery and lets notification processing retry independently. |
| Do you have auth? | Yes, a custom basic-auth login endpoint issues JWTs and handlers enforce role checks from the bearer token. |
| Where is CI/CD? | `.github/workflows/ci-cd.yml`. It runs tests/build checks and publishes Docker images on `main`. |

## Main Runtime Flow

![Runtime system Excalidraw diagram](diagrams/runtime-system.svg)

[Edit Excalidraw source](diagrams/runtime-system.excalidraw)

## Reset Local Data

```bash
docker compose down -v
```

That deletes PostgreSQL runtime data, including products, orders, line items, and outbox rows.

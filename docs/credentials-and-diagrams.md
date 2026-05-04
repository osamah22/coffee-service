# Credentials And Diagrams

This page collects the local demo credentials, role rules, and end-to-end diagrams for the current Coffee Service app.

## Local Login Credentials

Application users are managed by SuperTokens. The app does not seed fixed application passwords.

To log in as a demo role, sign up first through the frontend at `http://localhost` using one of the role emails below and a password with at least 6 characters. After signup, use the same email and password to log in.

| Role | Email | Password |
| --- | --- | --- |
| Admin | `admin@example.com` | Chosen during local signup |
| Barista | `barista@example.com` | Chosen during local signup |
| User | Any other valid email | Chosen during local signup |
| Guest | No login | Not applicable |

Role assignment is email based:

- `SUPERTOKENS_ADMIN_EMAILS` defaults to `admin@example.com`.
- `SUPERTOKENS_BARISTA_EMAILS` defaults to `barista@example.com`.
- Any signed-in email not in those lists receives the `user` role.
- Anonymous requests receive the `guest` role where handlers allow guest access.

If you need to reset local application accounts, remove the Docker volume:

```bash
docker compose down -v
```

That deletes PostgreSQL runtime data, including SuperTokens user/session metadata, orders, products, and outbox rows.

## Local Infrastructure Credentials

These credentials are for local development only.

| Component | URL or DSN | Username | Password | Notes |
| --- | --- | --- | --- | --- |
| Frontend | `http://localhost` | App email | Signup password | React console served by Nginx. |
| Order API | `http://localhost:8080` | SuperTokens session cookie | SuperTokens session cookie | `/auth` routes handle signup/login/logout. |
| PostgreSQL | `postgres://postgres:postgres@localhost:5432/coffee` | `postgres` | `postgres` | Runtime DB for order-service and SuperTokens. |
| RabbitMQ AMQP | `amqp://guest:guest@localhost:5672/` | `guest` | `guest` | Used by outbox dispatcher and notification service. |
| RabbitMQ UI | `http://localhost:15672` | `guest` | `guest` | Management UI from the RabbitMQ image. |
| SuperTokens | `http://localhost:3567` | None configured | None configured | Session/auth provider API. |
| MailHog UI | `http://localhost:8025` | None | None | Local email inspection UI. |
| MailHog SMTP | `localhost:1025` | None | None | Notification service sends here locally. |

## Runtime System

![Runtime system Excalidraw diagram](diagrams/runtime-system.svg)

[Edit Excalidraw source](diagrams/runtime-system.excalidraw)

## Service Ownership

![Service ownership Excalidraw diagram](diagrams/service-ownership.svg)

[Edit Excalidraw source](diagrams/service-ownership.excalidraw)

## Auth And Role Flow

![Auth and role sequence Excalidraw diagram](diagrams/auth-role-sequence.svg)

[Edit Excalidraw source](diagrams/auth-role-sequence.excalidraw)

![Role resolution Excalidraw diagram](diagrams/role-resolution.svg)

[Edit Excalidraw source](diagrams/role-resolution.excalidraw)

## Role Access Matrix

![Role access matrix Excalidraw diagram](diagrams/role-access-matrix.svg)

[Edit Excalidraw source](diagrams/role-access-matrix.excalidraw)

## Frontend Workflow

![Frontend workflow Excalidraw diagram](diagrams/frontend-workflow.svg)

[Edit Excalidraw source](diagrams/frontend-workflow.excalidraw)

## Checkout Sequence

![Checkout sequence Excalidraw diagram](diagrams/checkout-sequence.svg)

[Edit Excalidraw source](diagrams/checkout-sequence.excalidraw)

## Outbox And Events

![Outbox and events Excalidraw diagram](diagrams/outbox-events.svg)

[Edit Excalidraw source](diagrams/outbox-events.excalidraw)

Order event routing:

![Order event routing Excalidraw diagram](diagrams/order-event-routing.svg)

[Edit Excalidraw source](diagrams/order-event-routing.excalidraw)

## Notification Flow

![Notification flow Excalidraw diagram](diagrams/notification-flow.svg)

[Edit Excalidraw source](diagrams/notification-flow.excalidraw)

## Order State Machine

![Order state machine Excalidraw diagram](diagrams/order-state-machine.svg)

[Edit Excalidraw source](diagrams/order-state-machine.excalidraw)

## Data Model

![Data model Excalidraw diagram](diagrams/data-model.svg)

[Edit Excalidraw source](diagrams/data-model.excalidraw)

SuperTokens also stores user and session metadata in PostgreSQL tables owned by the SuperTokens service. The application code treats those tables as provider-owned data.

## Future Target Shape

![Future target shape Excalidraw diagram](diagrams/future-target-shape.svg)

[Edit Excalidraw source](diagrams/future-target-shape.excalidraw)

Future services should keep the same rule: events are facts, consumers are idempotent, and downstream services trust identity headers only from the gateway network path.

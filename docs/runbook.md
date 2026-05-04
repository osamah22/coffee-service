# Operations Runbook

This runbook covers local development and portfolio demo operation.

## Start

```bash
docker compose up --build
```

Expected endpoints:

| Component | URL |
| --- | --- |
| Frontend | `http://localhost` |
| Order API | `http://localhost:8080` |
| Health | `http://localhost:8080/ping` |
| SuperTokens | `http://localhost:3567` |
| RabbitMQ management | `http://localhost:15672` |
| MailHog | `http://localhost:8025` |

## Smoke Test

```bash
curl http://localhost:8080/ping
curl -I http://localhost
curl http://localhost:8025
```

Then place an order from the frontend and verify an email appears in MailHog.

## Stop

```bash
docker compose down
```

Reset local data:

```bash
docker compose down -v
```

## Checks

```bash
make check
```

The check target runs:

1. `go test ./...` in `order-service`
2. `go test ./...` in `notification-service`
3. `go test ./...` in `shared`
4. `npm run build` in `frontend`
5. `docker compose config`

## Important Environment Variables

Order service:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port. |
| `DB_URL` | local Postgres DSN | Runtime database connection. |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection. |
| `SUPERTOKENS_CONNECTION_URI` | `http://localhost:3567` | SuperTokens service URL. |
| `SUPERTOKENS_API_DOMAIN` | `http://localhost:8080` | API cookie/session domain config. |
| `SUPERTOKENS_WEBSITE_DOMAIN` | `http://localhost:5173` | Frontend domain config. Compose overrides this to `http://localhost`. |
| `SUPERTOKENS_ADMIN_EMAILS` | `admin@example.com` | Comma-separated admin email list. |
| `SUPERTOKENS_BARISTA_EMAILS` | `barista@example.com` | Comma-separated barista email list. |
| `AUTH_GUEST_RPS` | `10` | Guest requests per second. |
| `AUTH_USER_RPS` | `10` | User requests per second. |
| `AUTH_BARISTA_RPS` | `250` | Barista requests per second. |
| `AUTH_ADMIN_RPS` | `1000` | Admin requests per second. |

Notification service:

| Variable | Default | Purpose |
| --- | --- | --- |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection. |
| `SMTP_HOST` | empty | SMTP server host. If empty, sender logs instead of using SMTP. |
| `SMTP_PORT` | `1025` | SMTP port. |
| `SMTP_USERNAME` | empty | Optional SMTP username. |
| `SMTP_PASSWORD` | empty | Optional SMTP password. |
| `SMTP_FROM` | `Coffee Service <orders@coffee.local>` | Sender address. |
| `NOTIFICATION_FALLBACK_EMAIL` | `dev@coffee.local` | Fallback when an event has no customer email. |

Frontend:

| Variable | Default | Purpose |
| --- | --- | --- |
| `VITE_API_URL` | `http://localhost:8080` | Order API base URL used by the browser. |

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Frontend cannot log in | Confirm `SUPERTOKENS_WEBSITE_DOMAIN` matches the browser origin and cookies are accepted. |
| Orders are created but no email appears | Check RabbitMQ health, notification-service logs, and MailHog at `http://localhost:8025`. |
| Product list is empty | Check order-service startup logs for migration or seed errors. |
| `make check` fails at Docker Compose config | Run `docker compose version` and confirm Docker is installed. |
| Guest requests get `429` | Wait one second or raise `AUTH_GUEST_RPS` for local load testing. |

## Local Data

PostgreSQL data is stored in the named volume `order-service_postgres_data`. Removing the volume resets products, orders, outbox rows, and SuperTokens metadata.

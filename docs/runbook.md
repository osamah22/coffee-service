# Operations Runbook

This runbook covers the simplified local demo.

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
| RabbitMQ management | `http://localhost:15672` |
| MailHog | `http://localhost:8025` |

## Smoke Test

```bash
curl http://localhost:8080/ping
curl -I http://localhost
curl http://localhost:8025
```

Then place an order from the frontend and verify an email appears in MailHog.

Default login accounts:

- `customer@example.com` / `customer123`
- `staff@coffee.local` / `staff123`
- `admin@coffee.local` / `admin123`

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

The check target runs Go tests, builds the frontend, and validates Docker Compose.

## Important Environment Variables

Order service:

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port. |
| `DB_URL` | local Postgres DSN | Runtime database connection. |
| `RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ connection. |
| `API_DOMAIN` | `http://localhost:8080` | API origin allowed by CORS. |
| `WEBSITE_DOMAIN` | `http://localhost:5173` | Frontend origin allowed by CORS. Compose sets this to `http://localhost`. |
| `JWT_SECRET` | `coffee-service-local-jwt-secret` | HMAC secret for signed bearer tokens. |
| `JWT_ISSUER` | `coffee-service` | JWT issuer claim value. |
| `JWT_TTL_MINUTES` | `480` | Token lifetime in minutes. |
| `DEMO_USERS` | built-in customer/staff/admin accounts | Comma-separated `email:password:role[:subject]` entries. |

Notification service:

| Variable | Default | Purpose |
| --- | --- | --- |
| `RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ connection. |
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
| Staff queue returns `403` | Log in with the staff or admin demo account, then retry the queue request with the bearer token. |
| Products return `401` | Log in again so the frontend stores a fresh JWT, or send `Authorization: Bearer <token>`. |
| Orders are created but no email appears | Check RabbitMQ health, notification-service logs, and MailHog at `http://localhost:8025`. |
| Product list is empty | Check order-service startup logs for migration or seed errors. |
| `make check` fails at Docker Compose config | Run `docker compose version` and confirm Docker is installed. |

## Local Data

PostgreSQL data is stored in the named volume `order-service_postgres_data`. Removing the volume resets products, orders, line items, and outbox rows.

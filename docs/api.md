# API Reference

Local Compose API bases:

- Auth API: `http://localhost:8081`
- Order API: `http://localhost:8080`

## Authentication

Login is handled by `auth-service` with JSON email/password input.

```http
POST /auth/login
Content-Type: application/json

{
  "email": "customer@example.com",
  "password": "customer123"
}
```

Authenticated application routes use:

| Header | Values | Purpose |
| --- | --- | --- |
| `Authorization` | `Bearer <jwt>` | Carries the signed session token issued by `auth-service`. |

Default local demo accounts:

| Email | Password | Role |
| --- | --- | --- |
| `customer@example.com` | `customer123` | `user` |
| `barista@coffee.local` | `barista123` | `barista` |
| `admin@coffee.local` | `admin123` | `admin` |

## Auth Endpoints

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/ping` | public | Auth health check. |
| `POST` | `/auth/login` | public | Exchanges email/password for a JWT. |
| `GET` | `/auth/me` | authenticated | Returns the current token subject, email, and role. |
| `POST` | `/auth/password-reset-requests` | public | Enqueues `password_reset.requested` if the email exists. Always returns `202`. |

Login response:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_at": "2026-05-15T18:00:00Z",
  "user": {
    "id": "6f1ac76e-f1d3-45f0-a4da-f123456789ab",
    "email": "customer@example.com",
    "role": "user"
  }
}
```

Password reset request:

```json
{
  "email": "customer@example.com"
}
```

## Order Endpoints

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `GET` | `/ping` | public | Order health check. |
| `GET` | `/products` | user, barista, admin | Lists all menu products. |
| `POST` | `/orders` | user, admin | Creates an order and enqueues `order.created`. |
| `GET` | `/orders/mine?email=:email` | user, admin | Lists orders for the authenticated user email. |
| `GET` | `/staff/orders` | barista, admin | Lists all orders for the staff queue. |
| `POST` | `/staff/orders/:id/ready` | barista, admin | Moves `preparing` to `ready` and enqueues `order.status_updated`. |
| `POST` | `/staff/orders/:id/complete` | barista, admin | Moves `ready` to `completed` and enqueues `order.status_updated`. |
| `POST` | `/staff/orders/:id/cancel` | barista, admin | Cancels `preparing` or `ready` and enqueues `order.status_updated`. |

## Products

Product response:

```json
{
  "id": "0c94a67d-a6cb-4429-bf31-97f5fa8f673f",
  "name": "Caffe Latte",
  "category": "hot",
  "price_in_kurus": 8500,
  "image_path": "/products/latte.png",
  "available": true
}
```

## Orders

Create request:

```json
{
  "customer_email": "customer@example.com",
  "items": [
    {
      "product_id": "0c94a67d-a6cb-4429-bf31-97f5fa8f673f",
      "quantity": 2
    }
  ]
}
```

Prices and product names are loaded server-side from stored product records. The client sends only product IDs and quantities.

Order response:

```json
{
  "id": "711f2c78-bb2b-4192-b1ab-f69dc4b92775",
  "customer_email": "customer@example.com",
  "items": [
    {
      "product_id": "0c94a67d-a6cb-4429-bf31-97f5fa8f673f",
      "product_name": "Caffe Latte",
      "quantity": 2,
      "price_in_kurus": 8500
    }
  ],
  "total": 17000,
  "status": "preparing",
  "created_at": "2026-05-04T12:00:00Z"
}
```

Valid status transitions:

- `preparing -> ready`
- `ready -> completed`
- `preparing -> cancelled`
- `ready -> cancelled`

## Errors

Most errors return:

```json
{
  "error": "invalid_order_status_transition"
}
```

Common status codes:

| Status | Meaning |
| --- | --- |
| `400` | Invalid request, validation error, missing customer email, invalid transition, or unknown product. |
| `401` | Missing/invalid Basic credentials or missing/invalid JWT. |
| `403` | Authenticated role is not allowed. |
| `404` | Resource not found. |
| `500` | Unexpected service/database failure. |

# API Reference

Base URL for local Compose: `http://localhost:8080`.

The project exposes one HTTP API: `order-service`.

## Authentication

Login uses HTTP Basic authentication:

| Header | Values | Purpose |
| --- | --- | --- |
| `Authorization` | `Basic base64(email:password)` | Used only on `POST /auth/login`. |

Authenticated application routes use:

| Header | Values | Purpose |
| --- | --- | --- |
| `Authorization` | `Bearer <jwt>` | Carries the signed demo session token. |

Default local demo accounts:

| Email | Password | Role |
| --- | --- | --- |
| `customer@example.com` | `customer123` | `customer` |
| `staff@coffee.local` | `staff123` | `staff` |
| `admin@coffee.local` | `admin123` | `admin` |

## Endpoints

| Method | Path | Role | Description |
| --- | --- | --- | --- |
| `POST` | `/auth/login` | public | Exchanges HTTP Basic credentials for a JWT. |
| `GET` | `/auth/me` | authenticated | Returns the current JWT subject, email, and role. |
| `GET` | `/ping` | public | Returns `{"message":"pong"}`. |
| `GET` | `/products` | customer, staff, admin | Lists all menu products. |
| `POST` | `/orders` | customer, admin | Creates an order and enqueues `order.created`. |
| `GET` | `/orders/mine?email=:email` | customer, admin | Lists orders for the customer email. |
| `GET` | `/staff/orders` | staff, admin | Lists all orders for the staff queue. |
| `POST` | `/staff/orders/:id/ready` | staff, admin | Moves `preparing` to `ready` and enqueues `order.status_updated`. |
| `POST` | `/staff/orders/:id/complete` | staff, admin | Moves `ready` to `completed` and enqueues `order.status_updated`. |
| `POST` | `/staff/orders/:id/cancel` | staff, admin | Cancels `preparing` or `ready` and enqueues `order.status_updated`. |

## Login

Login response:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_at": "2026-05-15T18:00:00Z",
  "user": {
    "id": "customer-1",
    "email": "customer@example.com",
    "role": "customer"
  }
}
```

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

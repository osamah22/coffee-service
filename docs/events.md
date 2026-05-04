# Event Contracts

Order events are facts emitted by the order-service through the transactional outbox. They are intentionally minimal and stable.

## Transport

```mermaid
flowchart LR
  OrderDB[(order-service DB)] --> Outbox[outbox_events]
  Outbox --> Dispatcher[Outbox dispatcher]
  Dispatcher --> Exchange[(RabbitMQ topic exchange: coffee.orders)]
  Exchange --> Queue[notification-service.orders]
  Queue --> Consumer[Notification consumer]
```

| Setting | Value |
| --- | --- |
| Exchange | `coffee.orders` |
| Exchange type | Topic |
| Queue | `notification-service.orders` |
| Routing keys | `order.created`, `order.status_updated` |
| Publisher | order-service outbox dispatcher |
| Consumer | notification-service |

## Rules

- Events describe completed facts, not commands.
- Event payloads should remain backward compatible.
- Consumers must be idempotent by `event_id` where possible.
- Only add an event when another service needs that fact.
- Notification service must not write to order-service data.

## `order.created`

Published after a new order and its line items are committed.

```json
{
  "event_id": "c08c9d12-8579-42e4-bd29-3fd7b36f97d8",
  "order_id": "711f2c78-bb2b-4192-b1ab-f69dc4b92775",
  "customer_email": "customer@example.com",
  "status": "preparing",
  "items": [
    {
      "product_id": "0c94a67d-a6cb-4429-bf31-97f5fa8f673f",
      "product_name": "Caffe Latte",
      "quantity": 2,
      "price_in_kurus": 8500
    }
  ],
  "total": 17000,
  "occurred_at": "2026-05-04T12:00:00Z"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `event_id` | string UUID | Idempotency key. |
| `order_id` | string UUID | Aggregate identifier. |
| `customer_email` | string | Receipt destination. |
| `status` | string | Initial order status. |
| `items` | array | Product snapshot at checkout time. |
| `total` | integer | Total in kurus. |
| `occurred_at` | timestamp | UTC event time. |

## `order.status_updated`

Published after a valid order status transition is committed.

```json
{
  "event_id": "9f49e19b-b1af-48e0-b9a4-8af10ca0d1d2",
  "order_id": "711f2c78-bb2b-4192-b1ab-f69dc4b92775",
  "customer_email": "customer@example.com",
  "previous_status": "preparing",
  "status": "ready",
  "occurred_at": "2026-05-04T12:05:00Z"
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `event_id` | string UUID | Idempotency key. |
| `order_id` | string UUID | Aggregate identifier. |
| `customer_email` | string | Notification destination. |
| `previous_status` | string | Status before transition. |
| `status` | string | New status. |
| `occurred_at` | timestamp | UTC event time. |

## Outbox Lifecycle

```mermaid
sequenceDiagram
  participant Service as Order service
  participant DB as PostgreSQL
  participant Worker as Outbox dispatcher
  participant MQ as RabbitMQ

  Service->>DB: Begin transaction
  Service->>DB: Write order/status change
  Service->>DB: Insert outbox event
  Service->>DB: Commit
  Worker->>DB: Load unpublished outbox rows
  Worker->>MQ: Publish with event type and event ID
  MQ-->>Worker: Publish accepted
  Worker->>DB: Mark published_at
```

The dispatcher can safely retry unpublished rows. Consumers should still handle duplicate delivery because RabbitMQ delivery is at-least-once.

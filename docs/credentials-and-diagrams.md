# Credentials And Demo Notes

This page collects local credentials plus short defense-ready answers for the current split-auth architecture.

## Demo Auth

The frontend signs in through `auth-service` with email/password and then stores a bearer JWT.

![Demo auth credentials table Excalidraw diagram](diagrams/credentials-demo-auth.svg)

[Edit Excalidraw source](diagrams/credentials-demo-auth.excalidraw)

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

![Local infrastructure credentials table Excalidraw diagram](diagrams/credentials-local-infra.svg)

[Edit Excalidraw source](diagrams/credentials-local-infra.excalidraw)

## Quick Defense Answers

![Quick defense answers table Excalidraw diagram](diagrams/credentials-defense-answers.svg)

[Edit Excalidraw source](diagrams/credentials-defense-answers.excalidraw)

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

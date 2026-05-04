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

```mermaid
flowchart LR
  Browser[Browser]
  Frontend[React frontend\nNginx on :80]
  API[Order service\nGo/Gin on :8080]
  ST[SuperTokens\n:3567]
  DB[(PostgreSQL\n:5432)]
  MQ[(RabbitMQ\n:5672 / :15672)]
  Notify[Notification service\nGo consumer]
  MailHog[MailHog\nSMTP :1025 / UI :8025]

  Browser -->|loads app| Frontend
  Browser -->|HTTP JSON + cookies| API
  API -->|signup/login/session recipe| ST
  ST -->|metadata| DB
  API -->|GORM| DB
  API -->|publishes outbox events| MQ
  MQ -->|order facts| Notify
  Notify -->|SMTP| MailHog
```

## Service Ownership

```mermaid
flowchart TB
  subgraph Frontend
    UI[Menu, cart, orders, staff queue]
  end

  subgraph OrderService[Order service]
    Auth[Auth routes and role middleware]
    Products[Product/menu handlers]
    Orders[Order handlers and workflow]
    Outbox[Transactional outbox dispatcher]
  end

  subgraph NotificationService[Notification service]
    Consumer[RabbitMQ consumer]
    Email[Email formatting and sender]
  end

  UI --> Auth
  UI --> Products
  UI --> Orders
  Orders --> Outbox
  Outbox --> Consumer
  Consumer --> Email
```

## Auth And Role Flow

```mermaid
sequenceDiagram
  participant User as Browser user
  participant UI as React frontend
  participant API as Order service
  participant ST as SuperTokens

  User->>UI: Submit email/password
  UI->>API: POST /auth/signup or /auth/signin
  API->>ST: Create or verify email/password user
  API->>API: Resolve role from configured email lists
  API-->>UI: Set SuperTokens cookies
  UI->>API: GET /auth/me
  API-->>UI: {email, role}
```

```mermaid
flowchart TD
  Request[HTTP request] --> HasSession{Valid session?}
  HasSession -- No --> Guest[Role: guest]
  HasSession -- Yes --> Email[Read email claim]
  Email --> IsAdmin{Email in admin list?}
  IsAdmin -- Yes --> Admin[Role: admin]
  IsAdmin -- No --> IsBarista{Email in barista list?}
  IsBarista -- Yes --> Barista[Role: barista]
  IsBarista -- No --> User[Role: user]
  Guest --> Guard[Handler role guard]
  User --> Guard
  Barista --> Guard
  Admin --> Guard
```

## Role Access Matrix

```mermaid
flowchart LR
  Guest[guest] --> Browse[Browse products]
  Guest --> CreateOrder[Create order with email]
  Guest --> Mine[List orders by supplied email]

  User[user] --> Browse
  User --> CreateOwn[Create order with session email]
  User --> MineOwn[List own orders]

  Barista[barista] --> Queue[List all orders]
  Barista --> Status[Ready, complete, cancel]

  Admin[admin] --> Browse
  Admin --> CreateOwn
  Admin --> MineOwn
  Admin --> Queue
  Admin --> Status
  Admin --> ProductCRUD[Product create, update, delete]
  Admin --> DeleteOrder[Delete orders]
```

## Frontend Workflow

```mermaid
flowchart TD
  Start[Open frontend] --> LoadMenu[Load /products]
  LoadMenu --> Menu[Menu view]
  Menu --> Cart[Add products to cart]
  Cart --> Checkout[Submit /orders]
  Checkout --> Orders[Orders view]
  Orders --> Mine[Load /orders/mine]

  Start --> AuthPanel[Login/signup panel]
  AuthPanel --> Me[Fetch /auth/me]
  Me --> RoleView{Role}
  RoleView -- guest/user/admin --> Menu
  RoleView -- barista/admin --> Staff[Barista queue]
  Staff --> UpdateStatus[POST status action]
```

## Checkout Sequence

```mermaid
sequenceDiagram
  participant UI as React frontend
  participant Handler as Order handler
  participant Service as Order service
  participant DB as PostgreSQL
  participant Outbox as Outbox row

  UI->>Handler: POST /orders {customer_email, items}
  Handler->>Handler: Use session email when signed in
  Handler->>Service: CreateOrder(input)
  Service->>DB: Lookup products by ID
  Service->>Service: Calculate trusted prices and total
  Service->>DB: Insert order and line items
  Service->>DB: Insert order.created outbox event
  DB-->>Service: Commit transaction
  Service-->>Handler: Order model
  Handler-->>UI: 201 OrderResponse
```

## Outbox And Events

```mermaid
sequenceDiagram
  participant Order as Order service transaction
  participant DB as PostgreSQL outbox_events
  participant Dispatcher as Outbox dispatcher
  participant MQ as RabbitMQ coffee.orders
  participant Consumer as Notification service

  Order->>DB: Write order change and outbox event atomically
  Dispatcher->>DB: Poll unpublished events
  Dispatcher->>MQ: Publish with routing key
  Dispatcher->>DB: Set published_at or record attempt error
  MQ->>Consumer: Deliver order.created/status_updated
  Consumer-->>MQ: Ack after successful handling
```

Order event routing:

```mermaid
flowchart LR
  Outbox[Order service outbox] --> Exchange[coffee.orders topic exchange]
  Exchange -->|order.created| Queue[notification-service.orders]
  Exchange -->|order.status_updated| Queue
  Queue --> Consumer[Notification consumer]
  Consumer --> Mail[Receipt/status email]
```

## Notification Flow

```mermaid
flowchart TD
  Event[Order event] --> Known{Known event type?}
  Known -- No --> AckUnknown[Ack and ignore]
  Known -- Yes --> Duplicate{Event already seen?}
  Duplicate -- Yes --> AckDuplicate[Ack duplicate]
  Duplicate -- No --> Format[Format customer email]
  Format --> Send[Send through SMTP or log sender]
  Send --> Success{Send succeeded?}
  Success -- Yes --> Remember[Remember event id]
  Remember --> Ack[Ack message]
  Success -- No --> Nack[Nack/retry path]
```

## Order State Machine

```mermaid
stateDiagram-v2
  [*] --> preparing
  preparing --> ready: /orders/:id/ready
  ready --> completed: /orders/:id/complete
  preparing --> cancelled: /orders/:id/cancel
  ready --> cancelled: /orders/:id/cancel
  completed --> [*]
  cancelled --> [*]
```

## Data Model

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
    time created_at
  }

  LINE_ITEMS {
    uuid id PK
    uuid order_id FK
    uuid product_id FK
    int quantity
    int64 price_in_kurus
    string product_name
  }

  OUTBOX_EVENTS {
    uuid id PK
    string event_type
    string aggregate_type
    string aggregate_id
    string routing_key
    text payload
    int attempts
    text last_error
    time occurred_at
    time published_at
    time created_at
  }

  PRODUCTS ||--o{ LINE_ITEMS : "selected by"
  ORDERS ||--o{ LINE_ITEMS : "contains"
  ORDERS ||--o{ OUTBOX_EVENTS : "emits facts for"
```

SuperTokens also stores user and session metadata in PostgreSQL tables owned by the SuperTokens service. The application code treats those tables as provider-owned data.

## Future Target Shape

```mermaid
flowchart LR
  Browser[Browser] --> Gateway[Future Go gateway]
  Gateway -->|validate JWT with JWKS| Authentik[Authentik OIDC]
  Gateway -->|X-User-Sub / X-User-Email| Order[Order service]
  Gateway --> Product[Future product service]
  Order --> OrderDB[(Order DB)]
  Product --> ProductDB[(Product DB)]
  Order --> MQ[(RabbitMQ)]
  MQ --> Notify[Notification service]
```

Future services should keep the same rule: events are facts, consumers are idempotent, and downstream services trust identity headers only from the gateway network path.

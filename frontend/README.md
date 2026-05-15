# Coffee Service Frontend

React/Vite frontend for the simplified Coffee Service demo.

## Features

- Retro/pixel-styled menu UI.
- Product browsing and cart checkout.
- Email/password login against `auth-service` that stores a JWT bearer session.
- Customer order history.
- Barista queue with order status actions.
- Theme toggle and local customer email persistence.

## Local Development

Install dependencies:

```bash
npm install
```

Run the Vite dev server:

```bash
npm run dev
```

Build for production:

```bash
npm run build
```

The API base URLs are controlled by `VITE_API_URL` and `VITE_AUTH_API_URL`. Docker Compose builds the frontend against the Traefik gateway with `http://localhost/api` for orders and `http://localhost` for auth.

Default demo accounts:

- `customer@example.com` / `customer123`
- `barista@coffee.local` / `barista123`
- `admin@coffee.local` / `admin123`

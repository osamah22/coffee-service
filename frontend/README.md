# Coffee Service Frontend

React/Vite frontend for the simplified Coffee Service demo.

## Features

- Retro/pixel-styled menu UI.
- Product browsing and cart checkout.
- Basic-auth login that stores a JWT bearer session.
- Customer order history.
- Staff queue with order status actions.
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

The API base URL is controlled by `VITE_API_URL`; Docker Compose builds the frontend with `http://localhost:8080`.

Default demo accounts:

- `customer@example.com` / `customer123`
- `staff@coffee.local` / `staff123`
- `admin@coffee.local` / `admin123`

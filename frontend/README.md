# Coffee Service Frontend

React/Vite frontend for the Coffee Service demo.

## Features

- Retro/pixel-styled menu UI.
- Product browsing and cart checkout.
- Login, signup, and logout through the API auth integration.
- Customer order history.
- Barista/admin queue with order status actions.
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

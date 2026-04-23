# pdforge Web Frontend

This folder contains the coss ui-based React frontend for pdforge.

## What it uses

- React + Vite + TypeScript
- coss ui-style copy-paste components under `src/components/ui`
- Local-only backend endpoints exposed by `pdforge serve`

## Important endpoints

- `GET /api/csrf` for obtaining the local CSRF token
- `POST /api/merge`
- `POST /api/split`
- `POST /api/remove`
- `POST /api/compress`
- `GET /download?job=...&index=...`

## Build

Install dependencies and build the app once Node.js is available locally:

```bash
npm install
npm run build
```

The Go server prefers `web/dist` when it exists and will serve the built frontend automatically.
When you run `pdforge serve`, it also opens the local browser automatically unless `--no-open` is passed.

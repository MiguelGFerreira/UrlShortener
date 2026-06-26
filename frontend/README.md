# Front-end component

A self-contained React/Next.js client component that talks to the shortener
JSON API. Minimalist styling via [styled-jsx](https://github.com/vercel/styled-jsx)
(built into Next.js) — no UI dependencies, and it inherits the host page's font.

## Use it in your portfolio

1. Copy [`UrlShortener.tsx`](UrlShortener.tsx) into your Next.js app
   (e.g. `components/UrlShortener.tsx`).

2. Point it at your deployed API, either with an env var:

   ```bash
   # .env.local
   NEXT_PUBLIC_SHORTENER_API=https://api.your-domain.com
   ```

   ...or via the `apiBase` prop:

   ```tsx
   import UrlShortener from "@/components/UrlShortener";

   export default function Page() {
     return <UrlShortener apiBase="https://api.your-domain.com" />;
   }
   ```

   When unset, it defaults to `http://localhost:8080` for local development.

## API requirements

The component calls `POST {apiBase}/shorten`. The shortener service must allow
your site's origin via CORS — set `CORS_ALLOWED_ORIGIN` (defaults to `*`, which
lets any origin call it). Generated links use the service's `PUBLIC_BASE_URL`,
so set that to your public redirector address in production.

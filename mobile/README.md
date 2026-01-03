## Environment variables

Create a `.env` file in the `mobile/` directory for local development (or configure via EAS secrets) with:

```
EXPO_PUBLIC_APP_DOMAIN=https://decorebator.com
EXPO_PUBLIC_API_URL=https://api.decorebator.com
```

This is used for building shareable links in the app. In local development, if this is not set, the app will fallback to `http://localhost:3000`.

For local development, use `.env.local` or `.env.development` with a local API URL (for example `http://10.0.2.2:3000`). Do not commit `.env.local`. Production OTA updates should rely on EAS environment variables.

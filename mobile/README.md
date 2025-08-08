## Environment variables

Create a `.env` file in the `mobile/` directory (or configure via EAS secrets) with:

```
EXPO_PUBLIC_APP_DOMAIN=https://decorebator.com
```

This is used to build shareable public quiz links in the app. In local development, if this is not set, the app will fallback to `http://localhost:3000`.



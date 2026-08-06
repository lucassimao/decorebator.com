# Local Maestro suite

This suite is local-development evidence only. It must never point at the production API or seed a production database.

## Prerequisites

1. Start the API dependencies from `api/`. The compose file uses two explicit local volumes:

   ```sh
   docker volume create api_postgres_data
   docker volume create api_minio-data
   docker compose -f docker-compose.yml up -d
   make migrate-up
   ```

2. Start the API with `make run` and confirm `mobile/.env.development` resolves `EXPO_PUBLIC_API_URL` to the emulator-facing local API.
3. Seed the idempotent login and flashcard fixture without printing its credentials:

   ```sh
   godotenv -f .env.development ./scripts/seed-maestro-fixture.sh
   ```

4. Start the development bundler with `npx expo start -c --dev-client --host lan`, boot the configured Android emulator, and install the development client if it is absent.

The flashcard flow cold-starts the development client and uses its `__DEV__` credential prefill. The fixture script reads the same two test variables, hashes the password locally, and passes only the hash and email into the local Postgres container. The flow does not embed credentials.

## Full local gate

Run every checked-in flow together after the final UI change:

```sh
maestro test .maestro
```

Android emulator success is not iOS, physical-device, TestFlight, Play internal-track, or store-purchase evidence. Those remain separate release gates in `docs/MOBILE_APP_REVAMP_EXECUTION_PLAN.md`.

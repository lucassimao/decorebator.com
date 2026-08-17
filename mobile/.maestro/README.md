# Local Maestro suite

This suite is local-development evidence only. It must never point at the production API or seed a production database.

## Prerequisites

The suite uses the same stabilized emulator profile as
`calculadora-price-sac/mobile`: a visible Android 14/API 34 Play Store x86_64
AVD named `Large_Phone_414x896`.

Create it once when it is not already listed by `emulator -list-avds`:

```sh
export ANDROID_HOME="${ANDROID_HOME:-$HOME/ProgrammingTools/Android}"

"$ANDROID_HOME/cmdline-tools/latest/bin/sdkmanager" \
  "platform-tools" \
  "emulator" \
  "platforms;android-34" \
  "system-images;android-34;google_apis_playstore;x86_64"

echo no | "$ANDROID_HOME/cmdline-tools/latest/bin/avdmanager" create avd \
  --name Large_Phone_414x896 \
  --package "system-images;android-34;google_apis_playstore;x86_64" \
  --device pixel_5
```

Start it visibly with `npm run avd:maestro:cold`. If the default
`swiftshader_indirect` renderer fails, retry with
`GPU_MODE=swiftshader npm run avd:maestro:cold`; do not use `-no-window` for
layout-sensitive flows.

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

4. Build the x86_64 development client with
   `npm run build:android:maestro`, start the AVD with
   `npm run avd:maestro:cold`, and run `npm run avd:maestro:prepare`. The
   prepare script requires exactly one emulator target, waits for boot,
   disables animations, installs the APK, isolates the dedicated AVD's image
   MediaStore, and adds the single deterministic gallery fixture. It refuses
   to operate on a physical Android device.
5. Start Metro with
   `npx expo start -c --dev-client --host localhost --port 8081`.

The flashcard flow cold-starts the development client and uses its `__DEV__` credential prefill. The fixture script reads the same two test variables, hashes the password locally, and passes only the hash and email into the local Postgres container. The flow does not embed credentials.

The signup flow generates a timestamped local-only address on every run so its
three-per-hour account bucket cannot make repeated local or CI runs flaky. It
asserts the generic mailbox-instructions response and never requires mailbox
access or creates a production account.

## Full local gate

Run every checked-in flow together after the final UI change:

```sh
maestro test .maestro
```

Run the profile-upload flow alone with:

```sh
npm run ui:maestro:profile-upload
```

That command captures the seeded user's profile URL before the flow, waits for
the upload mutation to settle, requires a fresh server-owned URL afterward,
and verifies the exact object exists in local MinIO.

Android emulator success is not iOS, physical-device, TestFlight, Play internal-track, or store-purchase evidence. Those remain separate release gates in `docs/MOBILE_APP_REVAMP_EXECUTION_PLAN.md`.

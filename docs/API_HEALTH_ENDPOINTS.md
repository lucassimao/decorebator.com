# API health endpoints

The API exposes two unauthenticated probes. Neither endpoint returns dependency errors, connection details, credentials, or application data.

| Endpoint | Success | Failure | Intended use |
|---|---:|---:|---|
| `GET /healthz` | `200 {"status":"ok"}` | Process/network failure only | Process liveness and restart decisions |
| `GET /readyz` | `200 {"status":"ready"}` | `503 {"status":"not_ready"}` | Traffic admission and deployment readiness |

`/healthz` is dependency-free. A database outage must not make an orchestrator restart an otherwise live process repeatedly.

`/readyz` performs a PostgreSQL pool ping with a one-second request-derived deadline. A failed, canceled, or timed-out ping returns the same fixed `503` response without exposing the database error.

Deployment configuration must use `/healthz` for liveness/restart probes and `/readyz` for readiness or load-balancer admission probes. The repository currently has no production API deployment manifest; `OPS-DEPLOY-1` owns adding and validating those platform-specific probe declarations before release.

# Google Play RTDN operations

This runbook describes the external Google Cloud configuration expected by the
API. It is an operator checklist, not an executable deployment script; no
application startup path creates or mutates cloud resources.

## Required topology

- Google Play publishes subscription notifications to
  `GOOGLE_PUBSUB_TOPIC`.
- `GOOGLE_PUBSUB_SUBSCRIPTION` is an authenticated push subscription targeting
  `GOOGLE_PUBSUB_PUSH_AUDIENCE` and uses
  `GOOGLE_PUBSUB_PUSH_SERVICE_ACCOUNT_EMAIL` as its OIDC identity.
- Its dead-letter policy forwards after
  `GOOGLE_PUBSUB_MAX_DELIVERY_ATTEMPTS` (5-100) to
  `GOOGLE_PUBSUB_DEAD_LETTER_TOPIC`.
- `GOOGLE_PUBSUB_DEAD_LETTER_SUBSCRIPTION` must be attached to the dead-letter
  topic. A topic without a subscription loses forwarded messages.

Google documents that dead-letter delivery attempts are approximate and are
counted only when the Pub/Sub service agent has the required publisher and
subscriber IAM grants. Follow the current official setup instructions before
enabling `STORE_IAP_ENABLED`:

- <https://cloud.google.com/pubsub/docs/dead-letter-topics>
- <https://cloud.google.com/pubsub/docs/push>

## Command template

Resolve and review every placeholder before running these commands in the
intended project. Use infrastructure-as-code or the Google Cloud console when
that is the project's normal change-control path.

```sh
gcloud pubsub topics create PLAY_RTDN_DEAD_LETTER_TOPIC
gcloud pubsub subscriptions create PLAY_RTDN_DEAD_LETTER_SUBSCRIPTION \
  --topic=PLAY_RTDN_DEAD_LETTER_TOPIC

gcloud pubsub subscriptions update PLAY_RTDN_PUSH_SUBSCRIPTION \
  --dead-letter-topic=PLAY_RTDN_DEAD_LETTER_TOPIC \
  --max-delivery-attempts=10
```

Then grant the project Pub/Sub service agent permission to publish to the
dead-letter topic and consume/acknowledge from the source subscription, exactly
as documented by Google. Also grant the configured push service account
permission to invoke the deployed HTTPS service. Do not grant these roles to
the Android Publisher credential unless it separately needs them.

## Verification and alerts

Before enabling IAP, verify in Google Cloud that:

1. The source subscription's topic, push endpoint, OIDC audience, and service
   account exactly match the API environment variables.
2. The dead-letter policy references the expected topic and delivery count.
3. The dead-letter topic has its own active subscription.
4. Pub/Sub service-agent IAM grants are present.
5. Monitoring alerts cover non-ack push responses,
   `subscription/dead_letter_message_count`, oldest unacked message age, and
   dead-letter subscription backlog.

The API returns `204` for terminal/poison deliveries, `401` for rejected OIDC
identity, and `503` for retryable provider or persistence failures. It also runs
a five-minute inbox-health check that reports overdue `retryable` rows and
expired `processing` leases without logging provider identifiers. Those alerts
do not replay events because raw provider payloads are intentionally not stored;
use Google redelivery/DLQ and an authoritative Play API reconciliation path.
Per-process bounded webhook counters (provider, disposition, outcome, result code,
duplicate flag, and capped delivery attempt only) are available through the
static-authenticated `GET /static/admin/store-webhooks/metrics` endpoint; use
the structured webhook logs or a deployment-level collector for fleet totals.

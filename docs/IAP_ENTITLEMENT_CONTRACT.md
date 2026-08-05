# Store-Neutral IAP Entitlement Contract

Status: Phase 1 domain contract; persistence and provider identity bindings are defined in later Phase 1 items.

## Authority and trust boundary

`StoreEntitlement` is the backend-owned record that decides premium access. Mobile purchase callbacks are evidence to verify, not authority to grant access. The backend must obtain and validate current state with Apple or Google before creating or changing an entitlement.

Unknown stores, products, entitlements, statuses, environments, or unverifiable provider responses fail closed and grant no access.

## Canonical fields

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `id` | integer | after persistence | Internal entitlement record identifier. |
| `userId` | positive integer | yes | Internal account that owns the entitlement after server-side identity binding. |
| `store` | `apple` or `google` | yes | First-party store that verified the purchase. |
| `productId` | non-empty string | yes | Exact store product identifier. It must resolve through the server-owned product catalog. |
| `entitlement` | non-empty string | yes | App capability granted by the product, initially `premium`. |
| `status` | canonical status | yes | Store-neutral lifecycle state defined below. |
| `periodStart` | timestamp or null | when supplied by the store | Start of the currently verified access period. It may be absent for a pending purchase. |
| `periodEnd` | timestamp or null | for `active` and `grace_period` | Exclusive end of the currently verified access period. |
| `autoRenewEnabled` | boolean | yes | Whether the store currently says the subscription will renew automatically. |
| `canceledAt` | timestamp or null | when known | Time renewal was canceled. Cancellation does not itself end an already-paid period. |
| `revokedAt` | timestamp or null | for `revoked` | Time access was refunded, revoked, or otherwise terminated by the store. |
| `revocationReason` | string or null | when known | Normalized or safely retained reason for revocation. The later write-path contract is responsible for allowlisting this value so secrets or full provider payloads are never stored here. |
| `environment` | `sandbox` or `production` | yes | Store environment used for verification. Google license testers and test purchases map to `sandbox`. |
| `lastVerifiedAt` | timestamp | yes | Backend time of the latest successful provider verification. It is not a client timestamp. |
| `createdAt` / `updatedAt` | timestamp | after persistence | Internal audit timestamps. |

Provider transaction identifiers, purchase tokens, app-account identifiers, notification identifiers, and raw signed payloads are deliberately outside this lifecycle record. Their binding, uniqueness, and retention rules belong to the next Phase 1 contract items.

## Store account and purchase identity binding

The account binding is created by the backend before purchase. The purchase binding is created only after provider verification. Both are internal domain records with JSON serialization disabled; a later mobile API uses narrow response/request DTOs rather than exposing these records.

### Apple binding

1. For an authenticated internal account, the backend creates or retrieves a stable, nonzero UUID as its `appAccountToken`. The purchase-context response returns that token but accepts no client-selected `userId`.
2. Mobile supplies the backend-issued UUID to StoreKit when starting the purchase.
3. The backend verifies Apple's signed transaction, bundle, product, environment, and signature chain before reading identity fields.
4. The verified transaction's `appAccountToken` must exactly match an existing `AppleAccountIdentity`. A missing, unknown, or mismatched token grants no entitlement.
5. The backend records an `ApplePurchaseIdentity` containing the resolved internal `userId`, verified `appAccountToken`, `originalTransactionId`, environment, and verification time. The original transaction identifier represents the subscription lineage across renewals and restores.
6. Server notifications resolve the user from the stored original transaction/app-account binding without a mobile client or a user ID in the notification request.

An Apple purchase is never reassigned merely because another user is currently signed in during restore. Since there are no legacy paying customers, there is no unbound-transaction claiming exception in this contract.

### Google Play binding

1. For an authenticated internal account, the backend creates or retrieves a stable, PII-free `obfuscatedExternalAccountId`, at most 64 characters. Generate it from server-side entropy or a one-way keyed derivation; never send an email, raw internal user ID, or another clear-text identifier to Google.
2. The purchase-context response returns that identifier but accepts no client-selected `userId`; mobile supplies it through `setObfuscatedAccountId` when launching BillingClient.
3. Mobile sends only the purchase token as evidence. The backend verifies that token through `purchases.subscriptionsv2.get` before reading subscription or account state.
4. The verified `obfuscatedExternalAccountId` must exactly match an existing `GoogleAccountIdentity`. A missing, unknown, or mismatched identifier grants no entitlement.
5. The backend records a `GooglePurchaseIdentity` containing the resolved internal `userId`, verified obfuscated account identifier, verified purchase token, environment, and verification time.
6. Real-time developer notifications resolve the user from the stored purchase-token binding without trusting notification/client account data. Linked-token replacement behavior is defined with verification operations in a later Phase 1 item.

The Google purchase token is sensitive backend material: never serialize or log it, and design persistence so the recoverable value is encrypted or equivalently protected because subsequent Play API calls require it. A digest may support lookup/uniqueness, but cannot replace the protected token used for provider verification.

### Shared binding invariants

- Account identities require a positive internal `userId` and a valid provider identifier.
- Purchase identities require the resolved user, exact account identifier, verified provider purchase identifier, valid environment, and nonzero backend verification time.
- Provider identifiers are opaque and compared exactly; leading/trailing whitespace is rejected rather than normalized.
- `MatchesAccount` fails closed if either record is malformed, the users differ, or the provider account identifiers differ.
- Provider identity records are not API response models. Dedicated DTOs may expose only the pre-purchase account identifier required by StoreKit/BillingClient.
- Atomic uniqueness and idempotency constraints are intentionally specified in the next Phase 1 item; this item defines the identity and trust relationship they must preserve.

## Provider idempotency keys and uniqueness

Provider delivery identity and purchase identity solve different problems and must not be conflated:

- a purchase/transaction key identifies a provider purchase record that may be refreshed or updated as its store state changes;
- a transport key suppresses redelivery of the exact same notification envelope;
- a semantic event key suppresses equivalent provider events that arrive in distinct transport envelopes;
- none of these keys decides whether an older event may overwrite newer entitlement state. Provider-version ordering is defined in the verification/ingestion operations item.

Record keys and Apple notification keys use `namespace:environment:sha256`; Google RTDN event keys use `namespace:sha256` because an RTDN contains no environment before the Play API refresh. SHA-256 is calculated over length-prefixed components. Length-prefixing prevents ambiguous concatenation, namespaces prevent a transaction ID from colliding with a notification ID, and environment separation protects the purchase records for which that value is already verified. Keys are at most 128 characters and safe to log. Hashing is not encryption: any raw Google purchase token retained for future Play API calls still requires protected storage and redacted logging.

The Go contract uses distinct `ProviderRecordKey` and `ProviderEventKey` types. Transaction/purchase constructors return only record keys, while notification/RTDN constructors return only event keys, preventing a mutable purchase key from being accidentally inserted into the event-deduplication path.

### Apple keys

| Record | Key inputs | Constraint and behavior |
| --- | --- | --- |
| Verified transaction | environment + `transactionId` | `UNIQUE (transaction_key)`. Repeated verification upserts the same transaction record; it must not blindly ignore a later verified revocation or other newer signed state for that transaction. |
| Subscription lineage | environment + `originalTransactionId` | `UNIQUE (store, environment, original_transaction_key)` on the Apple purchase binding so one subscription lineage cannot be attached to two users. |
| Notification delivery | environment + `notificationUUID` | `UNIQUE (idempotency_key)` in the provider-event inbox. A duplicate UUID returns the already-recorded result and performs no second entitlement transition. |

Apple assigns a unique transaction identifier to each purchase/restore/renewal transaction and explicitly documents `notificationUUID` as the value to use to ignore duplicate notifications. The raw signed payload is verified before any key is accepted.

### Google Play keys

| Record | Key inputs | Constraint and behavior |
| --- | --- | --- |
| Purchase binding | environment + verified purchase token | `UNIQUE (store, environment, purchase_token_key)`. A token cannot be bound to multiple users. Repeated verification refreshes the same purchase record because subscription state changes under a stable token. |
| RTDN transport delivery | Pub/Sub topic + `messageId` | `UNIQUE (idempotency_key)` in the provider-event inbox. Pub/Sub guarantees message IDs are unique within a topic; including the topic preserves that scope and allows the inbox row to be claimed before a Play API call. |
| RTDN semantic event | package name + purchase token + notification type + `eventTimeMillis` | `UNIQUE (semantic_key)`. This suppresses equivalent RTDN business events delivered under distinct Pub/Sub message IDs before provider refresh. |

Google event keys deliberately omit environment because the RTDN envelope does not contain it. The verified environment is attached to the purchase record only after `purchases.subscriptionsv2.get` returns. Google event keys are quota and duplicate-transition guards, not a source of truth: every accepted RTDN still triggers a Play Developer API refresh, and the later operation contract must compare verified provider state/version before changing entitlement. A purchase token key uniquely identifies the mutable purchase binding; it must never cause all later events for that token to be discarded.

### Atomic persistence contract

The later additive IAP schema must provide these database-enforced constraint shapes before provider handlers are enabled:

| Target | Unique columns/index expression |
| --- | --- |
| Apple account binding | `app_account_token` |
| Google account binding | `google_obfuscated_account_id` |
| Apple subscription binding | `(store, environment, original_transaction_key)` |
| Apple transaction record | `(store, environment, transaction_key)` |
| Google purchase binding | `(store, environment, purchase_token_key)` |
| Provider event inbox | `idempotency_key` |
| Provider semantic event | partial unique index on `semantic_key WHERE semantic_key IS NOT NULL` |

The current `subscription_events.external_event_id UNIQUE` table is not the target contract: it has no explicit store/environment namespace and existing code performs a separate existence check before insert. New provider ingestion must claim the inbox row with one atomic insert (`INSERT ... ON CONFLICT`) and apply the entitlement change in the same database transaction. The exact inbox state machine, retry result, and out-of-order comparison belong to the next server-operations item; this item defines the keys and constraints that operation must use.

## Server operation contract

All operations return a stable outcome/code and never infer entitlement from a client callback. `CanGrantAccess` is true only for an `applied` result whose verified status is `active` or `grace_period`; the persisted entitlement must still pass its exclusive period-end check.

| Outcome | API/worker meaning | Entitlement effect |
| --- | --- | --- |
| `applied` | A newer verified provider state was committed atomically. | Apply the full canonical state; `revoked` removes access immediately. |
| `unchanged` | Verified state is older than the current stored provider version or produces no transition. | No change; return current entitlement. |
| `pending` | Provider confirms an incomplete purchase. | Persist pending evidence if useful, grant no access, and allow later refresh. |
| `duplicate` | Event inbox already contains the delivery key. | The future handler retrieves and returns the prior result stored on that inbox row; the pure decision model emits only `duplicate_event` and never runs a second transition. |
| `retry` | Provider timeout, rate limit, temporary 5xx, or transient infrastructure failure. | No entitlement change; keep/mark inbox retryable with bounded backoff. |
| `rejected` | Invalid evidence/signature/app, account mismatch, or unknown product. | No entitlement change; store only safe audit metadata and a stable rejection code. |

Permanent result codes are `invalid_purchase`, `account_mismatch`, and `unknown_product`; transient provider failures use `retryable_provider_error`; duplicate and stale deliveries use `duplicate_event` and `stale_provider_event`. Unknown enum/status values are invalid evidence and fail closed.

### Verify purchase

1. Authenticate the app session and ignore any request-body `userId`.
2. Verify signed Apple evidence or the Google purchase token with the provider, including app/bundle/package, environment, product, and provider time.
3. Resolve the user through the server-issued account identity and reject a missing/mismatched binding.
4. Reject unknown products before entitlement mutation; product metadata from mobile is never authoritative.
5. In one database transaction, upsert the mutable purchase record, compare provider ordering data, apply the canonical entitlement if newer, and persist the operation result. Pending returns success-with-pending but never premium access.

### Restore and status refresh

Restore is provider refresh, not ownership claiming. Apple transactions and Google tokens must resolve through an existing exact account binding; the currently signed-in user cannot take over an unbound or differently bound purchase. Refresh all known purchase lineages for the account, apply each newer verified state atomically, and return the resulting current entitlement. Empty restore is a successful no-entitlement result; transient provider failures are retryable; invalid or mismatched evidence is rejected.

### Provider notification ingestion

1. Authenticate the Pub/Sub push envelope or verify Apple JWS before trusting payload fields.
2. Atomically claim the transport key. A completed duplicate returns success immediately; a retryable prior attempt may be reclaimed under a bounded lease.
3. Resolve the stored purchase binding, then query the provider source of truth rather than applying RTDN/notification type alone.
4. Compare provider occurrence/version data against the stored provider occurrence/version under the entitlement row lock; never compare provider time with server `lastVerifiedAt`. Older or equal state is `unchanged`; newer pending, active, grace, hold/pause/expiry, or revocation state is applied according to the lifecycle contract.
5. Commit inbox result and entitlement mutation together. A crash before commit leaves neither applied; a crash after commit is a duplicate on redelivery.

Invalid authentication/signatures receive a permanent rejection and no provider API call. Unknown products and account mismatches are permanently rejected, audited without raw tokens/payloads, and surfaced to monitoring. Transient provider/API/database failures remain retryable and must not be acknowledged as successfully applied. Structured logs must use record/event digests and stable result codes; identity structs or raw Google tokens must never be formatted into logs.

## Canonical statuses and access

| Status | Grants premium access? | Meaning |
| --- | --- | --- |
| `pending` | no | Purchase exists but payment or store completion is not final. |
| `active` | yes, only before `periodEnd` | Store has verified a paid current period. |
| `grace_period` | yes, only before `periodEnd` | Store explicitly grants access while retrying renewal. |
| `on_hold` | no | Renewal payment failed and the store no longer grants grace access. |
| `paused` | no | Store has suspended access until a later resume. |
| `expired` | no | The access period ended or a pending purchase ended without entitlement. |
| `revoked` | no | Store refund, revocation, or voiding terminated access. `revokedAt` is required. |

There is intentionally no canonical `canceled` status. A user can turn off renewal while retaining access until the paid period ends. That state is represented as `active`, `autoRenewEnabled: false`, and `canceledAt` when the store supplies a cancellation time. Once the verified period ends, status becomes `expired`.

`periodEnd` is exclusive: access is valid only while `now < periodEnd`. A stale record whose period has elapsed grants no access even if its stored status still says `active` or `grace_period`.

## Provider status normalization

### Apple

| Apple subscription status | Canonical result |
| --- | --- |
| Active | `active` |
| Expired | `expired` |
| Billing retry | `on_hold` |
| Billing Grace Period | `grace_period` |
| Revoked | `revoked` |

Apple renewal status controls `autoRenewEnabled`. Turning renewal off does not override an otherwise active paid period. The signed transaction supplies product, purchase/expiry, revocation, and environment evidence; the backend records `lastVerifiedAt` only after signature and app/product validation succeed.

### Google Play

| Google subscription state | Canonical result |
| --- | --- |
| Pending | `pending` |
| Active | `active` |
| Paused | `paused` |
| In grace period | `grace_period` |
| On hold | `on_hold` |
| Canceled but not expired | `active` with `autoRenewEnabled: false` and cancellation metadata when available |
| Expired | `expired` |
| Pending purchase canceled | `expired`; any linked or replacement purchase is verified and represented independently |
| Voided, refunded with revocation, or developer-revoked | `revoked` |

Google test purchases map to the canonical `sandbox` environment. A purchase token is sent to the backend and verified with the Google Play Developer API; a client purchase state or order identifier never grants access by itself.

## Validation invariants

- `userId` is positive and is assigned only after server-side store identity binding.
- `store`, `status`, and `environment` must be recognized enum values.
- `productId`, `entitlement`, and `lastVerifiedAt` are required.
- If both dates exist, `periodEnd` cannot be before `periodStart`.
- `active` and `grace_period` require `periodEnd`.
- `revoked` requires `revokedAt`; `revokedAt` is invalid for every other status.
- Validation success does not itself grant access; `GrantsAccess(now)` also checks the status and unexpired period.

## Source references

- [Apple App Store Server API](https://developer.apple.com/documentation/appstoreserverapi/)
- [Apple app account token](https://developer.apple.com/documentation/appstoreserverapi/appaccounttoken)
- [Apple original transaction identifier](https://developer.apple.com/documentation/appstoreserverapi/originaltransactionid)
- [Apple transaction identifier](https://developer.apple.com/documentation/appstoreserverapi/transactionid)
- [Apple notification UUID](https://developer.apple.com/documentation/appstoreservernotifications/notificationuuid)
- [Apple subscription status](https://developer.apple.com/documentation/appstoreservernotifications/status?changes=_5)
- [Apple auto-renew status](https://developer.apple.com/documentation/appstoreserverapi/autorenewstatus?changes=_1%2C_1)
- [Google SubscriptionPurchaseV2](https://developers.google.com/android-publisher/api-ref/rest/v3/purchases.subscriptionsv2)
- [Google BillingFlowParams account identifier](https://developer.android.com/reference/com/android/billingclient/api/BillingFlowParams.Builder#setObfuscatedAccountId(java.lang.String))
- [Google real-time developer notifications](https://developer.android.com/google/play/billing/rtdn-reference)
- [Google Pub/Sub message identifiers](https://cloud.google.com/pubsub/docs/reference/rest/v1/PubsubMessage)
- [Google Play purchase verification and security](https://developer.android.com/google/play/billing/security)

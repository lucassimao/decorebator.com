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
- [Apple subscription status](https://developer.apple.com/documentation/appstoreservernotifications/status?changes=_5)
- [Apple auto-renew status](https://developer.apple.com/documentation/appstoreserverapi/autorenewstatus?changes=_1%2C_1)
- [Google SubscriptionPurchaseV2](https://developers.google.com/android-publisher/api-ref/rest/v3/purchases.subscriptionsv2)
- [Google Play purchase verification and security](https://developer.android.com/google/play/billing/security)

# IAP lifecycle test matrix

This matrix is the acceptance map for the Phase 2 store-entitlement test item. Provider calls use fixtures or fakes; persistence cases use the local PostgreSQL integration database. No live purchase or provider mutation is part of this evidence.

| Required case | Primary evidence | Boundary proved |
| --- | --- | --- |
| Grant | `TestAppleSubscriptionRefresherNormalizesEveryProviderStatus`, `TestGooglePurchaseVerifierAcceptsAccountBoundPurchase`, `TestEffectiveStoreAccessSeparatesEnvironmentAndRequiresGoogleAcknowledgement` | Only verified active or grace state grants, and Google additionally requires acknowledgement. |
| Renew | `TestStoreEntitlementRepeatedReceiptIsIdempotentAndRenewalAdvances`, `TestStoreEntitlementConcurrentUpgradeKeepsNewestSnapshotCurrent` | A newer provider snapshot extends the same persisted purchase while concurrent snapshots retain the newest state. |
| Cancel | `TestAppleSubscriptionRefresherPreservesProviderStatusAndHandlesCatalogTransitions`, `TestStoreEntitlementCancellationCyclesAndRevocationDominatesClockRaces` | Auto-renew cancellation preserves paid-period access and retains the first observation in each cancellation cycle. |
| Grace / pending | `TestAppleSubscriptionRefresherNormalizesEveryProviderStatus`, `TestGooglePurchaseVerifierMapsTestPendingCanceledAndExpiredStates`, `TestStoreEntitlementPendingRepurchaseCandidateCanActivateAfterRevocation`, `TestStoreIAPContextKeepsUnacknowledgedGooglePurchasePending` | Grace grants within the verified period; pending and unacknowledged purchases do not grant and can later activate. |
| Refund / revoke | `TestAppleSubscriptionRefresherRevocationDominatesStatusAndUsesCanonicalReason`, `TestGoogleRTDNVoidedSubscriptionRefreshesWithForwardCompatibleFields`, `TestGoogleVoidedPurchaseRevokesOnlyCurrentBindingAndHonorsEventClock` | Verified revocation dominates lifecycle state, and a Google void only revokes the exact current binding. |
| Restore | `TestStoreIAPRestoreDeduplicatesAppleLineageAndPreservesRejection`, `TestStoreIAPRestoreIncludesKnownBindingFailuresWithAndWithoutClientEvidence`, `TestStoreIAPGoogleRestoreRejectsProviderEnvironmentMismatch` | Restore is bounded, lineage-deduplicated, environment-pinned, non-claiming, and preserves mixed outcomes. |
| Invalid receipt | `TestApplePurchaseVerifierFailsClosedForInvalidProviderEvidence`, `TestApplePurchaseVerifierRejectsInvalidSignatureAndAuthenticatedAccount`, `TestGooglePurchaseVerifierFailsClosedForIdentityProductAndProviderErrors` | Invalid signatures, products, identities, environments, and malformed evidence fail closed before access. |
| Duplicate receipt | `TestStoreEntitlementRepeatedReceiptIsIdempotentAndRenewalAdvances`, `TestStoreEntitlementRejectsOwnershipTransferWithoutPartialWrite` | Repeating the same verified evidence is an unchanged no-op with one entitlement and one binding; ownership cannot transfer. |
| Duplicate notification | `TestAppleNotificationIngestorExportsTypedDuplicateDispositions`, `TestGoogleRTDNSeparatesVoidsAndRedeliveryWithoutExtraRefresh`, provider-inbox integration tests | Terminal redelivery is acknowledged without another provider refresh; nonterminal redelivery remains retryable. |
| Out-of-order notification | `TestDecideEntitlementOperationHandlesRequiredOutcomes`, `TestStoreEntitlementHistoricalGoogleTokenCannotOverwriteCurrentLifecycleOrBindingMetadata`, `TestStoreEntitlementCancellationCyclesAndRevocationDominatesClockRaces` | Older/equal events are unchanged and restrictive revocation cannot be undone by a later-arriving nonterminal snapshot. |

## Coverage instrumentation

The unit tests are in the external package `internal/tests/unit`, so every unit coverage command must pass `-coverpkg=./internal/...` to instrument application statements. The same rule applies to the external integration package. Aggregate coverage remains repository-wide and must not be relabeled as IAP-only coverage.

The corrected unit command currently reports 19.7% aggregate statement coverage. This is real baseline debt below the unchanged 70% threshold; it replaces the misleading 0% `[no statements]` profile but does not make the gate pass. Until repository-wide tests raise the aggregate to 70% or the owner explicitly authorizes a different coverage policy, the Phase 2 item remains blocked and downstream Gate 2 cannot be claimed.

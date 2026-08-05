# Native IAP paywall decision

Status: **approved for implementation**

## Candidate direction

Use the warm editorial **open study desk** paywall in `index.html`. It extends the accepted primitive system instead of preserving the current blue gradient/RevenueCat visual island. A code-native three-card shelf and bookmark provide the premium motif without adding a raster dependency.

## Contract represented

- The backend allowlist and current native store response are intersected before a plan renders.
- Plan title, display price, billing period, and eligible offer phases come from the native store product; required renewal prose is localized app copy filled only from those values.
- No plan is preselected. Plan selection is an explicit radio action, not an immediate purchase; only then does one stable primary button become available to open the native confirmation flow.
- Missing products never receive a hard-coded price, offer, or savings fallback.
- Purchasing locks the explicitly selected product and its disclosure. Pending and entitled states lock the server-returned product (represented by annual in the preview) so no card interaction can relabel an in-flight or owned entitlement.
- Loading, unavailable, empty, purchasing, pending, and already-entitled previews expose the required spatial/state pressure before implementation is split across roadmap items.

## Pre-review evidence

- Desktop, 390px compact, selected-plan, and dark pending specimens render with no horizontal overflow; `specimen-selected.png` captures the enabled action, radio state, offer, and paired disclosure.
- The ready state begins with zero selected plans and a disabled `Choose a plan` action; selecting a card is the only path that enables purchase continuation.
- The plan radio group uses one tab stop and wraps with arrow-key selection while keeping plan choice separate from the purchase action.
- The phone content scrolls (`902px` content in a `720px` viewport) without hiding the sticky action permanently.
- Light, dark, App Store, Google Play, annual/monthly selection, loading, unavailable, empty, purchasing, pending, and entitled states were exercised in-browser.
- App Store/Google Play and annual/monthly renewal disclosures remain paired with the selected store metadata after state changes.
- Restore remains a 44px target in unavailable and no-product states; an empty catalog disables purchase without hiding recovery, terms, or privacy actions.
- Initial store unavailability hides all commerce metadata, exposes Retry, and keeps restore/legal actions; it never presents stale-looking prices as current.
- Axe Core 4.10.3 reports zero automatic WCAG A/AA violations and 19 passes; only gradient/pseudo-element-dependent contrast remains a manual-review item.
- Manual contrast checks for the in-app ink, muted text, primary action, offer text, disabled action, and full-opacity focus indicators pass across light and dark themes; text pairs range from 5.42:1 to 15.16:1 and focus indicators exceed 3:1.
- Reduced-motion media rules remove stagger, transforms, and timing while keeping final state labels visible.

## Image-generation decision

No generated raster is used in the candidate. The shelf motif has semantic ties to word cards, adapts to light/dark themes, scales cleanly, and maps directly to Reanimated views. Claude should explicitly challenge this choice; image generation remains available if an illustration would improve comprehension enough to justify asset, localization, dark-mode, and motion costs.

## Rejected alternatives

- Keep the existing RevenueCat gradient paywall: rejected because it preserves provider-shaped UI, hard-coded savings, and a visual island that cannot represent the server/store authority split.
- Preselect annual as a default: rejected because the user should make an explicit plan choice before the purchase action becomes available.
- Launch purchase directly from a plan card: rejected because selection and native confirmation are separate decisions and need distinct accessibility/analytics semantics.
- Generate a decorative raster hero now: provisionally rejected because the code-native study-card shelf already explains the benefit, themes without duplicate assets, and has a direct native motion mapping; Claude's review may reopen this if visual comprehension is materially weak.

## Expected production surface

- Add a typed mobile API boundary for `/subscription/iap/context` that preserves explicit null states and intersects backend-allowlisted IDs with the current `expo-iap` subscription response.
- Add a focused native-IAP paywall component/hook using shared UI primitives and Reanimated 4, with store metadata adapters and unit-testable selection/loading/error behavior.
- Replace the provider chooser, Stripe checkout mutation, hard-coded price cards, and RevenueCat modal wiring in `mobile/app/settings.tsx` only after the dependent purchase-state item makes the new action safe end to end. No user-reachable build may expose either a dead new CTA or a CTA that calls `requestPurchase` without backend verify followed by backend-authorized `finishTransaction`; Settings keeps the old path until that safe state-machine tranche lands.
- Add localized paywall, disclosure, retry, empty, and accessibility copy across the existing eight locale catalogs plus focused component/API tests; purchase verification, pending, restore, foreground refresh, and analytics remain owned by their explicitly following roadmap items unless implemented and reviewed in the same safe state-machine tranche.
- Loading placeholders must be non-interactive and expose no hidden real price text to assistive technology; selection disclosure changes need an accessible association/announcement, and scaled-text tests must cover the legal block plus title/price grid.

## Review rounds

- Round 1: Claude `fable` retained the direction and no-image decision but rejected the low-contrast focus ring, mutable plan choice during purchasing/pending/entitled states, and unavailable-without-retry/stale-price ambiguity. All three blockers were corrected in the prototype and contract for same-thread reconciliation.
- Round 2: Claude `fable` re-exercised the revised browser artifact, measured passing focus contrast, verified locked annual disclosure through purchasing and Google pending states, verified unavailable hides commerce while preserving retry/restore, and returned `APPROVED` on 2026-08-05 with no remaining material blocker.

## Final decision

Implement the warm editorial open-study-desk direction with code-native views and Reanimated 4; do not add a generated raster. Preserve explicit no-default selection, store-authored commerce metadata, the locked product/disclosure policy, recovery access in every catalog state, reduced-motion semantics, and the safe-tranche gate above. Prototype consensus is complete; production implementation still requires its own tests, static/device evidence, and Claude diff review before the roadmap item can close.

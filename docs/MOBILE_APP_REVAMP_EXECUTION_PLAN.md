# Mobile App Revamp — Execution Plan

Status: implementation in progress; Phase 0.5 implementation readiness complete, final-release measurement owner-gated, Phase 1 next

This file is the execution tracker for `MOBILE_APP_REVAMP_STRATEGY.md`. It incorporates the repository audit and an independent Claude review/reconciliation pass completed on 2026-08-04.

## Product decisions and non-negotiable scope

- [ ] Remove Stripe and RevenueCat from the product and repository.
- [ ] Support paid entitlements only through Apple App Store and Google Play IAP.
- [ ] Do not migrate, grandfather, refund, or preserve existing real paid subscriptions. The owner confirms that the active records found in production are store-approval test accounts, not paying customers; those test accounts must remain covered by the store-test workflow and must not be treated as customer migration data.
- [ ] Disable web premium checkout when Stripe is removed. The web may consume an account entitlement granted by a mobile store purchase, but it must not initiate a web payment.
- [ ] Do not make unrelated revamp work a prerequisite for payment cutover unless it is a security or correctness blocker.

If the web is expected to keep selling premium, stop and change the first decision: that would require retaining a web payment provider and is incompatible with “remove Stripe completely.”

## Review outcome

The strategy is directionally feasible, but its original ordering is not executable. The critical correction is to build and prove native IAP before deleting the current providers. Claude agreed with the audit and added these gates:

1. “No paying users” must cover active, past-due, grace-period, comped/manual/lifetime entitlements, and any web premium state—not only a count of currently paying rows.
2. Web monetization must be an explicit product behavior, not an accidental consequence of deleting Stripe.
3. Apple and Google notification handling must be idempotent from its first implementation.
4. Sandbox/internal-track tests do not prove production notification configuration. Prefer one refundable production smoke transaction per store; an owner-approved waiver must record the remaining risk.
5. Provider schema/data cleanup and temporary provider-code rollback safety are separate concerns. With zero entitlements, the schema can be cleanly replaced after the zero-state gate; old provider code may remain disabled behind a short-lived rollback flag until the first IAP release is stable, then must be deleted.

## Evidence baseline

The following repository facts drive the order of work:

- Mobile startup currently routes unauthenticated users to sign-in: `mobile/app/index.tsx:30-47`.
- The mobile app still has Expo SDK 54, native `ios/` and `android/` projects, `react-native-purchases`, and no direct native IAP package in `mobile/package.json`.
- RevenueCat setup, restore, offering selection, and provider routing are concentrated in `mobile/hooks/useRevenueCat.ts:102-268`; settings consumes that flow in `mobile/app/settings.tsx:698-789`.
- The backend provider model is still Stripe/RevenueCat-shaped in `api/internal/model/subscription.go:70-147`; migrations `000049`, `000050`, and `000056` do not create Apple/Google providers.
- `api/internal/service/subscription.go:35-70` still requires Stripe configuration during service construction.
- Quiz answer saving returns `204 NoContent` without a session summary in `api/internal/http/quiz.go:127-180`.
- Progress summaries do not expose due counts in `api/internal/http/analytics.go:461-498`; due counts currently belong to reminder queries/payloads in `api/internal/repository/push_notifications.go:95-155` and `api/internal/service/push_notification_service.go:185-197`.
- `posthog.identify` and sign-in events already exist (`mobile/components/dashboard/Header.tsx:94-104`, `mobile/app/signin.tsx:75-81`), but the existing identity/funnel instrumentation includes PII and lacks the required failure coverage.

## Progress legend

- `[ ]` not started
- `[-]` in progress
- `[x]` complete with evidence linked in the notes
- `[b]` blocked; record the owner and decision needed

## Development protocol

This protocol is adapted from the completed `calculadora-price-sac/mobile/FEATURE_ROADMAP.md` workflow. It is the required workflow for each implementation item in this plan. Work phases and dependencies remain strictly ordered; once an item passes its gates, the next unlocked item starts automatically without waiting for another user message.

Every Claude review starts with the model order `fable` → `opus` → `sonnet`. Try `fable` first on every new prototype or implementation review, even if it was unavailable during an earlier review; fall back to `opus` only when that attempt is temporarily unavailable, then to `sonnet` only when `opus` is also temporarily unavailable. A model's substantive review findings are not an availability failure and must be reconciled with that model rather than triggering fallback. Record the model that completed each review in the item evidence.

For every item, in order:

1. **Create a UI/UX prototype when needed.** Before implementing any UI-bearing item, write a short design brief from the acceptance criteria and build one or more disposable HTML/React mockups under `docs/mockups/<item-id>/`. Keep prototypes isolated from production dependencies and use real copy, representative data, and mobile viewports.
2. **Iterate the prototype with Claude.** Review the mockup for hierarchy, interaction states, accessibility, content, responsive behavior, and consistency with the existing app. Revise until Codex and Claude agree there is no material objection against the defined checklist, for up to three review/revision rounds. Record the chosen variant, rejected alternatives, and final rationale in `docs/mockups/<item-id>/DECISION.md`. No additional user sign-off is required for an ordinary in-scope design choice; unresolved disagreement after three rounds becomes `[b] BLOCKED — owner decision`. If Claude is unavailable or cannot authenticate, record `[b] BLOCKED — owner action: restore Claude CLI access and rerun the design review`; do not silently treat an unreviewed prototype as approved.
3. **Write tests first.** Add a failing unit, integration, contract, or component test that reproduces the bug or specifies the behavior. For store integrations, use provider fixtures/mocks and test redelivery, invalid data, and out-of-order events without live purchases. Never weaken or delete an existing test to make the change pass.
4. **Implement to green.** Keep the change within the item’s stated scope. Update API contracts, migration notes, and user-facing copy in the same item when they are part of its acceptance criteria.
5. **Ask Claude to review the implementation diff.** Run an adversarial review before marking the item complete, with the item ID, approved mockup decision, and this plan as context. Address every finding that is real; record disagreements and their evidence in the completion note. Re-request review after substantial rework.
6. **Run the complete applicable test suites.** At minimum, API changes run the relevant Go unit/integration tests and coverage target; mobile changes run `npm test`; web changes run the web test/build checks available in the repository. Run cross-package tests when a shared contract or package changes.
7. **Run UI validation for UI changes.** Use Maestro on Android for affected mobile flows, with a scoped flow plus a fixed smoke subset when app-open behavior or shared surfaces change. Add or update the flow in the same item. Do not claim iOS/store validation from Android evidence. A single full Maestro suite is the final gate after the last UI item, before release-owner actions; fix regressions and rerun affected flows to green.
8. **Run static checks immediately before commit.** Use the current repository commands: API `make format-check`, `make lint`, and relevant tests; mobile `npm run typecheck`, `npm run lint`, and formatter checks; web `npx eslint .`, `npm run format:check`, and `npm run build` when applicable. Re-check generated migrations and `git diff --check`.
9. **Commit locally, one commit per item.** Use a concise imperative subject and include the roadmap checkbox/completion note in that commit. Do not push, open a PR, publish a release, change production data, or submit store builds unless the user explicitly authorizes that separate action.
10. **Close the item only after steps 1–9 pass.** Change its checkbox to `[x]` and append a dated evidence note listing tests, static checks, prototype/Claude decisions, implementation review, and any remaining owner gate. Then proceed immediately to the next unlocked item.

### Autonomous progression rule

The agent may continue through all locally executable milestones without pausing for user confirmation, including entitlement design, provider code, sandbox fixtures, and reversible local schema/code changes. It must pause before any production migration application—additive or destructive—because live schema changes can affect availability and workers, as well as before production data mutation, store configuration/submission, release publication, or any other irreversible external action. It must also pause if either required reviewer is unavailable for a consensus-gated UI decision, for a missing credential, or for a contradiction defined as conflict with an explicit user decision, this plan, repository instructions, or an accepted `DECISION.md`—not a stylistic disagreement. A normal design choice within the roadmap is resolved through the Codex/Claude prototype loop and recorded rather than escalated.

### Pending and blocker protocol

- **Pending (`[ ]`)** means the item has not started or is waiting on an earlier dependency; it is not an incident.
- **In progress (`[-]`)** means implementation or validation is actively underway.
- **Blocked (`[b]`)** means a specific external state, contradiction, approval, or unavailable required tool prevents completion; it must include the date, evidence already produced, exact owner action, and the condition that will make the item runnable again.
- Autonomous work is bounded at one milestone item per lane: no bundling multiple roadmap items into one commit, and no starting the next item until the current item has its tests, reviews, validation, and evidence note complete. This milestone cadence replaces routine user check-ins.
- When a blocker appears, complete every automatable part first: tests, local implementation, mockups, migrations drafted but unapplied, reports, and handoff instructions. Do not fabricate the missing live evidence or weaken acceptance criteria.
- A blocked sub-item does not stop independent work. Continue the next milestone whose dependencies are satisfied; pause only the blocked dependency branch and any items that genuinely depend on it.
- When external state changes, re-check it read-only, rerun the previously blocked validation, obtain the required Claude/design review, and replace `[b]` with `[x]` only after the normal completion gates pass.
- New findings are recorded in the relevant workstream and triaged as launch-blocking, fix-before-release, or deferred; they do not silently expand the current item.
- Business classification decisions—such as whether an entitlement is genuinely a customer or an approved test account—route to Lucas; Codex and Claude may collect evidence but must not classify ambiguous records autonomously.
- The status report should always state: current milestone, completed evidence, active blockers, independent work continuing, exact owner actions, and the next automatic step after unblocking.

### UI/UX prototype acceptance

- Prototype screens cover the happy path plus loading, empty, error, disabled, pending, success, and back/navigation states relevant to the item.
- Prototypes are checked at 390×844 and 430×932 phone widths, must meet WCAG 2.2 AA contrast and applicable keyboard/screen-reader semantics, and show the actual copy hierarchy before production implementation; a failed applicable check blocks approval unless an explicit `[b]` owner exception is recorded.
- HTML/React prototypes run an automated accessibility/contrast check with `axe` or an equivalent tool, and the tool/version/result is recorded; visual review alone is not sufficient evidence.
- `docs/mockups/<item-id>/DECISION.md` includes the selected variant, rejected alternatives, Claude review outcome, unresolved limitations, acceptance states checked, and the production files expected to change.
- Prototype approval does not replace Expo implementation tests or Maestro validation; it is the design gate that makes those implementation decisions explicit.

### Animation protocol

Animations are reviewed as behavior, not inferred from screenshots. Any item introducing motion must include an animation spec in its `DECISION.md` covering:

- trigger and user purpose: feedback, navigation, loading, celebration, or state transition;
- duration, delay, easing, properties, repeat count, and interruption/cancellation behavior;
- reduced-motion behavior and a non-animated fallback that preserves comprehension and access to the action;
- loading/error/offline behavior, unmount cleanup, and whether motion can block input;
- performance budget and the chosen implementation: Reanimated for interactive, gesture-driven, spring, scroll, and interruptible motion; the existing native-driver `Animated` API only as a temporary compatibility path while a touched surface is migrated; and Lottie only for finite decorative illustrations that justify the asset cost. State whether each animation is new or a migration of existing `Animated` code, and do not leave both systems driving the same component.

For each animated UI prototype:

1. Implement the motion first in the HTML/React mockup with CSS transitions/keyframes or a small state machine to explore timing and hierarchy; capture a short screen recording or frame sequence at both phone widths.
2. If the motion is interactive, gesture-driven, spring-based, scroll-driven, interruptible, or platform-sensitive, build a small Expo/Reanimated spike before approval. CSS alone is not evidence that native motion will feel or perform correctly.
3. Ask Claude to review the prototype and, where applicable, the Expo spike for timing, hierarchy, motion meaning, interruption, reduced-motion fallback, and whether the animation is helping or distracting; revise within the same three-round consensus limit. If consensus is not reached after three rounds, stop the item at `[b] BLOCKED — owner decision` with the competing options and evidence; do not implement an unresolved motion choice autonomously.
4. Implement the approved behavior in Expo with deterministic state tests, reduced-motion tests, cleanup/unmount coverage, and no animation-dependent assertions in Maestro; Maestro should wait for the resulting accessible state instead.
5. Check the affected flow on the Android emulator for visible jank, input lock, layout shift, and completion-state correctness; this is not evidence of iOS parity. Before an interactive, gesture-driven, spring, or haptic animation ships, require a physical-device pass on both iOS and Android for gesture latency, jank, interruption, and haptic fallback. If the devices or credentials are unavailable, finish all automatable checks and mark only the release gate `[b]` with the exact owner action.

Animation acceptance requires that the action and its result remain understandable with motion disabled, repeated taps do not create duplicate transitions, interrupted animations settle into a valid state, and the implementation does not introduce an obvious frame-rate or startup regression. The Claude diff review and static validation must also identify the animation system used by each touched component and fail the item if one component is driven by both legacy `Animated` and Reanimated.

The motion toolbox is selected by behavior, not novelty:

- Reanimated plus Gesture Handler for native, interruptible transitions, drag/swipe gestures, springs, scroll-linked effects, and shared-element-like choreography. Because neither is currently a direct mobile dependency, the first spike that needs them must add `react-native-reanimated` and `react-native-gesture-handler`, configure the Reanimated Babel/plugin requirements, provide the gesture-handler root wrapper, and confirm Expo 54/RN 0.81.5 New Architecture compatibility before feature work proceeds.
- SVG or ordinary React Native views for lightweight programmatic shapes, loaders, progress, and feedback that should remain themeable and accessible.
- Lottie for finite, authored vector illustrations such as onboarding or completion moments; keep the asset local, bounded, cancellable, and replaceable by a static fallback. The app already has `lottie-react-native`, but each asset still needs a bundle-size and reduced-motion review.
- Image generation for visual exploration, custom static illustrations, texture/background concepts, and storyboard frames that help choose motion. Generated raster art is not treated as an interactive animation implementation or as a substitute for a tested native animation asset.
- Haptics and sound only as optional reinforcement of a visible state change, with permission/platform checks and a silent fallback.

### Blocked protocol

When an item needs owner-only access—production subscription data, App Store Connect, Google Play Console, Apple/Google credentials, a real device, store review, PostHog, or a paid external service—complete every automatable part first. Mark the item `[b] BLOCKED — owner action: <exact steps for Lucas>` and list the evidence already produced. Never fabricate live purchase, notification, store-review, production, or analytics evidence.

### Scope and worktree rules

- Keep the IAP critical path sequential: entitlement contract → Apple/Google implementation → store test gate → cutover → provider removal.
- Use parallel lanes only after Phase 0 and the Phase 0.5 implementation-readiness gate are complete, and only with an explicit disjoint write scope. The Calculadora site-track exception is not automatically applicable here.
- Before starting parallel work, record each lane’s owner, worktree/branch, allowed paths, dependency gate, and merge order in the milestone note. No two lanes may edit the same production file, migration, shared contract, or roadmap section.
- Lanes that need database migrations coordinate sequence numbers through one integration owner; lanes that share an API contract must merge the contract first and run integration tests before either lane is marked complete.
- A parallel lane is complete only after its changes are integrated into the execution branch and the combined validation passes; isolated worktree tests alone are insufficient.
- Preserve unrelated working-tree changes, including the existing untracked `audit.txt`; do not reset, clean, or overwrite them.
- Historical migration files remain immutable. Add forward migrations and record rollback/backup expectations rather than rewriting migration history.

### Parallelization decision and lane matrix

The plan is intentionally hybrid rather than fully sequential:

| Lane | Can start after | Scope | Must wait for |
|---|---|---|---|
| Critical path | Phase 0 → 0.5 implementation readiness | Entitlement contract, IAP backend/mobile implementation, store validation, cutover/removal | Each prior gate; never skip the store gate before provider removal |
| Activation UX | Phase 0.5 implementation readiness | Onboarding, starter content, word entry, quiz/session loop, streaks, due/practice UX, mockups and Maestro flows | Store contract only for purchase/paywall surfaces; shared auth/subscription files require coordination |
| Reliability | Phase 0.5 implementation readiness | Scoped correctness, security, timeout, lookup, worker, and idempotency fixes | Any shared repository/service file already owned by another lane |
| Web | Phase 0 | Web checkout-disabled behavior, entitlement display contract, web validation | Final entitlement API shape before consuming mobile-store entitlements |

Within Phase 2, Apple and Google provider implementations may proceed in parallel after the store-neutral contract is reviewed. Mobile purchase UI may proceed against mocked contract fixtures, but integration cannot be marked complete until the backend contract and provider verification tests pass.

The following remain sequential and cannot be parallelized: zero-entitlement classification, activation baseline definition, entitlement contract approval, production/store validation, live cutover, and Stripe/RevenueCat deletion.

### Final release gate

After all code items are complete, run the full applicable API, mobile, and web validation, the full Maestro suite, and the production-readiness checks in Phase 3. Owner-only store configuration, production smoke purchase, release submission, and live monitoring remain `[b]` until the owner performs them and records the evidence.

## Phase 0 — Scope and zero-entitlement gate

This phase is read-only except for documenting decisions. Nothing that drops provider code or schema may start before it passes.

- [ ] Query production subscription state for Stripe and RevenueCat: active, trialing, past_due, grace-period, cancelled-but-entitled, and any other state that grants access.
- [ ] Check for manual, comped, lifetime, promotional, or feature-flag entitlements outside the subscription tables.
- [ ] Check whether web users can currently hold premium state independently of those providers.
- [ ] Record the result as a dated, non-secret evidence note: `zero paid/entitled users confirmed`, or stop with `[!]` and owner direction.
- [ ] Define the single test-account classification used by both Phase 0 and Phase 0.5: only owner-confirmed store-approval accounts listed in the controlled test registry qualify; names, email patterns, or subscription status alone are not sufficient.
- [ ] Confirm web behavior: no web purchase flow; mobile-store entitlements may be honored on web if the shared account contract supports it.
- [ ] Replace the contradictory sequence in the strategy with this dependency order.

Gate 0 acceptance:

- A production query/evidence artifact proves zero real customer entitlements across all known sources, with store-approval test records explicitly classified and excluded from customer migration handling.
- If any entitlement cannot be classified under the shared test-account definition, all dependent work stops and the finding is presented for owner resolution.
- Product sign-off explicitly accepts disabling web checkout.
- Product IDs, pricing, entitlement name, renewal model, supported countries, and free-tier limits are listed in one contract.

Phase 0 status:

- [x] **2026-08-04 — zero real paying users confirmed; test-account exception recorded.** A read-only query through `api/.env.prod` connected as `doadmin` to `defaultdb` and found 6 subscription rows: 3 Stripe rows currently `active` (1 monthly and 2 annual, with period ends in 2026–2027) and 3 historical RevenueCat rows marked `cancelled`; the `users` table contains 3 active non-free users plus 1 cancelled annual user. The owner identified the active accounts as store-approval test accounts, not real customers. No user IDs, payment identifiers, or secrets were written to this plan, and no production data was changed. Customer migration/grandfathering is therefore out of scope. Before Phase 4 schema replacement, preserve or explicitly re-seed these approval accounts against the Apple/Google test workflow and verify that test records cannot be mistaken for real customer entitlements. The earlier Phase 0 block is cleared for Phase 1; the test-account lifecycle remains a cutover acceptance item.

## Phase 0.5 — Activation baseline and measurement gate

The production audit found that sign-ups are reaching wordlist creation but rarely reaching the learning loop. This phase makes that funnel measurable before the corresponding UX work is evaluated. It is not satisfied by total sign-up or payment counts, and the baseline is measurement-only: poor activation does not silently halt the roadmap, but it determines UX priority and the owner-approved release target.

Baseline captured on 2026-08-04, using read-only production data and excluding the four owner-confirmed store-approval accounts:

- 33 non-test users total.
- 26 created at least one wordlist; 14 added at least one word; 3 attempted at least one quiz.
- Last 30 days: 4 sign-ups, 3 wordlist creators, 2 users adding words, 0 quiz users.
- Last 90 days: 6 sign-ups, 5 wordlist creators, 4 users adding words, 0 quiz users.
- Last 365 days: 32 sign-ups, 26 wordlist creators, 14 users adding words, 3 quiz users.
- The database does not measure app opens, screen impressions, failed client flows, or session completion reliably; those require the event contract below.

Required work:

- [x] Keep a stable, non-production-secret registry/fixture for store-approval accounts so they are excluded from activation reporting without relying on ad-hoc user IDs.
- [x] Define canonical events and properties: `user_signed_up`, `wordlist_created`, `word_added`, `quiz_session_started`, `quiz_session_completed`, `quiz_answered`, and `practice_cta_opened`. Specify one event per semantic action, session identity, source/entry point, wordlist context, and failure/error outcome.
- [x] Audit existing mobile/PostHog events before adding duplicates. Correct the current mismatch where `quiz_completed` represents individual answers rather than completed sessions.
- [x] Add a development/dry-run sink and tests proving event names, required properties, deduplication, and exclusion of raw word/content data.
- [x] Produce the reproducible pre-release database cohort report for 7/30/90/365-day sign-ups with signup→wordlist, signup→word, and signup→database quiz-answer conversion. Keep its quiz-answer definition explicitly separate from completed sessions.
- [b] Produce the post-instrumentation production PostHog comparison for signup→wordlist, signup→word, signup→first quiz, and first quiz→completed session; keep the database baseline alongside it and do not compare unlike definitions.
- [b] Define owner-approved activation targets and the observation window before the final release gate. Do not invent a target percentage or declare success from a local test run.

Phase 0.5 implementation-readiness gate — complete; unblocks Phase 1 and independent lanes:

- The activation funnel can distinguish real users from store-approval accounts and can be queried by signup cohort and app version.
- Tests prove that a quiz answer does not count as a completed session and that retries/background duplicates do not inflate conversion.
- The event contract, controlled registry, migrated call sites, validation/dedupe tests, and reproducible pre-release database baseline are complete.
- PostHog React Native supplies `$app_version` from the installed Expo application metadata, so production cohorts can be segmented by app version without raw user/content properties.

Final-release measurement gate — open; blocks final release only:

- A production PostHog dashboard or reproducible export shows post-instrumentation values for all four funnel transitions alongside the non-equivalent database baseline.
- The owner has chosen the activation success threshold and observation window.

Phase 0.5 progress:

- **2026-08-04 — canonical activation contract defined.** Added the typed, privacy-safe event contract and dry-run capture boundary in `mobile/utils/activationEvents.ts`, with unit coverage for canonical names, scalar allowlisting, raw-content rejection, falsy values, and dry-run behavior. Documented session semantics and allowed properties in `mobile/docs/posthog-events.md`. The production PostHog comparison and owner target selection remain final-release measurement work.
- **2026-08-04 — controlled store-approval registry prepared.** Added `docs/fixtures/store-approval-accounts.json` with the four owner-confirmed internal user IDs from the read-only subscription audit, provenance, re-verification date, and usage rules in `docs/fixtures/README.md`. The registry artifact excludes accounts by explicit numeric ID and contains no email, payment, or secret data; production report consumption remains a separate pending item.
- **2026-08-04 — existing activation events audited and migrated.** Removed duplicate `signup_completed` and raw email PostHog properties, removed the raw wordlist name, added `word_added`, and replaced per-answer `quiz_completed` with session-aware `quiz_session_started`, `quiz_answered`, and `quiz_session_completed` events. Quiz counts are ref-backed for cleanup accuracy, and Sentry sign-in context no longer stores email. Targeted lint, formatting, and activation tests pass; full typecheck remains blocked by pre-existing WebRTC event-typing errors, and the full Jest suite retains unrelated Expo/mock/API-environment failures. No bulk word-add path was found; the production PostHog comparison remains final-release measurement work.
- **2026-08-04 — activation dry-run/validation sink completed.** `mobile/utils/activationEvents.ts` now enforces event-specific required properties, rejects invalid captures before sending, supports caller-owned process-scoped dedupe keys, and keeps dry-run captures from mutating real dedupe state. Seven unit tests cover names, required fields, privacy filtering, falsy values, dry-run behavior, dedupe, and session IDs. Production PostHog delivery remains intentionally non-durable; the production comparison remains a final-release measurement gate.
- **2026-08-04 — database cohort baseline automated; analytics comparison blocked.** Added the read-only `docs/queries/run-activation-cohort-report.sh` and SQL query, which loads the controlled registry and reports 7/30/90/365-day cohorts without hardcoded test IDs. The current run produced 7d `0/0/0/0`, 30d `4/3/2/0`, 90d `6/5/4/0`, and 365d `32/26/14/3` for signups/wordlists/words/database quiz answerers; the query labels database quiz answers as non-equivalent to completed sessions. `[b] BLOCKED — owner action: provide an authenticated production PostHog read/export credential or dashboard export, then rerun the post-instrumentation cohort comparison and record the observation window. The database-only baseline is complete; no production data was changed.`
- **2026-08-04 — Phase 0.5 gate split corrected after Claude reconciliation.** The implementation-readiness gate is met, so Phase 1 and independent UX/reliability work may proceed. `[b] BLOCKED — owner action before final release: provide an authenticated production PostHog export and choose the activation thresholds plus observation window. These owner actions block final release measurement sign-off only, not Phase 1.`

## Phase 1 — Design the store-neutral entitlement contract

Keep this contract independent of SDK names. Do not delete the current providers yet.

- [x] Define canonical subscription fields: account/user, store (`apple` or `google`), product ID, entitlement, status, start/end dates, cancellation/revocation state, environment, and last verified time.
- [x] Define Apple identity binding using an app-account identifier and original transaction identifier; define Google binding using a verified purchase token and an obfuscated account identifier. Never trust a client-supplied user ID without server binding.
- [x] Define event idempotency keys and unique constraints for Apple transaction/notification IDs and Google purchase tokens/event IDs.
- [x] Define server operations for purchase verification, restore/status refresh, and provider notification ingestion. Specify retry, pending, duplicate, invalid, revoked, and unknown-product behavior.
- [x] Define the mobile API response consumed by settings/paywall: products, current entitlement, pending state, error state, and restore result.
- [x] Select and spike an Expo-compatible native IAP implementation for SDK 54. Confirm that it works with the existing native projects and development-client workflow before committing to it.

Phase 1 progress:

- **2026-08-04 — canonical entitlement lifecycle complete.** Added the server-owned `StoreEntitlement` model and `docs/IAP_ENTITLEMENT_CONTRACT.md` with Apple/Google normalization, cancellation-versus-expiration semantics, fail-closed access rules, environment separation, and explicit deferral of provider identity/event persistence to the following items. Test-first evidence covers every enum, required fields, invalid periods/revocation, cancellation during a paid period, missing/expired/exact-boundary period ends, and access for each status. Claude `fable` identified deterministic pending-cancellation, fail-closed nil-period, and scope-separation issues; all were reconciled and the second `fable` review returned `APPROVED`. `make test-unit`, race-enabled unit tests, `make format-check`, changed-file lint, and `git diff --check` pass; targeted coverage reports 100% for every function in `entitlement.go`. The repository's aggregate model lint still reports only pre-existing findings in untouched `word.go`, `definition_image.go`, and legacy `subscription.go`, while the two changed Go files pass file-scoped `golangci-lint`.
- **2026-08-04 — store-to-user identity binding complete.** Added internal, non-serializable Apple/Google account and verified-purchase identity models plus the pre-purchase, verification, restore, and notification trust rules in `docs/IAP_ENTITLEMENT_CONTRACT.md`. Tests prove exact server-issued identifier matching, malformed/mismatched fail-closed behavior, Google’s 64-character limit, environment/verification requirements, and JSON non-exposure for all four identity records. Claude `fable` returned `APPROVED`; its non-blocking serialization-test suggestion was added, while structured log redaction is recorded for the later verification/operation item where logging is introduced. `make test-unit`, race-enabled unit tests, patch-scoped model lint, unit-package lint, `make format-check`, and `git diff --check` pass; targeted coverage reports 100% for every function in `entitlement_identity.go`.
- **2026-08-04 — provider idempotency and uniqueness contract complete.** Added length-prefixed SHA-256 key builders with type-separated mutable `ProviderRecordKey` and deduplicating `ProviderEventKey` results, plus the required account, transaction, purchase, transport, and semantic-event uniqueness shapes. Google RTDN event keys intentionally omit environment until the Play refresh supplies it; Apple notification UUID and topic-scoped Pub/Sub message IDs deduplicate transport while semantic keys suppress equivalent RTDN envelopes. Claude `fable` found the original record/event type conflation, unavailable pre-refresh Google environment, omitted Apple transaction constraint, and error-style issue; all were corrected and the second `fable` review returned `APPROVED`. `make test-unit`, race-enabled unit tests, patch-scoped model lint, unit-package lint, `make format-check`, and `git diff --check` pass; targeted coverage reports 100% for every function in `entitlement_idempotency.go`.
- **2026-08-04 — verification, restore, and notification operation contract complete.** Added stable applied/unchanged/pending/duplicate/retry/rejected outcomes, permanent and retryable result codes, fail-closed access gating, and atomic verify/restore/inbox workflows. Ordering compares provider occurrence/version only with stored provider occurrence/version; older and equal events are no-ops, unknown statuses are permanent invalid-purchase rejections, and newer revocations apply immediately. Claude `fable` found the original mixed-clock comparison, equal-version reapply, unknown-status retry ambiguity, and high-risk test gaps; all were corrected and the second `fable` review returned `APPROVED`. `make test-unit`, race-enabled unit tests, patch-scoped model lint, unit-package lint, `make format-check`, and `git diff --check` pass; targeted coverage reports 100% for every function in `entitlement_operation.go`.
- **2026-08-04 — mobile IAP envelope contract complete.** Added server-allowlisted single-store `premium` products, a narrow Apple/Google pre-purchase account context, current entitlement, pending, safe errors, restore outcomes, and server time, with explicit null-state JSON and no user IDs or provider purchase evidence. StoreKit/Google ProductDetails remain authoritative for localized price/title/offer data, while only backend entitlement state grants access. Claude `fable` found duplicate not-requested restore encodings, context/catalog nullability ambiguity, and permanent-error/pending conflicts; all were corrected and the second `fable` review returned `APPROVED`. `make test-unit`, race-enabled unit tests, patch-scoped model lint, unit-package lint, `make format-check`, and `git diff --check` pass; every function in `mobile_iap.go` has at least 86% targeted coverage.
- **2026-08-04 — Expo SDK 54 native IAP implementation selected and compiled.** Selected and pinned `expo-iap` 5.0.1 after comparing it with `react-native-iap` 16.0.2; the Expo Module fits the existing SDK 54/New Architecture development client without the latter's additional Nitro Modules runtime. Added a type-checked contract spike for product fetch, Apple/Google account-bound purchase, restore, and backend-gated transaction completion. An isolated `expo prebuild` generated both platforms, and Android `assembleDebug` compiled all 609 tasks and produced a 227 MB development-client APK with OpenIAP and Play Billing; iOS generated a 15.1 target with OpenIAP CocoaPods integration, while actual Xcode/store execution remains a Phase 2/3 gate. Claude `fable` independently checked the installed 5.0.1 API/types, security boundary, dependency/plugin fit, evidence claims, and Gate 1, then returned `APPROVED`; its optional synchronous-rejection and three-day Google acknowledgement notes were incorporated. Full evidence and reconsideration triggers are recorded in `docs/IAP_NATIVE_LIBRARY_SPIKE.md`.

Gate 1 acceptance — complete:

- Contract review confirms that the server, not the client, grants entitlement.
- Duplicate purchase and duplicate notification requests are safe and produce one entitlement transition.
- The selected library can be built in an Expo development client and has a path to App Store/Play test environments.

## Phase 2 — Implement Apple and Google IAP while old providers are untouched

### Backend

- [x] Implement Apple purchase verification using signed transaction data/JWS and the App Store Server API.
- [x] Implement Apple App Store Server Notifications V2 with production and sandbox environment handling, signature validation, replay protection, and atomic event idempotency.
- [x] Implement Google purchase verification through the Play Developer API.
- [x] Handle Google pending purchases and acknowledgement; do not grant access from an unverified client callback.
- [x] Implement Google Real-time Developer Notifications, Pub/Sub authentication, redelivery handling, and atomic event idempotency.
- [x] Implement authoritative Apple subscription-status refresh and normalization; transaction verification alone must never derive entitlement lifecycle state.
- [x] Persist canonical Apple/Google entitlement and store-identity bindings with atomic verified-event application, encrypted/protected provider evidence, account/token lookup for restore and notifications, and no dependency on Stripe/RevenueCat rows.
- [x] Add store-specific secrets/configuration with startup validation and safe error reporting; wire the RTDN HTTP endpoint with `http.MaxBytesReader`, explicit disposition-to-status mapping, structured rejection/retry logs and metrics, Pub/Sub dead-letter configuration, and a reaper or alert for stranded retryable inbox rows.
- [ ] Add unit and integration tests for grant, renew, cancel, grace/pending, refund/revoke, restore, invalid receipt, duplicate receipt, duplicate notification, and out-of-order notification; repair the API coverage command so external-package tests instrument application packages and the 70% gate measures real statements instead of the current 0% no-statements profile.

Phase 2 progress:

- **2026-08-04 — Apple transaction verification boundary complete.** Added a bounded App Store Server API `Get Transaction Info` client with five-minute ES256 authorization, production-first lookup, and sandbox fallback only for Apple's documented `4040010`; Apple retryable error codes remain retryable independently of HTTP status. Added a Go-native signed-transaction verifier matching Apple's published ES256/x5c profile: exactly three certificates, pinned Apple root, Apple leaf/intermediate OIDs, certificate validity at `signedDate`, and a raw P-256 signature check. The purchase verifier then enforces authenticated account binding, exact app-account UUID, bundle, environment, transaction ID, auto-renewable product type, and the server product allowlist before returning redacted transaction evidence. It deliberately does not derive entitlement status or auto-renew state from transaction dates; verified renewal JWS from Get All Subscription Statuses is required before the later atomic entitlement operation. Claude review began with `fable` and used the allowed temporary `opus` fallback; it found status overreach, Apple retry-code, malformed-ID, and adversarial-test gaps. All were corrected, and the final `fable` reconciliation returned `APPROVED`. Targeted race tests, changed-file lint, `go vet`, formatting, and diff checks pass; no route, persistence mutation, or production provider call is enabled by this item.
- **2026-08-05 — Apple App Store Server Notifications V2 verified ingestion complete.** Added outer V2 and nested transaction/renewal JWS verification with production/sandbox app binding, UUID replay keys derived only after outer verification, payload-type discrimination, product/lineage/environment checks, a five-minute future-skew guard, and bounded request size. Migration `000074` adds the provider-event inbox with explicit store/environment/result/state constraints, transport and semantic uniqueness, attempt counters, retry timestamps, and renewable leases. Ingestion now durably claims a short lease, performs provider refresh without holding a database transaction, and atomically commits only the local entitlement mutation with the final inbox result; verified-envelope nested failures are safely recorded as rejected, transient refreshes become retryable, and stale workers cannot complete after losing their lease. Review started with `fable` (temporarily unavailable at its usage limit), then Claude `opus` found seven blocking conflict, retry, transaction-duration, type-confusion, timestamp, schema, and concurrency-test gaps; all were corrected and the same `opus` reviewer returned `APPROVED`. `make test-unit`, focused race-enabled PostgreSQL integration tests including an eight-way concurrent claim, migration `74` down/up, changed-file lint, `go vet`, formatting, and diff checks pass. The later handler/worker item must nack retry outcomes, add its poison-event/dead-letter policy, and may move lease timing fully to database time; no route, live migration, or provider call is enabled here.
- **2026-08-05 — authoritative Apple subscription-status refresh complete.** Added the App Store Server API `Get All Subscription Statuses` client and normalization across active, expired, billing-retry, grace, and revoked states, with exact environment, app, account, product, transaction-lineage, signed-transaction, and signed-renewal binding. Provider status is preserved without clock-derived lifecycle rewriting; signed revocation evidence is restrictive, cancellation is represented as snapshot observation, unknown future statuses retry, optional future renewal products remain forward-compatible, and transaction verification alone still cannot grant lifecycle state. Review started with `fable` (temporarily unavailable at its usage limit), then Claude `opus` found eight product, lifecycle-clock, event-order, revocation, cancellation, forward-compatibility, and test blockers; all were corrected and the same `opus` reviewer returned `APPROVED`. Focused race tests, changed-file lint, `go vet`, formatting, and diff checks pass. Persistence must keep snapshot timestamps distinct from notification occurrence ordering, retain the first cancellation observation instead of moving it forward on every refresh, and alert on indefinitely retrying unknown future statuses; no route, persistence mutation, or production provider call is enabled here.
- **2026-08-05 — canonical store entitlement and identity persistence complete.** Migration `000076` adds provider-independent entitlement, account-identity, and purchase-binding tables with strict user/store/environment composite foreign keys, current-token uniqueness, separate snapshot/event cursors, and an explicit retained-revocation result. Recoverable provider identifiers use AES-256-GCM with immutable owner/type AAD and versioned keys; stable HMAC blind indexes support restore/notification lookup without plaintext storage, and bounded re-encryption plus old-key row counts gate key retirement. Direct purchase and inbox paths share one savepoint-protected atomic apply operation, typed ownership conflicts, restrictive lifecycle merge, first-observation cancellation cycles, superseded-token suppression, concurrent upgrade serialization, and pending post-revocation candidate recovery. Review started with `fable` (temporarily unavailable at its usage limit), then Claude `opus` found six ownership, supersession, revocation, transaction, stale-metadata, and constraint-classification blockers plus two residual historical/pending-repurchase traps; all were corrected and the same `opus` reviewer returned `APPROVED`. Unit/race tests, 10 focused PostgreSQL integration tests, provider-inbox result coverage, migration `76` down/up, build, vet, and formatting checks pass. The repository-wide integration harness remains independently broken: `docker-compose.test.yml` pins Go 1.23 while `go.mod` requires 1.25, and migration `000052` drops River tables while `CleanTestData` still truncates `river_job`; these test-infrastructure pendencies do not affect the focused persistence suite and must be repaired before relying on the one-command full-suite gate.
- **2026-08-05 — Google Play subscription verification complete.** Added a bounded `purchases.subscriptionsv2.get` client with bearer-token injection, exact v2 URL encoding, HTTPS-only remote transport, loopback-only HTTP test support, a request-level 15-second timeout independent of the injected client, a one-MiB response limit, and safe retry classification including Google's documented `409 concurrentUpdate`. Verification requires the authenticated user’s exact server-issued obfuscated account ID, one allowlisted premium line item with a valid auto-renewing/prepaid plan union, recognized subscription and acknowledgement states, the documented body `etag`, coherent provider timestamps, and test-purchase environment isolation. It normalizes pending, active, paused, grace, hold, canceled, expired, and pending-canceled states; tolerates five minutes of expiry clock skew; carries linked-token replacement as an explicit superseded/expired lineage; and never serializes raw purchase tokens. Review started with `fable` (temporarily unavailable at its usage limit); Claude `opus` found timeout/transport, expiry, cancellation-union, transient-body, and adversarial-test gaps, which were corrected. Its claim that `etag` was absent was reconciled against Google's current primary resource documentation, and `opus` withdrew that finding and returned `APPROVED`. `make test-unit`, focused race tests, changed-file lint, `go vet`, formatting, and diff checks pass. OAuth credential construction, acknowledgement mutation, persistence, and live Play calls remain in their explicit later items; no route or provider call is enabled here.
- **2026-08-05 — Google pending-purchase and acknowledgement processing complete.** Added a verify-first processor that never grants from a client callback: pending, expired, and revoked purchases are not acknowledged, while verified paid-lifecycle purchases use the exact Play acknowledgement endpoint and provider product/token evidence. Explicit acknowledgement outcomes preserve the verified purchase across retryable failures, distinguish local request errors from provider failures, and re-verify permanent failures to resolve idempotent acknowledgement races; unknown access-token failures fail safely as retryable unless a credential adapter classifies them permanent. The client applies bounded reads, request timeouts, URL escaping, bearer authentication, safe errors, and retry classification. Review restarted with `fable` (temporarily unavailable at its usage limit), then Claude `opus` found six blocking retry, error-classification, local-validation, token-source, failure-test, and wiring-boundary concerns. All implementation findings were corrected; the same reviewer accepted that route/config wiring and the mobile durable retry queue belong to their explicit later items and returned `APPROVED`. Race-enabled unit tests, changed-file lint, `go vet`, formatting, and diff checks pass; no route, credential, persistence mutation, or live Play call is enabled here.
- **2026-08-05 — authenticated Google RTDN ingestion complete.** Added Google-signed OIDC token verification with audience, issuer, service-account email, and verified-email binding; exact topic/subscription/package scoping; bounded Pub/Sub/base64/JSON parsing; transport and semantic replay keys; subscription, forward-compatible future subscription type, voided-subscription, test, unsupported-product, and poison-event handling; and a redacted event representation. Provider refresh runs outside the database transaction under a 20-second deadline inside the 30-second inbox lease. Permanent verification failures complete atomically as rejected, transient failures persist retry state, non-terminal duplicates explicitly require redelivery, and every failure exports an acknowledge/retry/unauthorized disposition for the later HTTP handler. Migration `000075` adds a distinct accepted test-event result and safely remaps it during rollback. Review restarted with `fable` (temporarily unavailable at its usage limit); Claude `opus` found eight blocking duplicate-redelivery, failure-classification, timeout, forward-compatibility, void/refund, rollback, poison-envelope, and dependency-tidiness issues. All were corrected, the stale type-list claim was reconciled against Google’s current primary documentation, and the same reviewer returned `APPROVED` twice after the final disposition API hardening. Full race-enabled unit tests, real PostgreSQL inbox integration tests, migration `75` down/up, changed-file lint, `go vet`, formatting, module tidiness, and diff checks pass. The next configuration/wiring item owns the HTTP body limiter and status mapping, credentials/startup validation, structured logs/metrics, Pub/Sub DLQ, and stranded-retry reaper or alert; no route, live migration, credential, or provider call is enabled here.
- **2026-08-05 — store IAP startup, webhook, and operations wiring complete.** Added production fail-closed typed configuration for Apple credentials and pinned roots, Google Application Default Credentials and exact Pub/Sub bindings, premium product catalogs, protected-evidence keyrings, and the expected dead-letter topology; disabled environments expose no store webhook routes. The composition root now wires Apple and Google clients, verifiers, authoritative refreshers, blind-index binding resolution, canonical persistence, and authenticated notification ingestors. Apple and Google endpoints apply `http.MaxBytesReader`, a 28-second request deadline, explicit terminal/authentication/retry status mapping, typed permanent-failure handling, bounded non-sensitive `error_kind` logs, capped delivery-attempt counters, and a static-authenticated per-process metrics view. Google void notifications revoke only exact current primary bindings, apply at equal cancellation timestamps, and cannot revoke historical or linked replacement tokens. A River health job alerts on both overdue retryable rows and expired processing leases without pretending to replay deliberately unstored payloads; `docs/GOOGLE_PLAY_RTDN_OPERATIONS.md` records the no-mutation DLQ/IAM/monitoring runbook. Review restarted with `fable` (temporarily unavailable at its usage limit); Claude `opus` found equal-timestamp refund and undiagnosable retry-log blockers, both were corrected, its three configuration notes were adopted, and the same `opus` thread returned `APPROVED`. The complete race-enabled unit suite, build, `go vet`, formatting, changed-file lint, diff checks, focused provider-inbox PostgreSQL tests, and equal/stale/historical/linked Google void integration tests pass. `make coverage-threshold` remains an independently broken repository gate: `make test-unit` succeeds but omits `-coverpkg` for the external `internal/tests/unit` package, producing a 0% no-statements profile instead of measuring application code; the next explicit test-matrix item owns repairing that instrumentation without lowering the 70% threshold. No live cloud resource, provider call, production migration, or external route configuration was changed; routes remain inactive until operators explicitly enable valid IAP configuration.

### Mobile

- [ ] Add the selected IAP native dependency and rebuild the development client.
- [ ] Replace the settings paywall purchase path with the store-product API; show store-localized price and legal text from the store/product contract.
- [ ] Implement purchase, pending, failure, restore, and “already entitled” states.
- [ ] Remove platform/locale routing that selects Stripe or RevenueCat.
- [ ] Ensure purchase state refreshes on app foreground and after notification/deep-link entry.
- [ ] Keep the old provider path disabled behind a temporary rollback flag only if needed for the first release; it must not be reachable in the IAP-only product behavior.

Gate 2 acceptance:

- Backend tests prove verification and idempotency without live provider calls.
- Mobile tests cover loading, success, pending, cancellation, failure, restore, and retry states.
- No new code grants premium from price/product metadata or a client-only success callback.

## Phase 3 — Store test matrix and production plumbing gate

- [ ] Apple sandbox/TestFlight: new purchase, restore, renewal, cancellation, grace-period behavior, refund/revocation, invalid transaction, duplicate notification, and out-of-order notification.
- [ ] Google Play internal track: new purchase, acknowledgement, pending purchase, restore/query, renewal, cancellation, refund/revocation, invalid token, duplicate RTDN, and out-of-order RTDN.
- [ ] Verify production-vs-sandbox Apple signing configuration and notification endpoint registration.
- [ ] Verify Google production Pub/Sub/RTDN topic and service-account configuration.
- [ ] Run a production no-entitlement query immediately before release.
- [ ] Prefer one self-funded, refundable production smoke purchase per store to validate production notification plumbing. If the owner waives this, record the waiver and the residual risk; the release is not equivalent to full production plumbing verification.
- [ ] Confirm store review metadata, privacy disclosures, subscription terms, restore-purchase affordance, and account deletion behavior.

Gate 3 acceptance:

- The complete matrix passes on both stores with captured transaction/event IDs and no secret values.
- Production notification configuration is proven, either by the refundable smoke transaction or by an explicit risk waiver.
- Production still has zero pre-existing entitlements before the IAP-only release.

## Phase 4 — IAP-only cutover and provider removal

- [ ] Ship the IAP-only mobile build through the required store tracks/review process.
- [ ] Confirm entitlement grant, refresh, revoke, and restore in the live release before removing old code.
- [ ] Disable web checkout and return a deliberate “purchases are available in the mobile apps” state rather than a server error.
- [ ] Replace the provider-shaped subscription schema with Apple/Google fields and constraints. Preserve migration history; do not rewrite old migration files.
- [ ] Remove Stripe/RevenueCat provider enums, fields, SDKs, environment variables, webhook workers, checkout routes, and dead tests/docs.
- [ ] Remove the temporary old-provider rollback flag after the first stable IAP release and successful monitoring window.
- [ ] Add post-cutover checks for entitlement counts, verification failures, notification lag, duplicate-event rate, acknowledgement failures, and support errors.

Gate 4 acceptance:

- `rg` finds no runtime Stripe or RevenueCat purchase/provider path, except historical migration documentation explicitly retained for migration history.
- Mobile, API, worker, and web checks pass with Stripe/RevenueCat environment variables absent.
- Web checkout attempts fail gracefully and cannot create a premium entitlement.
- A live Apple and Google entitlement can be refreshed and revoked without a provider-specific client SDK.

## Parallel workstream A — Revamp UX and analytics (not payment-blocking)

Track these independently unless Phase 0/1 identifies a direct contract dependency.

- [ ] Correct the analytics plan: retain `identify`/sign-in coverage, remove PII from user properties/events, define anonymous-to-authenticated identity handling, and add onboarding, paywall impression/selection, purchase pending/success/failure, quiz start/answer/session completion, notification open, and restore outcomes.
- [ ] Decide whether quiz completion should return a session summary. If yes, add a versioned API contract with box transitions, counts, next due count, and error semantics; otherwise document the client query/aggregation contract explicitly.
- [ ] Add due count to the intended progress-summary endpoint or remove it from the strategy’s requirement; do not infer it from reminder payloads.
- [ ] Build a small coherent primitive layer for the revamp (Button, Card, Input, Sheet, typography/spacing tokens) on top of existing `components/ui` and shared styles.
- [ ] Fix first-launch/auth routing, notification tap navigation, and flashcard end-of-list behavior as independently testable UX tasks.
- [ ] Validate accessibility, reduced motion, loading/error states, offline/retry behavior, and small-screen layouts.

## Parallel workstream B — Severity-driven reliability triage

These are not automatically part of the revamp. Fix them independently when evidence shows they affect correctness, security, or the launch path.

- [ ] Triage unscoped distractor and definition lookups first; verify language/wordlist/user boundaries.
- [ ] Make all Apple/Google and existing webhook idempotency atomic before launch.
- [ ] Validate the Stripe-config startup panic is gone as part of provider removal.
- [ ] Correct the TTS language/voice contract and add a regression test.
- [ ] Fix image-worker error-path formatting/strategy handling if still reachable.
- [ ] Add timeouts and cancellation to outbound OpenAI clients where production work depends on them.
- [ ] Reconcile documented and actual worker concurrency before increasing load.

## Explicitly removed from this revamp

- [ ] Existing subscriber migration, grandfathering, refunds, or billing-history preservation.
- [ ] Stripe fee optimization, RevenueCat entitlement migration, or current-payer churn work.
- [ ] Web checkout growth work that assumes Stripe remains.
- [ ] A large unrelated rewrite of content generation, worker topology, or analytics infrastructure without a launch-critical finding.

## Required validation commands

Run after implementation, from the relevant directory:

```text
cd api && make test && make lint && make format-check
cd mobile && npm run lint && npm run typecheck && npm test
cd web && npm run lint && npm run format:check && npm run build
```

Also run targeted tests for store verification, notification redelivery/idempotency, entitlement refresh, web checkout-disabled behavior, and absent Stripe/RevenueCat configuration.

## External implementation references

- [Apple App Store Server API](https://developer.apple.com/documentation/appstoreserverapi/)
- [Apple App Store Server Notifications](https://developer.apple.com/documentation/AppStoreServerNotifications/receiving-app-store-server-notifications)
- [Google Play Billing integration](https://developer.android.com/google/play/billing/integrate.html)
- [Google Play billing security](https://developer.android.com/google/play/billing/security)
- [Expo development builds](https://docs.expo.dev/develop/development-builds/use-development-builds/)

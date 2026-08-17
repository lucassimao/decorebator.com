# Decorebator Revamp Strategy — Top 25 Priorities

**Date:** 2026-08-03
**Scope:** Mobile app (`/mobile`) + backend (`/api`)
**Goal:** Increase install→paid conversion, user retention, and engagement through a coordinated feature + design revamp; simplify payments to first-party native IAP; harden and optimize the backend.

This document is the output of a deep audit across the mobile codebase (conversion funnel, retention & engagement mechanics, design system / UX quality) and the Go backend (payments stack, correctness, security, performance, AI cost). Items **#1–#20** cover the mobile revamp. Items **#21–#25** cover the backend: the migration off Stripe and RevenueCat to first-party native IAP, and the bugs, security fixes, and optimizations surfaced by the backend audit. Appendix C carries the full backend bug inventory.

**The headline:** the product core is genuinely strong — a deterministic Leitner SRS, an AI enrichment pipeline (definitions, images, TTS), 9 quiz modes, and a mature server-side push notification system. But the *growth layer* around that core is broken or missing. Most new installs never see onboarding. Every upsell dead-ends in the Settings screen. Quiz sessions have no ending. And not a single monetization event is tracked, so none of this is currently measurable. Several of the highest-impact fixes are small precisely because the backend already does the hard part.

---

## How to read this document

Each item includes:

- **The problem** — what the code actually does today, with file references.
- **Why it matters** — the case for the product owner: which metric it moves and why.
- **Goal & success metric** — what "done and working" looks like.
- **Engineering notes** — effort estimate and key touch points.

Items are sorted by priority: expected impact on conversion/retention/engagement, weighted by effort and by how many downstream items depend on them.

### Suggested sequencing

| Phase | Items | Theme |
|---|---|---|
| **Weeks 1–2** | #3, first-launch routing from #1, deep links from #7 | Measure + unblock: tiny changes, huge leverage |
| **Weeks 3–6** | #1 (social auth), #2, #10 | Conversion sprint |
| **Weeks 5–10** | #5, #6, #7, #4 | Retention sprint (overlaps) |
| **Continuous** | #8, #9 | Design revamp track running alongside |
| **Next wave** | #11–#15 | Pull in as capacity allows |
| **Fix-as-you-touch** | #16–#20 + Appendix A | Paid down in the same PR whenever a screen is redesigned under #8 |
| **Backend week 1** | #22 critical fixes, #23 quick wins | Broken endpoint, panic loops, JWT key validation, Sentry PII, graceful shutdown, health checks |
| **Backend weeks 2–4** | #24 indexes + timeouts, #23 hardening, #25 cost | Index-only perf wins first, then rate limiting/token lifetimes, then AI cost |
| **Payments migration** | #21 phases 0–4 per `MOBILE_APP_REVAMP_EXECUTION_PLAN.md` | Additive native IAP first; remove Stripe/RevenueCat only after Gate 3 store validation, with no customer backfill |

---

## Tier 1 — Order-of-magnitude opportunities

### 1. Fix the broken front door: first-launch routing + Apple/Google sign-in

**Category:** Conversion · **Effort:** S (routing) + M (social auth)

**The problem.** A brand-new install with no session is redirected to the **login screen**, not onboarding (`mobile/app/index.tsx:30-37` → `router.replace("/signin")`). There is no first-launch flag anywhere in the app. The entire 3-step onboarding — welcome value props, feature carousel, plan comparison — is only reachable via a small "Sign Up" text link in the corner of the sign-in screen, competing visually with "Forgot password?". The primary CTA on the first screen a new user ever sees is "Sign In" — an action they cannot perform. On top of that, there is **no social auth**: no `expo-apple-authentication`, no Google sign-in. Signup is a 3-field manual form whose full-name validation even requires a space (`mobile/app/signup.tsx:36-38`). The sign-in screen fires no analytics event at all, so the top-of-funnel bounce is completely invisible.

**Why it matters.** Every other investment in this document sits *behind* this door. Ad-driven and store-search installs are landing on a login wall and bouncing before the product has said a single word about itself. Adding Apple/Google sign-in is the single highest-known-lift change for mobile signup rates — one tap versus three fields and a keyboard — and Sign in with Apple becomes an App Store *requirement* the moment any other third-party login is added.

**Goal & success metric.** First launch routes to onboarding; returning users route to sign-in. One-tap Apple/Google signup on both platforms. Metric: install→signup rate (measurable once #3 lands); target is a step-change, not a percentage tweak.

**Engineering notes.** The routing fix is nearly a one-liner (`hasLaunchedBefore` in AsyncStorage checked in `app/index.tsx`). Social auth: `expo-apple-authentication` + `@react-native-google-signin/google-signin`, one new API endpoint for token exchange, account-linking rules for existing email users. Note `expo-auth-session` is already installed but used only for the Stripe redirect URI.

---

### 2. Put the paywall at the point of intent — and rebuild it

**Category:** Conversion · **Effort:** M

**The problem.** Every single upgrade trigger in the app routes to `/settings`:

- Hitting the wordlist limit → `mobile/app/dashboard/index.tsx:117`
- The limit dialog's "View Plans" button → `mobile/components/dashboard/UpgradePromptDialog.tsx:60-63`
- Premium feature gates (Speak, Stats) → `mobile/components/common/PremiumUpsellModal.tsx:96-106`

The user then lands at the *top* of a long settings scroll, must find the Upgrade section, and tap **another** button before prices appear. That is 3 taps + 1 scroll + a full context switch at the exact moment of peak intent. `RevenueCatPaywall` is imported in exactly one file: `app/settings.tsx`.

The paywall itself works against conversion:

- Price cards come first; the premium feature list is rendered **below the legal fine print** (`mobile/components/RevenueCatPaywall.tsx:296-441`) — likely below the fold on small devices.
- There is no CTA button — the price card *is* the buy button, with no selected state and no confirmation step.
- "Save 17%" is hardcoded (`RevenueCatPaywall.tsx:358-361`) and will silently lie if store pricing diverges. There is no per-month framing of the annual plan ("$5.83/mo, billed annually" — the single most effective annual framing).
- **No free trial exists, and the paywall cannot render one**: it reads only `priceString`, never `introPrice` or `discounts`. An orphaned i18n string `upgrade.startFreeTrial` (`mobile/i18n/locales/en.json:638`) is referenced nowhere — evidence of an abandoned plan.
- The header gradient is deep blue (`#1A237E → #64B5F6`) in an orange-brand app — the paywall looks like a different product.
- A load failure shows a bare "no offerings" error with no retry (`RevenueCatPaywall.tsx:233-257`). A transient StoreKit hiccup is a lost sale.

**Why it matters.** Paywall-at-trigger is the most reliable conversion lever in subscription apps: intent decays with every tap and every screen change between "I want this" and "here's the price". A 7-day free trial is the standard step-change lever for language-learning subscriptions. Both are currently structurally impossible.

**Goal & success metric.** Any trigger opens the paywall in place via a global `usePaywall()` provider (mirroring the existing `useUpgradePromptDialog` pattern in `mobile/hooks/useUpgradePromptDialog.tsx`). Paywall reordered: benefit → social proof → features → price anchor → CTA → legal. Trial configured in App Store Connect / Play Console and rendered from real `introPrice` data. Metrics: trigger→paywall-view rate (~100% by construction), paywall-view→purchase-start rate, trial-start rate.

**Engineering notes.** Provider + trigger rewiring is straightforward. Paywall rebuild is UI work inside one component. Savings percentage must be computed from real package prices. Retry + cached-fallback for offering load failures. See also #11 for the US iOS path this doesn't cover.

---

### 3. Instrument the funnel — we are flying blind on revenue

**Category:** Foundation (conversion + retention) · **Effort:** S–M

**The problem.** The app has ~12 PostHog events, and **not one touches money**. Missing entirely: `paywall_viewed`, `paywall_dismissed`, `upgrade_prompt_shown` (with trigger attribution), `limit_reached`, `purchase_started/completed/failed/cancelled`, `restore_*`, `feature_gate_hit`, push `permission_granted/denied`, `notification_opened`. There is no `posthog.identify()` call, so events are device-anonymous and cross-device cohorts fracture. `quiz_completed` (`mobile/app/quiz.tsx:228`) actually fires on every individual *answer*, not per session — "quizzes completed" is really "questions answered", and session length/drop-off are unmeasurable. Flashcards and voice chat are entirely uninstrumented. There is zero experimentation infrastructure: no PostHog feature flags anywhere, and `getCurrentOffering()` returns `offerings[0]` (`mobile/hooks/useRevenueCat.ts:210-215`) instead of `offerings.current`, which structurally blocks RevenueCat Experiments.

**Why it matters.** This is the cheapest item on the list and the one everything else depends on. Today we cannot answer: *What % of installs pay? Which upsell trigger converts best? Where in the paywall do people drop? Does a push notification ever produce a session?* Every change in items #1, #2, #4–#10 is unverifiable without this. Shipping this first means every subsequent release generates learning instead of guesses.

**Goal & success metric.** A complete event taxonomy covering the signup funnel, session semantics (`quiz_session_started/ended`), the full paywall/purchase funnel with trigger attribution, and the push funnel. `identify()` wired on auth. PostHog feature flags bootstrapped so paywall copy/price/trigger can be A/B tested. Success = the dashboards can answer the four questions above.

**Engineering notes.** A few days of work. Also fix while in there: duplicate `signup_completed` + `user_signed_up` events (`mobile/app/signup.tsx:140-146`), `onboarding_completed` firing on the "Sign in" path (`mobile/app/onboarding/account.tsx:67-70`), and raw email addresses sent as event properties on three auth events (PII — see Appendix A).

---

### 4. Collapse time-to-value: starter content + a taste of the product before commitment

**Category:** Conversion + Activation · **Effort:** M

**The problem.** The product's pitch is "our AI generates definitions, images, audio, and 8 quiz types from your words." A new user must survive all of this before seeing any of that magic: 9 swipeable feature slides (`mobile/app/onboarding/features.tsx` — telling, not doing), a 3-field signup, a `WelcomeOverlay` repeating the same three value props a *third* time, a **4-step** wordlist creation wizard (`mobile/components/dashboard/CreateWordlistModal.tsx`, `totalSteps = 4`), and then manual word-by-word entry with an async AI round-trip per word. There is **no sample content of any kind** — no starter wordlists, no templates, no import, no bulk add (grep for `sample|demo|starter|seed|template` returns nothing). There are also **zero personalization questions**: nothing asks the user's target language, level, or goal.

**Why it matters.** Time-to-value is the strongest known predictor of D1 retention, and ours is gated behind *the user doing data entry* before the product pays off. A single tap on "🇪🇸 Spanish Essentials — 10 words, ready to go" collapses four screens and ten form entries into one and puts AI-generated imagery on screen within seconds of signup. Personalization questions do double duty: they build commitment before the paywall (the Duolingo/Babbel playbook) and let paywall copy be personalized ("Learn Spanish 3× faster").

**Goal & success metric.** Curated starter wordlists per language, selectable during onboarding; onboarding cut to ~3 slides with one *interactive* beat (a real quiz card with AI image + audio the user can answer); 2–3 personalization questions (language, level, daily goal — the goal also feeds #6). Metrics: time-to-first-quiz, signup→first-quiz-completion rate, D1 retention.

**Engineering notes.** Starter packs are seed data + one API endpoint + a picker screen; the enrichment pipeline already exists. The interactive demo can run on bundled static content pre-auth (note: no product surface is currently reachable pre-auth, but a static demo card avoids the API entirely).

---

### 5. Give quiz sessions an ending: summary, celebration, and the natural upsell moment

**Category:** Engagement + Conversion · **Effort:** M

**The problem.** The quiz is an infinite treadmill. `handleNextQuiz()` (`mobile/app/quiz.tsx:299-309`) just fetches one more question forever; there is no session length, no round, no results screen, no closure of any kind. `quizCount`/`correctCount` are ephemeral local state, reset on unmount, never persisted or reported. The only exit is the back arrow. The "progress" bar is actually *accuracy* (`correctCount/quizCount` — `mobile/components/quiz/QuizProgressBar.tsx`), so it can read 0% after real effort and never conveys approaching completion. Flashcards are the same: reaching the last card silently no-ops (`mobile/app/flashcard.tsx:274`) — no "deck complete", no CTA into a quiz. The only thing in the entire app that reacts to a good session is the App Store review prompt (`mobile/hooks/useAppReview.ts` — fires on back-press after 10+ questions at 70%+ accuracy). We spend our best emotional moment asking for a favor.

**Why it matters.** Every retention loop needs a satisfying terminal beat — it's what makes a session feel *complete* and worth repeating tomorrow. The session-end screen is also the highest-converting surface in subscription learning apps: the user has just succeeded and felt the product work. Rotating that moment between celebration, streak reinforcement, a paywall (for free users at natural friction), and the review prompt turns one dead-end into four levers.

**Goal & success metric.** Sessions of ~10 questions ending in a summary screen: score, streak status, words that advanced between Leitner boxes, celebration animation, and a context-appropriate next action. Flashcard decks get a completion state with a "Quiz these words" CTA. Metrics: sessions per user per week, session completion rate, paywall views originating from session-end.

**Engineering notes.** Backend already returns everything needed per answer; session bookkeeping is client-side. Pairs naturally with #9 (celebration animation, haptics) and #3 (`quiz_session_started/ended` events).

---

### 6. Build a real streak + daily goal system

**Category:** Retention · **Effort:** M–L

**The problem.** What exists today: a **per-wordlist** streak computed by recursive CTE (`api/internal/repository/analytics/learning_progress.go:177-215`), which the dashboard fakes into a global number via `Math.max(...wordlists.map(wl => wl.currentStreak))` (`mobile/components/dashboard/Stats.tsx:105-110`) — a user practicing wordlist A on Monday and wordlist B on Tuesday shows a streak of 1, not 2. The streak is **only rendered when `currentStreak > 0`**, i.e. it is invisible to exactly the users who need the hook. The fuller streak stat sits behind the premium gate in analytics. There is no user-level streak column, no streak freeze/repair, no grace period, no streak-at-risk notification, no milestone celebrations, and no XP/levels/badges/achievements of any kind (grep confirms zero matches). There is no daily goal — despite orphaned i18n copy that *promises* one (`onboarding.notifications.goal`: "Goal-based reminders react to the daily minutes you picked" — no minutes picker exists).

**Why it matters.** The streak is the single most proven retention mechanic in this product category. Ours is mathematically wrong, hidden from new users, and half-paywalled. Emotional investment must *precede* the paywall: a user protecting a 12-day streak is a user who converts and who doesn't churn. A daily goal chosen by the user during onboarding (#4) is a commitment device that makes every reminder (#7, #14) feel like a service instead of spam.

**Goal & success metric.** User-level streak (`users.current_streak`/`longest_streak` or equivalent), updated on any practice; visible from day zero including at 0; freeze/repair mechanics; milestone celebrations at 3/7/30/100 days; daily goal set in onboarding and reflected on the dashboard. Metrics: D7/D30 retention, % of users with a 3+ day streak, streak-repair usage.

**Engineering notes.** Migration + streak service on the API; `last_practice_at` is already maintained (`api/internal/repository/user.go:171`). Client work is dashboard + session-end surfaces (#5). Streak-at-risk pushes belong to #14 but should be designed together.

---

### 7. Surface the SRS: due counts, a "Practice now" CTA, and notification deep links

**Category:** Retention · **Effort:** S–M

**The problem.** The backend computes exact per-user, per-wordlist due counts and even puts `wordlistId`, `dueCount`, and `type` into the push payload explicitly "for deep-linking" (`api/internal/service/push_notification_service.go:191-197`) — **but the app never displays a due count anywhere and has zero notification tap handlers** (no `addNotificationResponseReceivedListener` anywhere in mobile). Tapping a reminder dumps the user on the default route. The progress-summary API response (`api/internal/model/analytics_types.go:13-23`) has no `dueCount` field, so `WordlistItem` can only show a percentage bar. Meanwhile the dashboard's primary CTA is a **pulsing FAB to create another wordlist** (`mobile/app/dashboard/index.tsx:250-264`) — an action free users will be *rejected* for (#10). App icon badges are disabled (`shouldSetBadge: false` in `mobile/app/_layout.tsx:21-28`) and never set.

**Why it matters.** Spaced repetition only works if users show up when reviews are due — that's the entire pedagogical promise. Today the app knows exactly what's due and tells no one. The notification deep-link fix is the single highest-leverage, lowest-effort item in this entire document: it directly monetizes the push pipeline that's already built and running. Reorienting the dashboard from "create" to "practice" aligns the UI's primary action with the behavior that retains users.

**Goal & success metric.** `dueCount` added to `WordlistProgress`; "N due" badges on wordlist cards; a "Practice N words" primary dashboard CTA; notification taps deep-link into the quiz for the right wordlist; app icon badge shows due count. Metrics: push-open → session-start conversion (measurable after #3), DAU/WAU ratio.

**Engineering notes.** API change is one field. Deep-link handler is a small addition to `_layout.tsx` using the payload that already exists. Also revisit the very conservative 2-pushes-per-7-days rolling cap and the hardcoded 11:00 slot (`api/internal/repository/push_notifications.go`) once user-chosen reminder times exist (#14).

---

## Tier 2 — The revamp foundation

### 8. Extract a real design system + bottom tab bar

**Category:** Design ("new design" foundation) · **Effort:** L (incremental)

**The problem.** There is no component library and no shared primitives — no `Button`, `Card`, `Input`, `Sheet`, or `Badge` component exists anywhere. The result: ~7,850 lines of hand-rolled `StyleSheet` (~32% of all TSX), 11 separate `backButton` style definitions, 33 distinct card style keys, **six** competing header implementations, and 13 bespoke `<Modal>`s with inconsistent animations. Theme tokens exist (`mobile/contexts/ThemeContext.tsx`) but adoption is weak: 266 hardcoded `fontSize` values (there is no `theme.typography` at all), 611 hardcoded spacing values, 162 hardcoded hex colors, and **four overlapping token sources of which two are fully dead code** (`mobile/theme/colors.ts`, `mobile/theme/webColors.ts` — zero imports) and one is a copy-paste duplicate (`mobile/theme/authTheme.ts`). Dark mode is implemented via **119 hand-written `theme.mode === "light" ? X : Y` ternaries** across 20 files, with real holes: the entire auth flow forces light mode, `ErrorState.tsx` and the dashboard background are light-only, and the splash is white.

Two findings deserve special weight:

- **The brand orange `#FF7B54` fails WCAG AA at 2.56:1 with white text — on every primary CTA in the app.** The premium gold (`#FFD700`, 1.40:1), success green, and error red all fail too. Dynamic type is entirely unsupported (zero `allowFontScaling` handling; fixed heights will clip at accessibility font sizes).
- **There is no bottom tab bar.** The app is one flat stack (`mobile/app/_layout.tsx:66-105`); analytics, settings, and voice chat are buried behind header icons and per-card actions.

**Why it matters.** For the product owner: a visual revamp without primitives means restyling 66 files by hand and drifting back into inconsistency within months — with primitives, the new design is applied *once* and everywhere. A bottom tab bar (Home / Practice / Progress / Profile) is the single largest perceived-modernity change available and directly raises the discoverability of the analytics and voice-chat features we want free users desiring (#10, #16). The contrast failures are simultaneously a usability, brand-quality, and App Store risk issue — and because colors *are* mostly tokenized, darkening the CTA orange is nearly a one-line fix.

**Goal & success metric.** Semantic tokens (surface/onPrimary/accent roles) that eliminate the mode ternaries; a typography scale on the theme; `Button/Card/Sheet/AppModal/ScreenHeader` primitives (much of this already exists unused in `mobile/styles/common.ts` — extract, don't rewrite); AA-compliant palette; tab bar via `app/(tabs)/_layout.tsx`; dark-mode holes closed. Success = a new screen can be built entirely from primitives, and a theme change propagates app-wide.

**Engineering notes.** Run as a track alongside feature work: primitives first, then migrate screens as they're touched by other items. Quick wins to start: delete the two dead theme files (~330 lines), collapse `authTheme` into `lightTheme`, add `theme.typography` from the existing `BASE_TYPOGRAPHY_SCALE` in `mobile/utils/responsive.ts`, darken primary-on-white. `docs/ui-design-system.md` documents a system that doesn't exist in code — rewrite it as part of this (#20).

---

### 9. The delight layer: haptics, sounds, and motion in the core loop

**Category:** Engagement + Perceived quality · **Effort:** S–M

**The problem.** `expo-haptics` is not even installed — zero tactile feedback anywhere in a tap-heavy quiz app. A correct answer looks identical to an incorrect one except for text color: no animation, no sound, no haptic (`mobile/components/quiz/QuizContent.tsx:555-585`). Lottie is installed and paid for but used only for a voice orb and one onboarding slide — no celebration, confetti, or streak animation exists (the once-per-lifetime `CongratulationsModal` uses a text emoji). The app uses only the legacy `Animated` API (no Reanimated, no gesture-handler), with several `useNativeDriver: false` shimmer loops janking the dashboard's first paint. 391 `TouchableOpacity` vs 5 `Pressable` — no press-state design language. Skeletons exist only on the dashboard; the quiz and flashcard loops — the core product — still show spinners. Images use raw `<Image>` with hand-rolled retry logic instead of `expo-image` (no caching, no progressive loading).

**Why it matters.** This is where "new design" becomes *felt* rather than seen. Feedback on the answer moment is the cheapest engagement win in the codebase — the difference between a quiz that feels like a form and one that feels like a game. It's the highest perceived-quality-per-line-of-code item on this list, and it compounds #5 (session-end celebration) and #6 (streak moments).

**Goal & success metric.** Haptic + sound + micro-animation on correct/incorrect answers, card flips, and streak increments; a `Pressable` wrapper with press-scale + haptic replacing `TouchableOpacity` incrementally; Reanimated for new motion; skeletons in quiz/flashcard loading; `expo-image` for word images; celebration Lottie at session end and milestones. Metric: session length and sessions/week (plus app-store review sentiment).

**Engineering notes.** Install `expo-haptics`, `react-native-reanimated`, `expo-image`. The `Pressable` wrapper migration can be mechanical and gradual. Fix the four `useNativeDriver: false` shimmer loops while in there.

---

### 10. Progressive limits + aspirational gating — replace walls and invisibility

**Category:** Conversion · **Effort:** S–M

**The problem.** Free limits (1 wordlist / 10 words) arrive as an ambush. There is no word counter, no warning at word 8 or 9 — then the 11th word **closes the user's wordlist modal** (`mobile/components/dashboard/WordlistDetailModal.tsx:393-404` calls `onClose()` first) and shows a dialog that contains *no price* and a "View Plans" button that goes to Settings (#2). "Maybe Later" dumps the user on the dashboard with their task destroyed. The limit is never disclosed upfront: the onboarding Free-plan card (`mobile/app/onboarding/account.tsx:29-41`) never mentions the caps — users build a mental model of an unlimited product and then hit a wall, the worst sequencing for trust.

Premium features simultaneously use the two worst gating patterns: **bait-and-tap** — Speak and Stats render identically to free features until the tap gets rejected (`mobile/components/dashboard/WordlistItem.tsx:294-312`, styling at `:497-596`) — and **invisibility** — quiz-type filtering is silently hidden (`showQuizTypeSelector={isPremium}`, `mobile/app/quiz.tsx:439`) and offline mode renders `null` for free users (`mobile/components/OfflinePreloader.tsx:90`), even though both are advertised as onboarding slides. The product advertises features it then makes undiscoverable. Limits are also duplicated magic numbers in the client (`dashboard/index.tsx:150`, `WordlistDetailModal.tsx:394`) rather than server-driven.

**Why it matters.** The limit-hit is the highest-intent moment a freemium app gets, and we currently spend it destroying the user's context and hiding the price. A visible "7/10 words" meter converts the wall into an anticipated, fair upgrade moment. Lock badges convert silent hiding into ambient desire — every dashboard view becomes a quiet reminder that there's more. Honest upfront framing of the free plan removes the betrayal moment and pre-frames the upgrade as expected rather than punitive.

**Goal & success metric.** Word-count meter with a soft nudge at 8/10; on the 11th word, a paywall overlay *on top of* the wordlist (no `onClose()`) so "Maybe Later" returns the user to their task; price visible in the limit dialog; lock badges on Speak/Stats/quiz-types/offline for free users; free-plan caps stated plainly in onboarding; limits served from the API. Metrics: limit-hit → paywall-view → purchase funnel (per-trigger, via #3).

---

## Tier 3 — Next wave (pull in as capacity allows)

### 11. Fix the US iOS purchase path

**Category:** Conversion · **Effort:** M

**The problem.** Payment routing sends US iOS users — the highest-ARPU segment — down the *worst* path: they never see the native `RevenueCatPaywall`. They get hardcoded plan cards in Settings (`mobile/app/settings.tsx:699-746`) with a **wrong price** ($69.99 vs the backend's $69.90 — `app/settings.tsx:74` vs `api/internal/model/subscription.go:172`), a hardcoded `$` symbol with no localization, different feature copy and visual design than the RevenueCat path, and an out-of-app browser bounce to Stripe checkout (`WebBrowser.openAuthSessionAsync`).

**Why it matters.** The segment most likely to pay gets the least-tested, highest-friction, and visually weakest purchase experience — and is quoted a price that doesn't match the backend. Whatever the Stripe-vs-IAP margin math says, the current implementation gap is costing conversions in the most valuable market.

**Goal & success metric.** Either route US iOS through native IAP like everyone else, or invest in making the Stripe path first-class: same paywall design, correct server-driven prices, in-app presentation. Metric: US iOS paywall→purchase rate vs other segments (measurable after #3).

---

### 12. Win-back and offer mechanics

**Category:** Conversion · **Effort:** M

**The problem.** When a user dismisses the paywall, `onClose` just closes (`RevenueCatPaywall.tsx:237`). There are no exit-intent offers, no promo codes (`presentCodeRedemptionSheet` never called), no RevenueCat promotional/win-back offers for lapsed subscribers, no lifetime tier, no student pricing, no referral program. A "Maybe Later" tap has no memory — no counter, no escalation, no different treatment on the third dismissal versus the first.

**Why it matters.** Most users say no to a paywall several times before saying yes. Offer mechanics are how mature subscription apps monetize the "no": a dismiss-triggered discount, a win-back offer 30 days after churn, a redeemable code for influencer campaigns. These are among the highest-ROI additions once the paywall funnel is instrumented (#3), because they act on users who already demonstrated intent.

**Goal & success metric.** Exit-intent offer on Nth paywall dismissal (A/B tested via #3's flags); win-back offers for lapsed subscribers; promo code redemption. Metrics: recovered-conversion rate on dismissals, resubscription rate.

---

### 13. Bulk word entry and import

**Category:** Engagement + Conversion (power users) · **Effort:** M

**The problem.** Words can only be added one at a time, each with an async AI round trip. No paste-a-list, no CSV/Anki import, no photo/OCR capture. An unimplemented `todo` entry for bulk add already exists in the repo. Free users who exceed the 10-word cap silently have AI enrichment skipped (see #17), compounding the pain.

**Why it matters.** Power users — students with vocabulary lists, teachers, serious learners — are the likeliest subscribers, and this is their biggest friction in the core loop. "Snap a photo of your textbook page, we extract the vocab" is both a strong premium differentiator and a genuinely magical demo moment that markets itself.

**Goal & success metric.** Paste-a-list bulk add (free, within limits); CSV/Anki import and photo/OCR capture as premium features. Metrics: words added per user, activation rate of imported-list users, premium conversion among importers.

---

### 14. Expand the notification repertoire

**Category:** Retention · **Effort:** M

**The problem.** The mature server-side push pipeline (`api/internal/service/push_notification_service.go`, 537 lines — batching, receipts, token cleanup) sends exactly **two** notification types: a daily reminder hardcoded to 11:00 local and a due-items reminder, jointly capped at **2 per rolling 7 days**. Missing: streak-at-risk ("your 12-day streak ends in 3 hours" — the most effective retention push in this category), "your words are enriched and ready", weekly recap, milestone congratulations, and win-back sequences for 7/14/30-day churned users. No user-chosen reminder time (the i18n copy literally promises "pick morning, afternoon, or evening slots" for a screen that doesn't exist), no Android notification channels (`setNotificationChannelAsync` never called — no per-category user control on Android 8+), no local scheduled notifications as offline fallback, and sounds are disabled client-side (`shouldPlaySound: false`) even though the server sends `Sound: "default"`. The permission prompt only fires after the 2nd quiz answer — users who never reach question 2 are permanently unreachable.

**Why it matters.** The expensive infrastructure is built; it's being used at a fraction of capacity. The 2-per-week cap makes the system structurally incapable of supporting a daily habit — appropriate caution for generic nags, but wrong for notifications the user *asked for* (a chosen reminder time, a streak they're protecting). Design this together with #6 and #7: goal-based and streak-based notifications feel like a service, not spam, and earn a higher cap.

**Goal & success metric.** Streak-at-risk, enrichment-ready, weekly recap, and win-back notification types; user-chosen reminder time (raising the cap for user-scheduled reminders); Android channels; an onboarding-time permission ask framed around the user's chosen goal. Metrics: push opt-in rate, push-open rate, push→session conversion (via #3).

---

### 15. Ambient presence: widgets and share cards

**Category:** Retention + Acquisition · **Effort:** M–L

**The problem.** The app has zero presence outside its own icon: no home-screen widget, no Live Activities, no app shortcuts/quick actions, no watch app, and nothing shareable — no streak card, no milestone image, no "I mastered 100 words" moment. There is no viral loop of any kind.

**Why it matters.** A home-screen widget showing due words + current streak is a proven daily-habit anchor for learning apps — it's a reminder that can't be swiped away and doesn't count against notification budgets. Share cards are the only cheap acquisition channel available to a small team: every streak milestone becomes a potential install. Both also reinforce #6's streak investment.

**Goal & success metric.** iOS/Android widgets (due count, streak, one-tap into practice); shareable streak/milestone cards from the session-end screen (#5). Metrics: widget-install rate and widget→session opens; shares per milestone.

---

## Tier 4 — Fix-as-you-touch (bundle into revamp PRs)

### 16. Free-tier analytics teaser

**Category:** Conversion + Retention · **Effort:** S

**The problem.** Every motivating number in the app — streak, words today, accuracy, mastery curve, box distribution — is fully paywalled at the entry point (`mobile/components/dashboard/WordlistItem.tsx:305-312` blocks navigation for free users). Free users see only a percentage bar and a conditionally-rendered flame. Analytics is also per-wordlist only, with no account-level view, and reachable only via a per-card action.

**Why it matters.** This is upside-down for conversion: users pay to *keep* something they're invested in, not to *see* whether they'd be invested. Showing free users a real-but-limited view (this week's stats) with blurred/locked previews of the full history converts curiosity into demonstrated value.

**Goal.** Free tier gets current-week stats + streak; historical charts and advanced views stay premium with visible locked previews. Wire the "Progress" tab (#8) as the entry point.

### 17. Fix silent free-tier degradations

**Category:** Trust + Conversion · **Effort:** S

**The problem.** Free users over the 10-word cap have AI enrichment **silently skipped** (`api/internal/service/user.go:311-362` gates worker eligibility) — words sit un-enriched with no explanation, indistinguishable from a bug. Free users also get 1-hour analytics cache TTL vs 1 minute for premium (`api/internal/http/analytics.go:26-31`) and 15-minute stale session data with no refetch-on-focus (`mobile/hooks/useUserSession.ts:110-115`).

**Why it matters.** Silent degradation reads as "the app is broken," not "I should upgrade." It generates support tickets and bad reviews instead of conversions. Every deliberate limitation must be visible and attributed: "Upgrade to enrich these words" is an upsell; a word that mysteriously never gets its image is a defect.

**Goal.** Surface every intentional limitation in the UI with an upgrade path, or remove it.

### 18. Accessibility and dynamic type pass

**Category:** Quality + Compliance · **Effort:** M (incremental)

**The problem.** Beyond the contrast failures in #8: only ~1 in 5 touchables has an accessibility label (88 labels vs 391 touchables); the entire analytics chart tree, settings, and flashcards have zero a11y props; ~37 accessibility strings are hardcoded English bypassing i18n (including one describing a long-press menu that no longer exists — `WordlistItem.tsx:326`); and OS font scaling is entirely unsupported — zero `allowFontScaling`/`maxFontSizeMultiplier` usage, with fixed heights that will clip text at accessibility sizes.

**Why it matters.** It's a real user-base slice, an App Store review risk, and — pragmatically — the cheapest time to fix it is while every screen is being rebuilt under #8 anyway. Baking a11y requirements into the new primitives means the whole app inherits compliance.

**Goal.** A11y props required on the #8 primitives; `maxFontSizeMultiplier` + `minHeight` conventions; a11y strings through i18n. Audit with each screen migration.

### 19. Performance and code-health cleanup

**Category:** Engineering health · **Effort:** S–M (incremental)

**The problem.** 53 call sites run `createStyles(theme)` unmemoized on every render (only `signin.tsx` memoizes); four shimmer loops animate on the JS thread (`useNativeDriver: false`) during dashboard first paint; six files exceed 1,000 lines (`settings.tsx` and `WordlistDetailModal.tsx` at 1,367 each); a route URL is built from an unencoded wordlist name (`WordlistItem.tsx:147` — inconsistent with line 299 which does encode); free-plan limits are duplicated client-side magic numbers; `router.replace()` throughout onboarding breaks Android hardware back; and RevenueCat annual detection matches the string literal `"p2m"` (`useRevenueCat.ts:38`).

**Why it matters.** Each is small; together they're the difference between a revamp that ships fast and one that fights the codebase. The god-file decomposition in particular is a prerequisite for parallelizing the #8 screen migrations across the team.

**Goal.** Memoization convention enforced in the new primitives; god files decomposed as they're migrated; the specific bugs fixed outright.

### 20. i18n and documentation hygiene

**Category:** Quality · **Effort:** S

**The problem.** Non-English locales are missing 25–33 keys each and carry 13–15 stale orphaned keys; the `t("key", "English fallback")` inline-default anti-pattern is widespread, hiding missing keys and duplicating copy; an entire orphaned `onboarding.notifications` i18n block exists in all 8 locales for screens that were never built; `en.json` contains dead strings (`upgrade.startFreeTrial`). Meanwhile `docs/ui-design-system.md` documents a `Colors` object, typography scale, breakpoints, and a `useResponsive` hook that **do not exist in the code** — it will actively mislead anyone executing this revamp.

**Why it matters.** Seven of the app's eight locales are the product for most of the world; silent English fallbacks erode the quality perception this revamp is meant to build. And a design doc that contradicts reality is worse than no doc — especially at the start of a design-system project.

**Goal.** Locale gap closed and orphans pruned (scriptable); inline defaults replaced by a missing-key CI check; `ui-design-system.md` rewritten to describe the actual #8 system. The orphaned notification-preferences copy becomes the spec for #14's reminder-time screen — it was clearly designed once already.

---

# Part 2 — Backend: Payments Migration, Bugs, Security, Performance, Cost

## Tier 5 — Backend workstreams

### 21. Replace Stripe and RevenueCat with first-party native IAP

**Category:** Payments simplification (decided direction) · **Effort:** L (multi-phase)

**Direction.** Eliminate both Stripe and RevenueCat. iOS purchases go through StoreKit 2 verified server-side via the App Store Server API; Android purchases go through Google Play Billing verified via the Play Developer API. One store per platform, no third-party billing SDK, no RevenueCat fees, no split-brain provider routing.

**Why this is the right call.** Today the payment stack is *four* paths (Android→RevenueCat, iOS US→Stripe via external browser, iOS non-US→RevenueCat, Web→Stripe) with routing that lives only in the mobile client (`mobile/hooks/useRevenueCat.ts:229-268` — the backend will create a Stripe checkout for anyone). The two paths have different prices, different copy, and different designs; the audit found the US iOS Stripe flow is the weakest experience aimed at the highest-ARPU users (#11), and the external-browser purchase's App Store legality rides on the current US injunction posture. Consolidating to native IAP removes an entire class of complexity, risk, and reconciliation bugs — and the audit found plenty of the latter (see #22).

**What the audit found we can reuse (~60% of the plumbing survives):**

- The `subscriptions` + `subscription_events` tables are already provider-agnostic since migration `000056` (`external_event_id`, `provider` enum). Only the enum values and the `check_provider_fields` CHECK from `000050` need widening.
- The `Subscription` model with `IsActive()` grace-period logic (`api/internal/model/subscription.go:200-213`), the `update_user_subscription` Postgres trigger that propagates entitlement onto `users`, the transactional `CreateSubscription`/`UpdateSubscription` repository methods, and the `HasProcessedEvent` idempotency helper — all reusable as-is.
- The River webhook-worker pattern (`stripe_webhook_worker.go` / `revenuecat_worker.go` are near-identical templates for an Apple-notifications worker and a Google-RTDN worker).
- The reminder worker and all five email templates; the `GET /subscription/status` client contract; the mock-based integration test harness.

**What must be built:**

1. **App Store Server API client** — ES256 JWT auth, subscription status/history lookups — plus **JWS verification** (x5c chain validation against Apple Root CA G3). This is the one piece with no analogue in the codebase and must not be shortcut (never "decode without verify").
2. **App Store Server Notifications V2 endpoint + worker** — handle `SUBSCRIBED`, `DID_RENEW`, `DID_CHANGE_RENEWAL_STATUS`, `EXPIRED`, `GRACE_PERIOD_EXPIRED`, `DID_FAIL_TO_RENEW`, `REFUND`, `REVOKE`. Note: **refund/revocation handling does not exist anywhere in the current code** — today a refund changes nothing. It becomes mandatory.
3. **Play Developer API client** (`purchases.subscriptionsv2.get`) + **purchase acknowledgement** — Google auto-refunds unacknowledged purchases after 3 days, a failure mode with no precedent in this codebase.
4. **RTDN endpoint** — Google Cloud Pub/Sub push with OIDC verification, idempotent on `messageId`.
5. **Unified verification service** — `POST /subscription/iap/verify` and `/iap/restore`, replacing the RevenueCat restore endpoint.
6. **Store↔user linkage table** (`user_id ↔ original_transaction_id / purchase_token`), populated at verify time via `appAccountToken` (iOS) / `obfuscatedExternalAccountId` (Android). Server notifications must resolve to a user without the client present — today the link is `app_user_id == users.id`, which only RevenueCat provided.
7. **Schema migration** — with no paying users, simplify rather than extend: recreate `subscription_provider` as `('apple','google')`, drop the `check_provider_fields` constraint and the Stripe/RevenueCat columns, add `original_transaction_id`, `purchase_token`, and **`store_environment`** (sandbox/production — without it a tester's sandbox renewal grants real premium).
8. **Mobile client** — `expo-iap`/`react-native-iap` with a `useIAP` hook; finish/acknowledge transactions only *after* server confirmation; a **pending-purchase retry queue** (a purchase can complete while the verify call fails — RevenueCat absorbed this for us; we now own it). The paywall styles survive; only the product-loading top half of `RevenueCatPaywall.tsx` is rewritten. `openNativeSubscriptionManagement` in `mobile/api/subscriptions.ts:79-143` already uses pure store deep links — keep it.

**Why this is the ideal moment: there are no paying users.** With zero existing subscribers there is nothing to grandfather, nothing to backfill, and no dual-stack transition period. Stripe and RevenueCat can be deleted outright in a single cutover — the schema can even be simplified rather than extended-for-compatibility (drop the `stripe`/`revenuecat` enum values and columns instead of carrying them). This window closes the moment the first subscription is sold; doing the migration **before** the conversion work in #1–#2 ships is strictly cheaper than after.

**Remaining considerations (all small):**

1. **Web loses the ability to sell.** Near-zero cost — the web app is a marketing site whose pricing buttons link to `#download`; no web checkout exists. The real loss is optionality: promo codes, B2B/invoice sales, and any future commission-free channel. Accepted trade-off; revisit only if a web learning platform ships.
2. **US iOS margin.** Native IAP means 15% under Apple's Small Business Program (a near-certainty at these price points), vs Stripe's ~2.9%+30¢ — model the delta at ~12pp. In exchange: native purchase UX for the highest-ARPU segment, no external-browser flow, no injunction-dependent legality, and one less integration to maintain.
3. **Sandbox/production bifurcation and store test accounts** become our problem directly (RC normalized this before) — hence the `store_environment` column.

**Phasing (authoritative dependency order):**

The implementation and validation sequence is owned by `docs/MOBILE_APP_REVAMP_EXECUTION_PLAN.md`; this strategy describes direction and must not be used as a competing execution checklist. The required order is: (0) classify every effective entitlement and fix the activation measurement baseline; (1) approve the store-neutral entitlement, identity, idempotency, operation, and mobile contracts; (2) implement additive Apple/Google verification, notifications, persistence, effective-access projection, and native mobile IAP while the legacy providers remain available for rollback; (3) prove the complete lifecycle in App Store sandbox/TestFlight and Google Play internal testing, reconcile the controlled approval accounts, and rerun the production entitlement audit; then (4) remove Stripe/RevenueCat code, routes, jobs, schema, dependencies, configuration, and reachable checkout behavior in one reviewed cutover. Live provider deletion cannot precede Gate 3, and final release still requires the measurement and store-owner gates recorded in the execution plan.

**Success metric:** one purchase path per platform; purchase→entitlement latency < 5s; refunds revoke access automatically; zero reconciliation bugs of the class listed in #22.

---

### 22. Fix the critical backend correctness bugs

**Category:** Correctness / data integrity · **Effort:** S–M per bug (mostly small)

**The problem.** The backend audit found live, user-facing breakage — not theoretical risk. The worst offenders:

- **`PUT /wordlists/:id` is permanently broken**: a missing comma in the UPDATE statement (`api/internal/repository/wordlist.go:174` — `language_code=$3 updated_at=NOW()`) makes every wordlist rename/edit return 500. One-character fix.
- **Stripe webhook status-enum mismatch**: Stripe statuses are written raw into a 4-value Postgres enum (`subscription.go:195,274`), and the enum spells **`cancelled`** while Stripe sends **`canceled`** — every subscription-update webhook for a canceled/trialing sub would throw an enum error and retry forever. With no paying users and Stripe slated for deletion (#21), this needs no fix — but the same raw-status-into-enum pattern **must not be ported** to the Apple/Google workers, whose status vocabularies also won't match the enum.
- **Error-report image regeneration panics on every attempt**: `image_generator_worker.go:139-140` constructs a zero-value `LeitnerSystemStrategy` (nil DB pool) — each user image report triggers up to 25 retries × 25 *paid* image generations, then fails anyway.
- **Reporting a content error deletes the user's learning history**: `_unrelated_meaning`/`_unrelated_example` reports call `DeleteWordDefinitions`, which `ON DELETE CASCADE`s through `leitner_system_tracking` and `quiz_performance` — one tap on "this example is wrong" resets the word to box 1 and erases its analytics.
- **Updating a word wipes its audio and notes**: the "mark as learned" flow (`http/word.go:103`) sends zero values for `audio_url`/`notes` and the repository writes them unconditionally — silent data loss on a common operation, which then disables audio quiz types for that word.
- **Case-sensitive signup + case-insensitive login = permanent lockout**: `Alice@x.com` and `alice@x.com` can both register (unique constraint is on raw `email`), after which login matches both rows and rejects the "ambiguous" credentials for **both** accounts (`repository/user.go:64,100`, `service/user.go:161`).
- **Transactions that commit on error**: at least four `defer`-commit blocks have shadowed/unassigned `err` (`repository/definition.go:72`, `service/error_reporting.go:170`, `service/word.go:153-155`, `definition_fetcher_worker.go:221-223`) — failures commit partial state, and in the definition-fetcher case River then retries and **inserts the entire definition set twice**.
- **`errors.Is(err, &common.NotFoundError{})` is always false** (pointer vs value receiver) in three handlers — every not-found becomes a 500 + Sentry panic event.
- **Webhook idempotency is check-then-act**: both webhook workers check `HasProcessedEvent`, then process, then insert — concurrent redeliveries double-send emails and double-apply plan changes before the unique constraint catches the insert.
- **Cross-user writes**: `PUT /wordlists/<victim_id>/words/<my_word_id>` reparents an attacker's word into another user's wordlist (no ownership check on the target wordlist, `repository/word.go:177,198`), inflating the victim's free-tier quota.
- **Missing uniqueness**: no unique constraint on `words(wordlist_id, name)` or `word_definitions(word_id, definition_id)` — duplicate adds spawn duplicate AI jobs and inflate every join.

**Why it matters.** These are support load, corrupted analytics, and data loss happening *now* — and they directly undermine the revamp: the data-loss bugs will be blamed on the redesign if they surface during it. The webhook-pattern bugs (idempotency, enum mapping, commit-on-error) matter because #21 will otherwise copy them into the new Apple/Google workers — fix the patterns in the templates before they're cloned.

**Goal.** All critical/high items in Appendix C fixed and covered by regression tests. The full inventory (25 medium + low findings, including timezone-dependent streak breakage, positional Expo receipt matching, and non-transactional account deletion) is in Appendix C.

---

### 23. Backend security hardening

**Category:** Security · **Effort:** M

**The problem.** The audit found several genuinely serious issues:

- **Empty-key JWT verification**: the auth middleware reads `JWT_KEY` per-request with no emptiness check (`http/midlewares.go:48-53`). If the env var is ever missing or typo'd, the API boots normally and **accepts any token signed with an empty key** — full authentication bypass. There is no startup config validation for any secret.
- **1-year JWTs with the subscription plan baked in, and no revocation**: quiz gating, analytics cache TTLs, and error-report limits read the *token claim*, not the DB (`service/user.go:42,78-88`). An upgrader stays gated until re-login; a churned user keeps premium features for up to a year; password reset does not invalidate existing sessions; logout only clears the cookie. This is simultaneously the biggest entitlement bug (blocks #21) and the biggest session-security gap. Fix: short-lived access tokens + refresh rotation, a `token_version` bumped on password change, plan read from DB.
- **No rate limiting on any auth endpoint**: unlimited credential stuffing against `/login` (each attempt burns a ~60ms bcrypt hash — also a CPU-DoS), unlimited password-reset email bombing via Resend, unlimited signup (each free account can trigger OpenAI spend). The codebase's only rate limiter guards `/errorReports`.
- **Auth tokens are being shipped to Sentry**: `SendDefaultPII: true` (`common/sentry_init.go:29`) attaches request headers and cookies — where the 1-year `Authorization` JWT lives — and production logs at Debug level into Sentry Logs, including every login email. Credential-theft path + GDPR exposure + ingest cost.
- **Unauthenticated open redirect**: `GET /subscription/checkout-redirect` redirects to any `redirect_uri` (`http/subscription.go:190-200`); a webhook-signature test bypass is compiled into the production binary (`subscription.go:126-130`); the RevenueCat webhook handler does a non-constant-time compare and **panics if its env var is unset** — both go away with #21, but not for ~12 months of dual-stack.
- **Unvalidated uploads (closed locally by `UPLOAD-1` on 2026-08-17)**: the original path had no body/decoded/canonical size caps, trusted caller-controlled media metadata and extension, and allowed unsafe object naming. The implemented contract now performs bounded admission and canonical raster validation, uses server-owned user-scoped keys/content types, reuses a startup MinIO client, and reconciles ambiguous object/reference outcomes durably; matching 1.1.2 store binaries and OTA proof remain owner release gates.
- **Quiz distractors leak other users' private content**: `GetRandomMeanings`/`GetRandomTokens` draw from the **global** definitions table with no user, wordlist, or even *language* filter (`repository/definition.go:105-166`) — cross-tenant disclosure, wrong-language distractors, and (combined with prompt injection via user-supplied tokens interpolated into a *system* message, `openai/chat_completion.go:281-286`) a content-poisoning vector into other users' quizzes.
- Assorted: login timing side-channel enables user enumeration; 5-char minimum passwords (and reset/update paths enforce even less); `BCRYPT_COST` env override accepted in production; deprecated `dgrijalva/jwt-go` (CVE-2020-26160); permissive dev CORS that fails *open* if `ENV` is misspelled; no security headers; missing `c.Abort()` after a failed middleware check.

**Why it matters.** Any one of the first four is incident-report material. The JWT items are also hard blockers for #21 (store-driven renewals/refunds must take effect without re-login). And the pre-revamp period is the cheapest time to rotate token semantics — fewer users are affected than after the growth work succeeds.

**Goal.** Startup fail-fast config validation; short-lived tokens + refresh; rate limiting on all four auth endpoints; `SendDefaultPII` off + PII scrubbed from logs; upload validation + global body limits; distractors scoped by language and pool; the checklist in Appendix C closed.

**Current execution status (2026-08-17).** `AUTH-1`–`AUTH-3`, `HTTP-SEC-1`, `UPLOAD-1`, `API-BOUND-1`, and `API-RATE-1` are locally complete with the evidence recorded in `MOBILE_APP_REVAMP_EXECUTION_PLAN.md`. Remaining deployed configuration, legacy-cookie rollout, store binaries, OTA proof, and later security/cutover milestones retain their explicit owner gates.

---

### 24. Backend performance & reliability

**Category:** Performance / Ops · **Effort:** S (indexes) + M (rest)

**The problem.** The scaling cliffs are concentrated in the hottest path — the quiz loop:

- **Zero indexes on `word_definitions`** — the join table in every quiz-selection, box-distribution, and analytics query has only a surrogate PK (`migrations/000004`). Every quiz request sequential-scans it. Two `CREATE INDEX CONCURRENTLY` statements are the single highest-ROI change in the entire backend.
- **`ORDER BY RANDOM()` over the whole corpus** for every question's distractors (`repository/definition.go:126,163`), with no index on `definitions(part_of_speech)` — cost grows with total corpus size, not user data.
- **Analytics caching is broken twice over**: the stats cache is written with an already-canceled errgroup context so it *never persists* (`analytics_cached.go:57,94`), and invalidation keys omit the `days` suffix so learning-progress/practice-time caches are *never invalidated* (`analytics_cached.go:337,340` vs `:160,268`). Premium users effectively run raw recursive-CTE streak queries (up to 365 iterations per wordlist) on nearly every dashboard load, under a 2-second global route timeout that doesn't actually cancel the work.
- **Worker concurrency required an explicit database budget**: the audited configuration actually exposed 235 nominal River slots, not 226, against a 10-connection worker pool. MaxWorkers is a scheduling ceiling rather than a connection reservation. The initial containment capped definition work at 4; the durable extraction now assigns 2 slots to definition generation and 2 to transactionally enqueued meaning-audio jobs, preserving the four-worker provider budget and the aggregate ceilings of 189 with the temporary legacy rollback surface and 179 in IAP-only mode. OpenAI calls and workers are bounded and cancellable, and no TTS or MinIO call remains inside the definition transaction. Queue depth, job latency, production PostgreSQL capacity, and worker replica count must still be measured before increasing either queue or the database pool.
- **Production graceful shutdown is inverted**: the prod branch sleeps 5s and exits without calling `srv.Shutdown` — in-flight requests die on every deploy (`cmd/api/server.go:64-77`). There is **no health-check endpoint** of any kind, so load balancers can't gate deploys or detect a dead DB pool. Redis failure at boot permanently disables caching until restart (`common/redis.go`, memoized error).
- Assorted: an expensive 4-table diagnostic query runs on every "no due items" response (the steady state for engaged users); quiz selection materializes 50 full definitions with JSONB aggregation to use one; unbounded queries and `IndentedJSON` on the largest payloads; per-upload MinIO client construction; per-answer unbounded goroutine doing a DB write + ~60 Redis DELs with no `recover()` (a panic there kills the API process); pgBouncer wired up without pgx being configured for it; migrations run without lock timeouts.

**Why it matters.** The revamp's entire purpose is to grow traffic into this backend. The quiz loop — the thing #5–#7 will drive users into many times a day — is the least-indexed, most-expensive path in the system, and the worker fleet that powers the AI magic (#4's starter packs will hammer it) can starve unrelated work if long transaction holders are not bounded. The week-2 index work requires *zero code changes* and should be measured with `pg_stat_statements` before/after.

**Goal.** Indexes shipped; OpenAI clients get timeouts/retries/status-code handling; TTS moved out of transactions; queue caps and pools sized from measured database capacity rather than nominal worker totals; graceful shutdown fixed; `/healthz` + `/readyz`; caching actually caches. Success = p95 quiz-question latency flat as corpus and DAU grow.

---

### 25. AI cost optimization

**Category:** Cost / Margin · **Effort:** S–M per lever

**The problem (and the levers, ranked by expected saving):**

1. **Definition dedup is racy and language-blind** (~60–80% of OpenAI spend): reuse only triggers when a row already exists — two users adding the same word concurrently both pay for GPT-4o + images + 7 TTS calls. Worse, reuse matches on token *only*, ignoring language (`service/word.go` → `FindArgs{Name}`) — "no" in a Spanish list reuses the English definition (a correctness bug *and* a cost bug). Fix: unique key on `(token, language)` + advisory-lock "generation in progress" coalescing.
2. **`gpt-4o` pinned for a strict-JSON dictionary lookup** (~85–90% of chat spend): a mini-tier model handles this task; `max_tokens` is currently unbounded and `temperature` unset (`openai/chat_completion.go:295`). The existing `validateDefinitions` check is a ready-made pass/fail benchmark oracle.
3. **7 eager TTS calls per non-verb definition** (~70% of TTS spend): all seven examples get audio up front, but only one is ever played per quiz. Generate 2–3 lazily on first audio-quiz use — verbs already do the lazy pattern correctly.
4. **Every example audio for every non-English language is generated with the wrong voice**: `example_audio_worker.go:82,92` passes a *voice name* where the API expects a *language code*, so it silently falls through to the English `alloy` voice with English pronunciation instructions. We are **paying for wrong-language audio right now** — this is a one-line fix that also improves the product. (Related: `buildImagePrompt` passes 4 args to 3-verb format strings, so every image prompt ends with `%!(EXTRA string=...)` — `image_generator_worker.go:212`.)
5. **Content-hash TTS dedup**: the SHA-256 of the example text is already computed but only used for naming — key audio objects on `sha256(text+voice)` globally and identical sentences across the corpus cost nothing.
6. **Images generated eagerly per sense**: a word with 5 senses generates 5 DALL-E images up front, though image quizzes start at box 3. Generate lazily on box-3 arrival, cap to the primary sense.
7. **Realtime voice chat has no cost ceiling**: no `max_output_tokens`, no session duration cap, no per-day quota — and its usage telemetry is *client-supplied and unvalidated* (`http/realtime_telemetry.go`), so cost dashboards built on it are forgeable. This is the most expensive API in the OpenAI catalog, gated only by "is premium".
8. **Sentry ingest**: production ships Debug-level logs (4+ per quiz answer) to Sentry Logs — raising the threshold to Warn cuts ingest by an order of magnitude.

**Why it matters.** #4 and #10 will multiply free-tier AI usage (starter packs, more words enriched); #2's trial will multiply premium usage. Unit economics need to improve *before* the growth work succeeds, not after the bill arrives. Items 1, 2, and 4 together plausibly cut AI spend by more than half while *fixing* product bugs (wrong-language definitions and audio) at the same time.

**Goal.** Language-correct dedup with generation coalescing; model right-sizing with the validation oracle as the quality gate; lazy TTS/images; hash-based dedup; realtime quotas. Metric: AI cost per new enriched word and per premium user per month, tracked before/after each lever.

---



| Bug | Location |
|---|---|
| Raw email sent as PostHog event property (PII) on 3 auth events | `app/signup.tsx:140-146`, `app/signin.tsx:79` |
| Duplicate `signup_completed` + `user_signed_up` events inflate funnel counts | `app/signup.tsx:140-146` |
| `onboarding_completed` fires when users tap "Already have an account? Sign in" | `app/onboarding/account.tsx:67-70` |
| Premium comparison card lists a *free*-plan feature (copy-pasted i18n key) | `app/onboarding/account.tsx:57` |
| Annual price mismatch: mobile $69.99 vs backend $69.90; hardcoded `$` | `app/settings.tsx:74,722` vs `api/internal/model/subscription.go:172` |
| Paywall load failure has no retry, no fallback, no support link | `components/RevenueCatPaywall.tsx:233-257` |
| Android hardware back exits the app throughout onboarding (`router.replace`) | `app/onboarding/*.tsx` |
| Pulsing FAB invites free users to an action that will be rejected | `app/dashboard/index.tsx:250-264` |
| Wordlist name interpolated into route without `encodeURIComponent` | `components/dashboard/WordlistItem.tsx:147` |
| Stale a11y label describes a removed long-press menu | `components/dashboard/WordlistItem.tsx:326` |
| Returning users with zero wordlists get the newbie `WelcomeOverlay` | `hooks/useWelcomeState.ts:80-111` |
| `SpaceMono-Regular.ttf` shipped but never loaded; app is 100% system font | `mobile/assets/fonts/` |
| Only MONTHLY/ANNUAL package types handled; others mislabeled "yearly" | `components/RevenueCatPaywall.tsx:301-303` |

## Appendix B — What's already strong (build on, don't rebuild)

- **Server-side push pipeline**: batched Expo sends, receipt reconciliation, token cleanup, localized copy, timezone-aware targeting (`api/internal/service/push_notification_service.go` + workers).
- **Leitner SRS**: deterministic 7-box system with a sensible practice-ahead fallback (`api/internal/service/leitner_system_strategy.go`).
- **AI enrichment pipeline**: definitions, images, TTS, example audio via River workers with retries and dedup.
- **i18n foundation**: 8 locales, 57/70 components wired to `useTranslation`.
- **App review prompt**: well-built eligibility logic (`hooks/useAppReview.ts`) — just move it into the #5 session-end rotation.
- **RevenueCat error handling + optimistic entitlement**: purchase unlocks premium instantly without waiting for the webhook (`hooks/useRevenueCat.ts:26-70`).
- **Dashboard craft**: skeletons, empty-state illustration, entrance animations — the quality bar the rest of the app should be raised to.

## Appendix C — Backend bug inventory (from the API audit)

Severity: **C** = critical (broken feature, panic loop, or exploitable), **H** = high (data loss/corruption, security, or major perf), **M** = medium. Items marked ⚡ are one-line-ish fixes.

### Correctness & data integrity

| Sev | Bug | Location |
|---|---|---|
| C ⚡ | `PUT /wordlists/:id` dead: missing comma in UPDATE SQL (`language_code=$3 updated_at=NOW()`); also drops `pronunciation_system` | `repository/wordlist.go:174` |
| C | Stripe statuses written raw into 4-value enum; `cancelled` (DB) vs `canceled` (Stripe) — subscription-updated webhooks for canceled/trialing subs fail and retry forever | `service/subscription.go:195,274` vs `model/subscription.go:44` |
| C | Zero-value `LeitnerSystemStrategy` (nil DB) in image worker → panic on every error-report-triggered image job; 25 retries × paid image generations | `service/image_generator_worker.go:139-140` |
| H | Error reports of type `_unrelated_meaning`/`_unrelated_example`/`_processing_failed` cascade-delete `leitner_system_tracking` + `quiz_performance` — user's learning history erased by one report | `service/error_reporting.go:191,198` + migrations `000005`/`000030` cascades |
| H | Word update writes zero-value `audio_url`/`notes` unconditionally — "mark as learned" wipes audio and notes | `http/word.go:103`, `repository/word.go:174-180` |
| H | Case-sensitive signup + case-insensitive login: duplicate-case emails lock out **both** accounts; `LOWER(email)` also defeats the unique index (seq scan per login) | `repository/user.go:64,100`, `service/user.go:161` |
| H | Deferred tx commit/rollback with shadowed or unassigned `err` — commits on failure; definition-fetcher variant double-inserts entire definition sets on retry | `repository/definition.go:72`, `service/error_reporting.go:170`, `service/word.go:153-155`, `definition_fetcher_worker.go:221-223` |
| H | `errors.Is(err, &common.NotFoundError{})` never matches (pointer vs value) — not-found becomes 500 + Sentry panic | `http/wordlist.go:139,155`, `http/word.go:81` |
| H | Webhook idempotency is check-then-act; concurrent redeliveries double-apply side effects (emails, plan changes) before the unique insert fails | `stripe_webhook_worker.go:53`, `revenuecat_worker.go:49` |
| H | RevenueCat worker stores the event only after successful processing — partial failures re-run side effects on retry | `revenuecat_worker.go:74-83` |
| H | `GetDefinitionByID` returns `(nil, nil)` on error/not-found → nil-deref panics in image + example-audio workers when a definition was deleted | `service/definition.go:49-55` |
| H | Definition-fetcher swallows errors inside an open tx (`25P02` aborted-tx cascade); job reports success, nothing saved, word stuck at `processing` forever | `definition_fetcher_worker.go:187-227` |
| H | Restore with lost entitlement is a silent no-op — expired sub's stale active row survives; `EXPIRATION` after grace never marks the row canceled (`GetActiveSubscriptionForUser` returns nil) | `revenuecat.go:100-104,340,378,433` |
| H | Nil deref on `*entitlement.ExpiresDate` (null for lifetime entitlements) | `revenuecat.go:113` |
| H | Unchecked `Items.Data[0]` indexing on Stripe payloads → worker panic | `subscription.go:186-199,275-276` |
| H | Provider switch creates a **second active subscription** (double-billing); "current sub" then picked by `created_at DESC LIMIT 1` | `revenuecat.go:151`, `repository/subscription.go:165` |
| H | Three subscription SELECTs omit `provider`/platform columns → reminder worker's provider branch never matches; renewal emails carry empty IDs | `repository/subscription.go:244-247,324-327,364-367` |
| H | Stats cache written with already-canceled errgroup context — never persists | `analytics_cached.go:57,94` |
| H | Cache invalidation keys omit `days` — progress/practice-time caches never invalidated (free tier: hour-stale); free-tier invalidation gated on `isPremium` so never runs | `analytics_cached.go:337,340` vs `:160,268`; `leitner_system_strategy.go:1337` |
| H | Per-answer analytics goroutine launched **before** tx commit (stale cache repopulation) and with no `recover()` — a panic kills the API process | `leitner_system_strategy.go:1314-1364` |
| M | Box-distribution/unlearned-count queries join `word_definitions` without `AND wd.word_id = lst.word_id` — inflated counts; skip-rescue never fires ("no due items" instead) | `leitner_system_strategy.go:348,369,1081,1107` |
| M | Daily aggregates mix Go-UTC dates with Postgres `CURRENT_DATE`/session TZ; no user-local timezone — streaks break at date boundaries | `service/analytics.go:108` vs `repository/analytics/learning_progress.go:91,149,185` |
| M | `DeleteWordDefinitions` decides shared-vs-exclusive from an aggregate count across *all* definitions — exclusive definitions orphaned forever | `repository/definition.go:271-298` |
| M | No unique constraints on `words(wordlist_id,name)` or `word_definitions(word_id,definition_id)` — duplicate adds spawn duplicate AI jobs | migration `000004`, `repository/word.go:64-88` |
| M | Cross-user write: word can be reparented into a victim's wordlist (no target-ownership check), inflating their free-tier quota | `repository/word.go:177,198` |
| M | Deleting another user's word returns 204 instead of 404 | `service/word.go:169-186` |
| M | Grace period anchored on `CurrentPeriodEnd`, not the payment-failure time — access silently extended or cut short | `model/subscription.go:208` |
| M | `MarkRenewalReminderSent` hardcodes `provider='stripe'`; reminder "sent" log reports candidate count, not sent count | `repository/subscription.go:443`, `subscription_reminder_worker.go:161` |
| M | Unknown Stripe price ID silently maps to `PlanFree` — a typo downgrades a paying customer via the DB trigger | `subscription.go:530-542` |
| M | Expo push tickets matched to messages positionally — partial responses deactivate the **wrong** devices | `push_notification_service.go:458-460` |
| M | `CheckReceipts` loop has no progress guarantee on persistent update failure | `push_notification_service.go:223-275` |
| M | One invalid user-supplied timezone string kills the reminder query for **all** users in that run | `repository/push_notifications.go:58,134` |
| M | Account deletion non-transactional; child-delete failures only logged | `service/user.go:256-265` |
| M | Error-report rate limit bypassable: original implementation failed open on DB error and undercounted upserts. Closed locally by `API-RATE-1` with serialized append-only quota events and atomic committed-work accounting | `rate_limiter.go`, `repository/error_report.go`, migration `000081` |
| M | pgx **v3** sentinel (`pgx.ErrNoRows`) compared against pgx/v5 errors — never matches; v3 driver shipped in the binary | `repository/user.go:12,125,244` |
| M | `learning_progress.words_mastered` never written — permanently 0; 30-day progress query returns 29 days (exclusive bound) | migration `000030`; `learning_progress.go:91` |

### Security

| Sev | Issue | Location |
|---|---|---|
| C | JWT verification reads `JWT_KEY` per-request with no emptiness check — missing env var = auth bypass via empty-key HMAC; no startup validation of any secret | `http/midlewares.go:48-53` |
| C | 1-year JWTs embedding `subscriptionPlan`; no refresh rotation, no revocation; password reset/logout don't invalidate sessions; gating reads the claim, not the DB | `service/user.go:42,78-88`, `http/quiz.go:44,146` |
| H | No rate limiting on `/login`, `/users`, `/password/send-reset-email`, `/password/reset` — credential stuffing, bcrypt CPU-DoS, email bombing, signup-driven OpenAI spend | `http/setup.go:54-73` |
| H | `SendDefaultPII: true` ships cookies (incl. the 1-year JWT) to Sentry; prod logs Debug→Sentry incl. login emails | `common/sentry_init.go:29`, `common/logger.go:8,25`, `service/user.go:176-190` |
| H | Unauthenticated open redirect on `/subscription/checkout-redirect` | `http/subscription.go:190-200` |
| H | Profile upload: original path had no body cap, caller-controlled MIME/extension, unsafe object naming, and public-origin stored-content risk. Closed locally by `UPLOAD-1`; matching 1.1.2 store binaries and OTA proof remain owner-gated | `http/user.go`, `common/profile_image.go`, `common/minio.go` |
| H | Quiz distractors drawn from the global corpus with no user/list/language filter — cross-tenant content disclosure + wrong-language options + poisoning vector | `repository/definition.go:105-127,151-166` |
| M | User token interpolated into the OpenAI **system** message with naive quoting — prompt injection; poisoned output propagates into image prompts | `openai/chat_completion.go:281,286`, `image_generator_worker.go:212` |
| M | Stripe webhook test-signature bypass compiled into production; RC webhook non-constant-time compare + panics on missing env var (unauthenticated route) | `subscription.go:126-130`, `http/revenuecat.go:23,26` |
| M | Login timing side-channel (bcrypt skipped on unknown user) — user enumeration | `service/user.go:161-169` |
| M | 5-char min password on signup; reset/update paths enforce less; `BCRYPT_COST` override honored in prod | `http/user.go:25,36`, `common/utils.go:82-101` |
| M | Deprecated `dgrijalva/jwt-go` (CVE-2020-26160); `Environment` claim never verified | `go.mod`, `service/user.go:80` |
| M | Dev CORS reflects any origin with credentials; fails **open** on unrecognized `ENV`; no security headers anywhere; missing `c.Abort()` in limits middleware | `http/midlewares.go:117-120,252-256` |
| M | Realtime telemetry: client-supplied token counts, unbounded `Turns` payload, no wordlist ownership check | `http/realtime_telemetry.go:14-85` |
| M | Auth cookie `Max-Age` set in milliseconds (~999 years); domain hardcoded to `localhost` | `http/user.go:252,255` |
| M | Static-auth token compared with `!=` (non-constant-time), guarding OpenAI-spend-triggering routes | `http/midlewares.go:91` |

### Performance & operations

| Sev | Issue | Location |
|---|---|---|
| C | Zero indexes on `word_definitions` (`word_id`, `definition_id`) — seq scans in every quiz/analytics query. **Highest-ROI fix in the backend** | migration `000004` |
| C | `ORDER BY RANDOM()` over the global corpus per quiz question; no index on `definitions(part_of_speech)` | `repository/definition.go:126,163` |
| C | All 4 OpenAI HTTP clients: no timeout, no context, status codes ignored — hung calls pin workers/transactions forever; 429s unhandled | `openai/chat_completion.go:313`, `image_generation.go:51`, `text_to_audio.go:83`, `realtime.go:78` |
| C | Worker DB pool remains 10; definition generation and its extracted meaning-audio queue are capped at 2 each, preserving the prior four-worker provider budget and nominal River ceilings of 189 with legacy rollback queues or 179 in IAP-only mode. No provider/storage I/O remains in the definition transaction; measured capacity planning remains required before increasing load | `cmd/workers/main.go`, `river.go`, `definition_fetcher_worker.go`, `meaning_audio_worker.go` |
| C | Production graceful shutdown inverted — prod sleeps 5s and exits without `srv.Shutdown`; in-flight requests killed on deploy | `cmd/api/server.go:64-77` |
| H | No `/healthz`/`/readyz` endpoints — LBs can't gate deploys or detect dead pools | `http/setup.go` |
| H | Recursive-CTE streak query (≤365 iterations per wordlist) on the dashboard path; 60s premium TTL + per-answer cache wipe + no singleflight = stampedes | `analytics/batch_progress.go:82-109`, `http/analytics.go:26-31` |
| H | Non-sargable `DATE(created_at)` predicate + missing `(user_id, wordlist_id, created_at)` index — runs inside the quiz-answer write tx | `analytics/learning_progress.go:44-50`, migration `000030` |
| M | Expensive 4-table diagnostic query on every "no due items" response (the engaged-user steady state) | `leitner_system_strategy.go:278-291` |
| M | Quiz selection materializes 50 full definitions (19-column GROUP BY + JSON_AGG) to use 1; practice-ahead fallback re-runs the whole query unindexed | `leitner_system_strategy.go:104,124-181,256-266` |
| M | Unbounded list/batch/telemetry and secret-serialization paths. Closed locally by `API-BOUND-1` with stable pagination, explicit caps, ownership checks, telemetry bounds, and JSON secret redaction | `http/word.go`, `http/wordlist.go`, `http/realtime_telemetry.go`, `repository/user.go` |
| M | Per-upload MinIO client construction and ambiguous object/reference cleanup. Closed locally by `UPLOAD-1` with one startup client, server-owned metadata, and durable version-aware reconciliation | `common/minio.go`, `common/profile_image.go`, `service/profile_upload_reconciliation_worker.go` |
| M | 2s global timeout on all authenticated routes, and the middleware doesn't cancel work — timed-out requests still run to completion | `http/setup.go:78`, `http/midlewares.go:136-161` |
| M | pgBouncer URL swapped in but pgx not configured for transaction pooling (prepared-statement errors); documented `DISABLE_PREPARED_STATEMENTS` flag doesn't exist | `cmd/api/server.go:24-26`, `common/database.go:47-59` |
| M | Migrations without `lock_timeout`/`statement_timeout`; non-concurrent index builds; `sslmode=disable` hardcoded | `cmd/migrate/main.go:48,78` |
| M | Redis init failure memoized forever (caching disabled until redeploy); no dial/read/write timeouts | `common/redis.go:23-67` |
| M | No `MaxAttempts`/`JobTimeout`/`UniqueOpts` on any River job — 25 default retries × paid API calls on every panic bug | `river.go:126-145`, `job_service.go` |
| M | Renewal-reminder scheduling in a bare goroutine from the periodic-job factory — no recover, duplicated per replica (duplicate emails when scaled) | `river.go:83-88` |
| M | `runtime.ReadMemStats` (stop-the-world) on every job execution | `common/worker_context.go:60-61` |
| M | Docker: runs as root, unpinned `alpine:latest`; compose has no healthchecks/limits; default MinIO creds in `.env.example`; leftover test/placeholder migrations | `Dockerfile`, `docker-compose.yml`, migrations `000022`,`000048` |

# Mobile App Revamp Strategy — Top 20 Priorities

**Date:** 2026-08-03
**Scope:** Mobile app (`/mobile`) + supporting API changes (`/api`)
**Goal:** Increase install→paid conversion, user retention, and engagement through a coordinated feature + design revamp.

This document is the output of a deep audit of the mobile codebase across three dimensions: the conversion funnel (install → signup → paywall → purchase), retention & engagement mechanics (notifications, streaks, learning loop, SRS surfacing), and design system / UX quality. Roughly 60 individual findings were consolidated into the 20 prioritized workstreams below.

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

## Appendix A — Small bugs worth fixing regardless of strategy

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

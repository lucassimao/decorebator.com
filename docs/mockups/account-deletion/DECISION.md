# Account deletion compliance surface — decision record

Status: **production implementation approved**

## Candidate selected by Codex

A calm two-route deletion ledger in the approved warm editorial direction:

1. Direct in-app deletion for users who can sign in.
2. A public email-request route for users without app access.

The subscription warning precedes both routes because deleting a Decorebator account cannot cancel billing controlled by Apple or Google. The page avoids a password form and asks users never to send passwords, receipts, tokens, or payment details.

## Alternatives rejected before review

- **Authenticated web deletion form:** rejected because the marketing web application has no authenticated account surface and adding one would widen the security and API scope solely for compliance.
- **Unauthenticated email-input form backed by a new API:** rejected because it would collect personal data, require abuse controls and identity verification, and create a second deletion workflow when the existing support email already provides a human verification route.
- **Privacy-policy paragraph only:** rejected because the store requirement calls for a dedicated public resource where account deletion can be requested, and a buried policy paragraph is difficult for users and reviewers to find.
- **Danger-red confirmation page:** rejected because this page explains routes and outcomes; the destructive confirmation remains inside the authenticated app.

## Acceptance evidence produced

- Happy-path in-app instructions, no-app-access route, subscription warning, data outcomes, legal links, and support contact are represented.
- Axe Core 4.10.3 WCAG A/AA scan: zero violations.
- 390×844 and 430×932 specimens: no visible horizontal overflow.
- Semantic header/main/footer landmarks, one H1, ordered H2 sections, labeled warning, keyboard-visible focus, and 48px email action.
- No animation is used; reduced-motion behavior is identical.

## Claude review status

On 2026-08-05 the required order was followed:

1. `fable` started first and produced no output through the bounded review window.
2. `opus` returned a temporary session-limit error after Fable was stopped.
3. `sonnet` returned the same temporary session-limit error.

No model supplied substantive findings or approval. Retry with `fable` first after the reported 18:00 America/Fortaleza reset. Production implementation is not authorized by this decision record until the review reaches `APPROVED`.

The review was retried on 2026-08-05 at 14:52 America/Fortaleza after first strengthening the plan's model-order rule. `fable`, then `opus`, then `sonnet` each immediately returned the same account-wide session-limit response with the 18:00 reset time. This was another availability check, not a review round and not approval; the next attempt still starts with `fable`.

After access returned at 21:25, the first repository-wide attempt again followed `fable` → `opus` → `sonnet`; each model remained active but produced no output through its bounded window. A minimal Fable probe succeeded, proving model access, so the review was restarted from `fable` with an exact six-file evidence package and the already-verified backend facts rather than an open-ended source crawl.

### Round 1 — Fable changes required

Fable accepted the visual direction, responsive structure, warning placement, hedged retention language, focus treatment, recorded Axe result, and two-route model, but required three prototype corrections:

1. Put a visible, selectable `privacy@decorebator.com` address inside Route 02 so the path still works when `mailto:` does not.
2. Explain that a human verifies account ownership and that verified requests follow the privacy-policy timeframe.
3. Promote the subscription warning title from `<strong>` to `<h2>` so heading navigation encounters it before either route.

All three were implemented without changing the selected layout or visual direction. The optional lede tension and footer-link consistency were also corrected. Refreshed 390×844 and 430×932 specimens have no horizontal overflow; Axe Core 4.10.3 again reports zero WCAG A/AA violations; the heading outline is now H1 followed by the warning, both routes, and outcome as H2 peers. The same Fable review reconciled the actual files and returned `APPROVED` with no new material objection.

## Expected production files after approval

- `web/src/app/[locale]/delete-account/page.tsx`
- `web/src/app/[locale]/delete-account/layout.tsx`
- `web/src/components/home/FooterSection.tsx`
- `web/src/app/sitemap.ts`
- `web/messages/*.json`
- `mobile/app/profileSettings.tsx`
- `mobile/i18n/locales/*.json`
- focused web/mobile tests or static contract checks

## Production implementation evidence

The approved direction is implemented in the expected web and mobile files. The production web route was rendered at 390×844 and 430×932 with no horizontal overflow; refreshed captures are stored as `production-390.png` and `production-430.png`. Its heading outline is H1 → peer H2 sections → footer H2 with H3 groups, and Axe Core 4.10.3 reports zero violations. All seven web locales satisfy the static deletion-resource contract, the Next.js production build includes the localized route, and the mobile confirmation flow passes its focused unit test and TypeScript check.

Fable's implementation review found two material copy mismatches: the public directions named a nonexistent Profile settings row instead of Account, and the success alert still claimed permanent deletion despite the verified retention limitations. Both were corrected in every locale. The optional test-fidelity and destructive-label suggestions were also adopted: the unit test reads shipped English copy, and the final action uses an explicit localized delete verb. The same Fable reconciliation returned `APPROVED` with no remaining material objection. The single web Portuguese locale and English-only mail subject remain documented non-blocking localization limitations.

## Unresolved limitations

- Live store metadata and console deletion URLs are owner-only evidence.
- The database deletion path is being made atomic, but the audit found that public MinIO profile-picture objects are not currently removed and PostHog/Sentry/backup erasure or retention is not operationally defined. The plan now blocks store sign-off until that inventory is reconciled; the prototype deliberately says “removed from active use” rather than promising immediate deletion from every system.
- Legal/privacy retention language needs to remain consistent with the verified backend and third-party lifecycle plus any counsel-approved policy; this prototype does not replace that review.
- The web request route depends on `privacy@decorebator.com` remaining monitored and on the operator verifying account ownership before deletion.

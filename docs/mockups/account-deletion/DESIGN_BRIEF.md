# Account deletion compliance surface — design brief

## Purpose

Provide the public deletion resource required by Google Play while making the existing in-app deletion path, subscription cancellation boundary, data outcome, and support route unambiguous.

## Audience

- A signed-in learner who can delete directly in the mobile app.
- A learner who cannot access the app and needs a web request route.
- Apple and Google reviewers verifying account-deletion behavior.

## Required content and behavior

- Lead with the direct in-app path: **Settings → Profile → Delete account**.
- Offer an accessible email request route for users who cannot sign in.
- State that deleting a Decorebator account does not cancel an Apple or Google subscription; cancellation happens in the relevant store first.
- Explain at a high level what is deleted and what may be retained for legal, security, or transaction-record obligations without promising an unsupported exact schedule.
- Link to privacy and terms pages.
- Avoid asking for passwords, receipts, purchase tokens, or payment information.
- Use a calm, consequential tone rather than a marketing or destructive-danger aesthetic.

## Visual direction

Use the revamp's warm editorial study-desk language: cream paper, ink text, coral action accents, ruled dividers, and a small bookmark motif. The memorable element is a clear two-route decision ledger rather than a generic support form.

## Acceptance states

- 390×844 and 430×932 layouts have no horizontal overflow.
- Heading order and landmarks are semantic.
- Keyboard focus is visible and follows reading order.
- Text and controls meet WCAG 2.2 AA contrast.
- The email action has a text alternative and does not imply automatic deletion.
- The subscription warning appears before either deletion route.
- No animation is required; reduced-motion behavior is therefore identical.

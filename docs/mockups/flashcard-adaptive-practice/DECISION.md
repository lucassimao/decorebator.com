# Adaptive flashcard practice decision

Status: `APPROVED` after revision 2 and same-thread Opus reconciliation.

## Problem being validated

- The legacy card flip and slide animate even when the operating system requests reduced motion.
- The fixed 400-point card competes with the header, save-position row, progress, navigation, and safe areas on short phones; large text can crowd controls outside the viewport.
- Flashcard controls and loading/definition errors mix localized visible copy with hardcoded English accessibility and recovery copy.
- A terminal word-load error offers only Back even though the query already has an explicit retry mechanism.

## Proposed production contract

- Keep the header, progress, and navigation stable; cap the wordlist title at two lines, let the card region consume the remaining height down to 220 points, and scroll long definition content inside the card. If safe areas plus scaled chrome leave less than 220 points, the whole practice column becomes vertically scrollable with navigation reachable below the card; content must never be clipped by the screen container.
- Migrate the touched flip/slide/press feedback to Reanimated so one `useReducedMotion()` source can render transitions immediately without rotation, scale, or horizontal travel.
- Preserve the final-card completion contract and its code-native deck animation; this item changes only in-session practice and recovery states.
- Localize every visible and accessibility-only flashcard string in all supported locales. Card, audio, report, save-position, retry, and back controls expose roles, disabled/checked state, and content-aware labels without reading decorative icons. The front uses separate sibling flip and audio controls so the audio action is not swallowed by an accessible parent.
- Exactly one card face participates in accessibility and input at a time. The inactive face is hidden with `accessibilityElementsHidden` and `importantForAccessibility="no-hide-descendants"` (plus disabled pointer events); focus moves to the newly visible face only when screen-reader focus would otherwise be stranded.
- Loading uses a named status; the ten-second timeout exposes Try again and Back. A terminal online/network error exposes the same recovery actions without claiming a saved position. Offline-without-entitlement remains explanatory and non-destructive, then offers Back because retrying without connectivity cannot succeed. Empty wordlists keep distinct “No flashcards yet” copy and Back; definition-processing errors keep their existing processing explanation plus Try again and Back. These states must not collapse into the generic network error.
- The back face owns an internally scrollable definition region with a visible platform scroll indicator so 155% text does not move navigation off-screen. No text container uses a fixed content height. Wordlist titles stop after two lines, and long words wrap while preserving at least 155% system text scaling without horizontal overflow.

## Motion specification

- Flip: 420 ms rotation with an explicit 900-point perspective and an ease-out curve; interruption reverses from the current value.
- Card navigation updates the index immediately, then runs a 160 ms, 12-point directional settle on the new content. Repeated navigation cancels and restarts from the current value; controls are never gated for animation, preserving rapid deck browsing.
- Press feedback: short spring with no layout properties animated.
- Reduced motion: flip, navigation settle, and press feedback set their final state immediately; no rotation, scale, translation, stagger, spinner rotation, or decorative substitute. Input remains available throughout. The prototype's “Force reduced motion” control can only add reduction; it cannot override an operating-system reduced-motion preference, and its visible label reports the forced state.
- Unmount cancels active worklets; loading, error, and offline states never animate indefinitely.

## Prototype states

- 390×844 normal text, front card.
- 360×640 compact-height viewport at 155% text, with separate front/audio and internally overflowing back-card specimens.
- 430×932 recoverable terminal load error.
- Initial loading, ten-second timeout, definition-processing, offline-unavailable, empty, and recoverable online-error states.
- Light and dark themes, keyboard focus, animated flip, forced reduced-motion end state, and a 360px no-horizontal-overflow check.

## Copy and token contract

- New recovery, control, and accessibility-only strings are added to all eight locale files; existing `cardCounter`, `tapToFlipBack`, and `savePosition` translations are retained unless the production UI visibly adopts replacement wording everywhere.
- Shipping code uses the approved semantic theme roles (`action`, `onAction`, `surface`, `controlBorder`, text, success, danger, disabled, focus). Prototype canvas and muted combinations are evidence-only and do not become hardcoded native colors.

## Browser evidence

- Axe-core 4.10.3 reports zero violations for WCAG 2 A/AA, 2.1 AA, and 2.2 AA checks. At 360px, document width equals viewport width with no horizontal overflow. Forced reduced motion reports a `0.000001s` card transition and spinner animation.
- The compact back definition region is 305px high with 781px of scroll content at a computed 24.8px body size (155% of the 16px base). The compact front computes its long word at 47.12px versus the normal 32px and preserves an independently focusable 52px audio control.
- `specimen.png` captures the Axe result, compact front/back, visible definition scrollbar, and complete loading/error/offline/empty state gallery. `specimen-dark.png` covers the same matrix in dark mode, while `specimen-360-large-text.png` isolates the actual 360×640 overflow specimen. Keyboard Tab focus uses a visible solid outline; inactive faces carry both `inert` and `aria-hidden`.

## Claude review reconciliation

- Round 1 (`fable` temporarily usage-limited, then `opus`) rejected the draft for false large-text evidence and inverted hierarchy, no scaled front/audio state, focusable hidden faces, an unquantified clipping fallback, unbounded header wrapping, absent scroll evidence, conflated recovery states, untrue saved-position copy, and unevidenced/gated navigation motion.
- Revision 2 resolves all ten findings in the contract and prototype: true inherited 155% content scaling, separate compact front/back specimens, sibling flip/audio controls, synchronized inactive-face removal, a quantified 220-point card floor plus whole-column scroll fallback, two-line headers, real overflowing definitions, six distinct async/empty states, truthful network copy, and an interruptible non-gating 160ms navigation settle.
- The same Opus thread rechecked every original finding and motion concern against revision 2 and returned `APPROVED` with no remaining material objection.

## Expected production files

- `mobile/app/flashcard.tsx`
- `mobile/components/flashcard/FlashcardContent.tsx`
- `mobile/components/flashcard/FlashcardHeader.tsx`
- `mobile/components/flashcard/FlashcardLoadingState.tsx`
- `mobile/components/LoadingWithTimeout.tsx`
- `mobile/i18n/locales/*.json`
- focused component/session tests and the validation evidence note

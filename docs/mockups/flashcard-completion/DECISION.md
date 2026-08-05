# Flashcard completion UX decision

## Selected interaction

Replace the current terminal disabled arrow with a real end state:

- Every card uses stable navigation geometry: a 56pt previous control and a wide labeled action. The action reads `Next card` until the final card, then `Finish wordlist`.
- Finishing swaps the card content for an in-route, scroll-safe completion screen. It reports only the number of unique card indices viewed in this visit; it does not claim recall, accuracy, or mastery.
- `Review again` resets the index and visit count to the first card. `Back to wordlists`, the completion close control, Android system back, and iOS back navigation all leave the practice route through `router.back()`.
- `Return to last card` is a quiet recovery action for an accidental finish: it restores the card state without changing the index or per-visit Set. Finishing again in that same pass does not replay the deck entrance; a second celebration requires completing a new pass after choosing `Review again`.

## State and persistence contract

- Resolve the saved-position read before initializing the per-visit `Set<number>` so a resumed session starts with exactly its restored index counted once.
- Add each displayed index to that Set; the receipt copy is `{{count}} cards viewed this visit`, including the singular translation.
- `Review again` sets `currentIndex` to `0`, seeds the visit Set with `0`, clears completion state, resets flip/definition state, and explicitly persists position `0` when save-position is enabled. When save-position is disabled, no position key is created.
- Finishing does not fabricate a new saved index. A user who exits completion and later returns still resumes according to the existing save-position preference.

## Layout and accessibility contract

- The completion content uses a `ScrollView` with `flexGrow: 1` and centered content, preserving top and bottom safe-area insets under large text and translated copy.
- On entry, move accessibility focus to the localized `Wordlist complete` heading after layout; that focus change is the sole announcement mechanism. Do not add a live region or a second manual announcement. The decorative card stack is hidden from the accessibility tree.
- The navigation progress bar exposes `progressbar` semantics and current/min/max values. Buttons use localized labels without embedding decorative glyphs in their accessible names.
- Interactive outlines use the stronger semantic border role; the lighter divider role is not the sole boundary of a control.
- All new strings are added to every locale: next card, finish wordlist, wordlist complete, end message, cards viewed this visit, review again, back to wordlists, and return to last card. Remove the false `navigationHint` swipe copy from every locale.
- The navigation action allows two lines and is verified with the longest translation at 1.3x text scale on a 360pt phone. The first-card previous control exposes a disabled state.
- The progress bar uses min `0`, max `totalWords`, current `currentIndex + 1`, and a localized `Flashcard progress` label, including the one-word wordlist case.
- Move the navigation component from its hardcoded palette to `useTheme()` and consume the approved semantic action, on-action, surface, strong-border, text, and disabled roles. Prototype-only canvas/muted pairings are not promoted as shipping token combinations.

## Motion contract

- Implement the completion illustration as three theme-aware blank native card views with a small action-colored bookmark in a dedicated Reanimated component; do not add a raster asset or mix score/correctness/in-progress imagery into the completion message.
- On each completed pass, two back cards fan into place and the front card settles with a restrained spring, followed by a short copy fade/translate. The sequence is interruption-safe and cannot replay merely by navigating backward because completion has no backward-to-card transition.
- `useReducedMotion()` renders the assembled deck and fully visible copy immediately with no stagger, spring, scale, or translation.
- The existing legacy card flip/slide animation can remain isolated in the current card component; the new completion component uses Reanimated only.

## Visual states verified by the prototype

- Regular and final-card navigation with identical geometry.
- Completion in light and dark themes.
- 360px viewport without horizontal overflow.
- Compact-height completion with internal scrolling rather than clipping.
- Browser reduced-motion end state and visible keyboard focus.

## Image-generation decision

No generated raster illustration is used. The small card-deck mark is clearer as code-native geometry, adapts to dark mode and reduced motion, stays sharp at every density, and avoids introducing an asset for a state that carries no content-specific illustration.

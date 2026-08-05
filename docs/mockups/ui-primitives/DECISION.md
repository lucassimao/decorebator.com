# UI primitive layer decision

## Selected direction

Use the **warm editorial study desk** direction shown in `index.html` and the three accepted captures:

- `specimen-light.png`
- `specimen-dark-sheet.png`
- `specimen-compact.png`

The system retains the existing warm paper and coral brand character, separates decorative coral from the AA-compliant action role, uses the already bundled Space Mono Regular for restrained editorial display text, and uses the platform sans for body copy. Production primitives are native Expo components; the HTML is a disposable interaction/specification artifact.

## Primitive contract

- `Button`: primary, on-ink, secondary, and quiet variants; 44pt minimum target, visible focus, pressed spring/scale, loading text with an announced busy state, and a readable dedicated disabled palette.
- `Card`: semantic surface/border roles, continuous corner treatment, optional emphasis without per-screen color invention.
- `Input`: label, hint, error, disabled, and focused states; 48pt control height; programmatic error association and no color-only errors.
- `Sheet`: native modal/form-sheet behavior where possible; safe-area bottom padding, bounded scroll content, modal accessibility isolation, Escape/back dismissal, and focus restoration on web-compatible surfaces.
- Typography and spacing: named semantic text styles backed by the explicit type scale; all component spacing uses the `4 / 8 / 12 / 16 / 20 / 24 / 32 / 48` scale.
- Motion: Reanimated only in production; press feedback under 180ms, sheet settling around 240ms, interruption-safe state, and a reduced-motion path that retains understandable busy/state feedback.

## Rejected alternatives

- Keep `#FF7B54` as the universal primary background with white text: rejected because that pair fails AA; coral remains a brand/on-ink accent while the light action role uses dark rust.
- Preserve the existing per-screen button/card/input styles: rejected because it perpetuates contrast, motion, dark-mode, and spacing drift.
- Use Georgia or another platform-only editorial face: rejected because Android would not preserve the design; no new font dependency is needed for this foundation.
- Use raster/image-generated decoration inside primitives: rejected because these forms must remain themeable, accessible, and code-native. Image generation remains available for later content illustrations where raster art is appropriate.
- Implement a custom animated modal everywhere: rejected in favor of native Router form sheets/modals when routing permits, with a shared fallback only for embedded flows.

## Review and iteration outcome

Review began with `fable`, which was temporarily unavailable at its usage limit, then moved to Claude `opus` under the required fallback protocol.

- Round 1 rejected the first prototype for ornament/copy overlap, competing primary treatments, aspirational token usage, non-portable type, small targets, missing safe-area handling, incomplete sheet accessibility, reduced-motion busy ambiguity, and weak focus indication.
- Round 2 confirmed those fixes and found three remaining implementation details: focus restoration ordering, residual spacing literals, and synthetic bold requested from a regular-only font.
- Round 3 verified all findings were corrected and returned `APPROVED` with no material objection.

## Acceptance states checked

- Light and dark themes.
- Regular and 330px compact phone widths.
- Desktop specimen and 360px browser viewport with `scrollWidth === clientWidth`.
- Primary, on-ink, secondary, quiet, pressed, loading, disabled, focus, input hint, input error, card, progress, due badge, and sheet states.
- Closed sheet absent from the accessibility tree; open sheet isolates underlying app controls, traps focus, and restores focus to the trigger when closed.
- Critical contrast pairs measured from 4.59:1 to 7.75:1; focus colors measured above 6.6:1 against their surfaces.
- Reduced-motion media behavior keeps visible/announced busy text while removing spinner and transform motion.

## Production files expected to change

- `mobile/contexts/ThemeContext.tsx`: semantic colors, typography, geometry, spacing, and motion tokens while preserving compatibility during migration.
- `mobile/components/ui/button.tsx`
- `mobile/components/ui/card.tsx`
- `mobile/components/ui/input.tsx`
- `mobile/components/ui/sheet.tsx`
- `mobile/components/ui/text.tsx`
- `mobile/components/ui/index.ts`
- Co-located primitive tests under `mobile/components/ui/__tests__/`.
- `mobile/docs/ui-design-system.md`: make documentation match the implemented contract.

This milestone builds the primitives and tests only. Existing screens migrate incrementally in their own roadmap items; it does not introduce the bottom-tab information architecture or restyle all screens at once.

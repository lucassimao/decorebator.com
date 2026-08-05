# Native IAP paywall design brief

## Goal

Replace the RevenueCat/Stripe chooser with one native-store paywall that makes the premium value clear, uses only StoreKit/Google Play commerce metadata, and keeps the backend as the entitlement authority.

## Audience and moment

The screen opens from Settings or a premium-feature upsell. The user may be comparing plans, returning after a failed product load, or already entitled on another installation. The design should feel like extending a personal study desk rather than entering a generic checkout funnel.

## Required content and behavior

- Show only backend-allowlisted products that also resolve in the current store response.
- Render the store-returned localized title, display price, billing period, and eligible offer phases; build required renewal disclosure from localized app copy populated only with those store values. Never invent a fallback price or savings percentage.
- Keep monthly and annual plans distinguishable without preselecting or purchasing on card tap. Selection updates one explicit purchase button.
- Include loading, unavailable/retry, no-products, selected, purchasing handoff, restore entry, terms, privacy, and auto-renew disclosure states.
- Preserve a 44pt minimum target, screen-reader grouping, large-text scrolling, light/dark themes, and a 360px no-horizontal-overflow layout.
- Treat purchase completion, pending, restore results, and already-entitled states as the immediately following implementation item; this prototype includes a preview switch so their spatial impact can be reviewed before code is split into milestone commits.

## Visual direction

Extend the accepted warm editorial study-desk system. A small code-native “library shelf” composition is the memorable premium motif: three study cards settle into a shelf while a coral bookmark rises. It should remain themeable and accessible without requiring a raster asset.

## Motion contract for native implementation

- Use Reanimated 4 shared values and entering/layout transitions; CSS in this disposable artifact only demonstrates timing and choreography.
- On first ready render, settle the three cards with staggered spring transforms and reveal the bookmark over roughly 420ms total.
- Plan selection uses an interruption-safe 160–220ms border/scale transition and moves the selection mark without layout jumps.
- Product-state changes crossfade and translate by at most 8pt; the purchase button keeps its geometry while busy.
- Reduced motion removes stagger, parallax, rotation, and scale while preserving immediate state and busy-label changes.

## Review checklist

- Value hierarchy and honest plan comparison.
- Store-authored localized commerce/legal metadata is unmistakable.
- Selection is separate from purchase confirmation.
- Loading, unavailable, retry, no-product, selected, purchasing, pending, and entitled previews remain understandable.
- Restore and legal links are discoverable but do not compete with the primary action.
- Contrast, focus, target size, screen-reader semantics, large text, reduced motion, compact height, and dark mode.
- Whether the code-native motif is sufficient or an image-generated illustration would materially improve comprehension.

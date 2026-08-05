# Mobile UI design system

This document describes the implemented primitive foundation in `components/ui`. It is intentionally smaller than the final revamp: existing screens migrate incrementally as their roadmap items are touched.

The approved visual rationale and prototype evidence live in `docs/mockups/ui-primitives/DECISION.md`.

## Direction

The system uses a warm editorial study-desk character:

- warm paper and quiet surfaces for long learning sessions;
- dark ink for hierarchy and legibility;
- coral as a brand accent, not as an automatic white-text CTA background;
- dark rust for the light-theme action role and coral with dark ink in dark mode;
- Space Mono Regular for display/feature text and the platform sans for body text.

`expo-font` embeds Space Mono at build time. Do not request synthetic weights from this regular-only file.

The embedded face requires a new native store build. Do not migrate a screen to `feature` or `display` in an OTA-only release targeting binaries built before this plugin; the first adopting screen is gated on a compatible store build.

## Source of truth

`contexts/ThemeContext.tsx` owns all new tokens while retaining the older keys during screen migration.

### Semantic color roles

- `action` / `actionPressed` / `onAction`
- `brandAccent` / `onBrandAccent`
- `brandAccentPressed`
- `error`
- `emphasisSurface`
- `focus`
- `disabledBackground` / `disabledText`
- `scrim`

Never infer foreground color from a raw brand color. Use the matching `on*` role.

The approved critical pairs measure between 4.59:1 and 7.75:1. New role combinations still require contrast validation; semantic naming is not proof of contrast by itself.

### Spacing and geometry

Existing keys remain available: `xs=4`, `sm=8`, `md=16`, `lg=24`, `xl=32`, `xxl=48`.

The primitive layer adds `compact=12` and `comfortable=20`, plus:

- `geometry.touchTarget = 44`
- `geometry.controlHeight = 48`

Use spacing tokens for gaps and padding. Component dimensions such as touch targets belong to geometry tokens.

### Typography

Available variants are `micro`, `caption`, `label`, `small`, `body`, `heading`, `title`, `feature`, and `display`.

Use `feature` and `display` sparingly. Both use Space Mono Regular at weight 400. All text scales by default; avoid fixed-height text containers.

### Motion

- `fast = 130ms`
- `standard = 180ms`
- `sheet = 240ms`
- `pressScale = 0.975`

Production primitive motion uses Reanimated 4 only. Do not drive one component with both legacy `Animated` and Reanimated. Reduced-motion users keep understandable text/state feedback while transform/spinner motion is removed.

## Primitives

Import from `@/components/ui`.

### `UiText`

```tsx
<UiText variant="title">Spanish essentials</UiText>
<UiText variant="label" tone="action">7 words due</UiText>
```

Tones are `primary`, `secondary`, `inverse`, `danger`, and `action`. Font scaling is enabled unless a caller explicitly overrides it for a proven reason.

### `Button`

```tsx
<Button onPress={startQuiz}>Practice now</Button>
<Button variant="onInk" loading loadingLabel="Preparing quiz">
  Practice now
</Button>
```

Variants:

- `primary`: standard action role;
- `onInk`: coral brand action on a dark/inverse surface;
- `secondary`: bordered surface action;
- `quiet`: low-emphasis text action.

Loading makes the control disabled and sets an accessibility busy state. Under reduced motion the spinner is omitted; loading text is preferred, and a static mark remains when no loading label is supplied. Press motion is bounded and interruption-safe through shared Reanimated values. Use `leading` and `trailing` for icons; button text is intentionally a string so native views are never nested inside text.

### `Card`

```tsx
<Card variant="elevated" padding="comfortable">
  <UiText variant="heading">Today</UiText>
</Card>
```

Variants are `surface`, `elevated`, and `emphasis`; padding is `none`, `compact`, `default`, or `comfortable`.

### `Input`

```tsx
<Input
  label="Daily goal"
  value={goal}
  onChangeText={setGoal}
  hint="Choose a goal you can repeat."
  error={goalError}
/>
```

Labels are required. Errors change the border, provide visible text, populate the accessibility hint, and are connected through `aria-describedby` on supported surfaces. Error text uses an Android polite live region and an iOS accessibility announcement.

### `Sheet`

```tsx
<Sheet
  visible={visible}
  onClose={() => setVisible(false)}
  eyebrow="Practice setup"
  title="How long do you have?"
  returnFocusRef={triggerRef}
>
  <Button onPress={startQuiz}>7 due words</Button>
</Sheet>
```

The shared sheet is the embedded-flow fallback. Prefer Expo Router native `formSheet` or modal routes when the interaction is naturally a route.

The fallback:

- handles Android back through `onRequestClose`;
- applies modal screen-reader isolation;
- restores accessibility focus when a trigger ref is supplied;
- accounts for the bottom safe area;
- scrolls tall/large-type content;
- cancels and replaces interrupted Reanimated transitions;
- removes itself after its exit settles.

## Accessibility acceptance

Every primitive or migrated use must preserve:

- 44pt minimum interactive targets;
- readable text at accessibility font sizes;
- labels for icon-only actions;
- non-color state communication;
- visible focus on keyboard/web surfaces;
- a meaningful reduced-motion state;
- safe-area and keyboard behavior;
- valid loading, disabled, error, empty, and retry states.

## Migration boundary

Do not mass-rewrite existing screens merely to adopt these components. Migrate a screen when its roadmap item changes that screen, and remove the superseded local styles in the same focused change. Legacy token files and legacy `Animated` users remain tracked debt until their owning roadmap items execute.

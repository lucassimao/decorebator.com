# Welcome Overlay Test Scenarios

## 1. First-Time Signup Flow
- **Preconditions**: User completes signup within the app; AsyncStorage keys `justSignedUp` set to `"true"`, `hasSeenDashboard` absent.
- **Steps**:
  1. Launch app and sign in with newly created account.
  2. Observe navigation to dashboard after splash screen.
  3. Dismiss welcome overlay using `Skip`.
- **Expected**:
  - Overlay is visible immediately on first dashboard load.
  - After `Skip`, overlay disappears and `hasSeenDashboard` key persists with value `"true"`.
  - Overlay does not reappear on next dashboard visit (force-close & relaunch).

## 2. First-Time Signup + Create Wordlist
- **Preconditions**: Same as scenario 1.
- **Steps**:
  1. On overlay, tap `Create Your First Wordlist`.
  2. Complete wordlist creation flow.
  3. Close congratulations modal.
- **Expected**:
  - Create wordlist modal opens immediately after CTA tap.
  - Dashboard refreshes with new wordlist and congrats modal appears once.
  - `hasSeenDashboard` remains `"true"`; overlay does not reappear on subsequent visits.

## 3. Returning User With No Wordlists (Overlay Once)
- **Preconditions**: Existing user with no wordlists; AsyncStorage lacks both `justSignedUp` and `hasSeenDashboard`.
- **Steps**:
  1. Sign in and land on dashboard.
  2. Allow fallback timer (~1s) to elapse.
  3. Dismiss overlay via `Skip`.
  4. Immediately navigate away and return to dashboard.
- **Expected**:
  - Overlay appears once after the delay.
  - Dismissal sets `hasSeenDashboard` and prevents further overlays even while wordlist count remains zero.

## 4. User With No Wordlists Who Already Saw Overlay
- **Preconditions**: AsyncStorage `hasSeenDashboard` is `"true"`; no wordlists present.
- **Steps**:
  1. Sign in and navigate to dashboard.
  2. Stay on screen for ≥5 seconds.
- **Expected**:
  - Overlay never appears.
  - No additional AsyncStorage writes occur.

## 5. User With Existing Wordlists
- **Preconditions**: User account contains ≥1 wordlist; any AsyncStorage state.
- **Steps**:
  1. Sign in and land on dashboard.
- **Expected**:
  - Overlay is suppressed regardless of flags.
  - Dashboard content animates normally.

## 6. Overlay Responsiveness
- **Preconditions**: Device simulator/emulator set to small screen (e.g., iPhone SE or Android 360x640), overlay triggered (see scenarios 1 or 3).
- **Steps**:
  1. Observe overlay layout on portrait orientation.
  2. Scroll the feature list if needed.
- **Expected**:
  - Primary CTA and Skip remain visible without clipping.
  - Feature list scrolls smoothly; modal height never exceeds screen bounds.

## 7. Dismissal Persistence Across App Restarts
- **Preconditions**: Overlay previously dismissed, `hasSeenDashboard` present.
- **Steps**:
  1. Force quit the app.
  2. Relaunch and sign in.
- **Expected**:
  - Overlay stays hidden; AsyncStorage key still `"true"`.
  - No duplicate welcome overlay animation plays.

## 8. Analytics/Logging Smoke
- **Preconditions**: Dev build with console/log access; overlay triggered.
- **Steps**:
  1. Trigger overlay via scenario 3.
  2. Inspect console logs.
- **Expected**:
  - Dev logs show the fallback message exactly once per fresh session (`Dashboard welcome check`, `Fallback: Showing welcome...`).
  - No repeated logs after dismissal in the same session.

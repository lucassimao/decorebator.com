# Offline Manual Test Checklist

This checklist validates offline behavior across the core mobile flows.

## Setup

- Use a premium account and a free account.
- Ensure you have at least two wordlists on the premium account.
- Ensure the premium account has:
  - Quizzes taken in at least one wordlist.
  - Flashcards opened and definitions loaded for at least one wordlist.
- For offline simulation, disable Wi-Fi and mobile data in the emulator.

## Scenarios

1. Cold start offline (premium)

- Steps: Kill app, disable Wi-Fi/data, launch app.
- Expect: App loads without redirecting to signin; dashboard shows cached wordlists; offline banner visible.

2. Cold start offline (free)

- Steps: Kill app, disable Wi-Fi/data, launch app.
- Expect: App loads if session cached; offline banner shows free/premium messaging; offline-only features blocked.

3. Dashboard list cached

- Steps: Go online, open dashboard with multiple wordlists, then go offline, kill app, relaunch.
- Expect: Wordlists still listed; no empty dashboard.

4. Wordlist detail read-only

- Steps: Go offline, open a wordlist.
- Expect: Add/delete/learned actions disabled; words list visible; offline warning shown on empty state.

5. Offline preloader visibility (premium)

- Steps: Go online, open wordlist detail.
- Expect: "Download for offline" control visible and enabled.

6. Preload flow

- Steps: Tap "Download for offline" while online.
- Expect: Status shows downloading, then "Available offline" with cache progress.

7. Quiz offline (cached)

- Pre: While online, take at least 2 quizzes for a wordlist.
- Steps: Go offline, open quiz screen for that wordlist.
- Expect: Quiz loads; no network errors; results shown; answers are not synced.

8. Quiz offline (not cached)

- Steps: Choose a wordlist never quizzed; go offline; open quiz screen.
- Expect: Error state indicating no cached quiz available.

9. Flashcards offline (cached definitions)

- Pre: Online, open flashcards and flip cards to load definitions.
- Steps: Go offline, reopen flashcards.
- Expect: Words load; definitions load when flipped; audio plays if cached.

10. Flashcards offline (uncached definitions)

- Steps: Online, open flashcards but do not flip; go offline; open and flip.
- Expect: Definition fetch fails gracefully with offline messaging.

11. Cache expiry behavior

- Pre: Lower cache expiry temporarily or advance device time beyond expiry window.
- Steps: Go offline, open cached wordlist/quiz.
- Expect: Expired cache not used; UI shows offline unavailable state.

12. Connectivity flapping

- Steps: Toggle Wi-Fi/data on and off while on dashboard/quiz.
- Expect: Offline banner updates; app recovers cleanly when back online.

13. Sign-out cleanup

- Steps: Online, sign out; sign in with another account; go offline and open dashboard.
- Expect: No data from previous user appears; offline cache cleared.

## Notes

- If any scenario fails, capture logs and note device state (network type, emulator vs device, premium/free).
- Validate on both Android emulator and a physical device if possible.

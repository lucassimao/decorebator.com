

  🔄 Flow 4: New User Signup → Sign Out → Sign Back In

  Steps:
  1. Complete Flow 1 (new signup) if not done already
  2. Navigate to settings and sign out
  3. Sign back in with the same account
  4. Observe the journey

  Expected Behavior:
  - Brief simple loading spinner during auth check
  - 1.5-second dashboard skeleton with shimmer animation
  - NO welcome overlay (user has already seen it)
  - Dashboard shows normally

  Report: Did welcome overlay appear again? What was the skeleton behavior?

  ---
  🔄 Flow 5: Welcome Overlay Interaction Testing

  Steps:
  1. Trigger welcome overlay (through new signup or clearing app data)
  2. Test "Skip for now" button
  3. Test "Create Your First Wordlist" button
  4. Test creating wordlist through welcome flow

  Expected Behavior:
  - "Skip for now" → dismisses overlay, shows empty dashboard
  - "Create Your First Wordlist" → dismisses overlay, opens creation modal
  - Creating wordlist → special congratulations message → dashboard with wordlist

  Report: Did both buttons work correctly? What happened with wordlist creation?

  ---
  🔄 Flow 6: Skeleton Animation Quality

  Steps:
  1. Sign in as existing user (Flow 2)
  2. Focus specifically on the 1.5-second skeleton screen
  3. Observe the shimmer animation and visual quality

  Expected Behavior:
  - Skeleton elements should have smooth pulsing/shimmer effect
  - Good contrast against background (light/dark theme)
  - Skeleton should resemble actual dashboard layout
  - Animation should be smooth, not jarring

  Report: How did the skeleton look? Was the animation smooth? Good contrast?

  ---
  🔄 Flow 7: Empty Dashboard State (For Users With No Wordlists)

  Steps:
  1. Sign in with account that has NO wordlists
  2. OR delete all wordlists and refresh dashboard
  3. Observe empty state handling

  Expected Behavior:
  - Shows empty dashboard illustration
  - "Create your first wordlist" call-to-action
  - If no welcome was shown before, fallback welcome overlay should appear after 1 second

  Report: What happened with empty dashboard? Did fallback welcome trigger?

  ---
  📝 Reporting Template

  For each flow, please report:

  ## Flow X: [Name]
  ✅ PASSED / ❌ FAILED / ⚠️ PARTIAL

  **What I saw:**
  - Step 1: [describe]
  - Step 2: [describe]
  - etc.

  **Issues found:**
  - [list any problems]

  **Screenshots/Videos:**
  - [if possible, share screenshots of key steps]

  ---
  🐛 Special Things to Watch For

  1. Skeleton timing: Should be exactly 1.5 seconds, not too fast/slow
  2. Welcome overlay: Should appear for new users, not for returning users
  3. Android skeleton bug: Should NOT appear before signin screen
  4. Animation quality: Shimmer should be smooth and visible
  5. Theme compatibility: Test in both light and dark modes if possible
  6. Success messages: Special congratulations for first wordlist vs. normal success

  Take your time with testing and note even small details - they help identify what's working well and what needs adjustment! 🚀

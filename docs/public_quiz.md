# Decorebator – Public Quiz: Final Product Specification & Delivery Plan (Vision → MVP → Post‑MVP)

> **Purpose**: A single, tech‑agnostic specification that combines the vision, business rules, MVP scope, phased delivery plan, prompts for an implementation agent, and review governance. This is the authoritative reference for what we are building and how we’ll validate it.

---

## 1) Vision & Feature Overview

### 1.1 Ultimate Goal (North‑Star)

Enable teachers and premium users to transform their wordlists into **viral, playable quizzes** that spread across social networks, **attract new learners**, convert them into active Decorebator users, and sustain engagement through friendly competition and creator incentives.

**North‑Star outcomes**

* Consistent flow of new players from social links.
* Healthy conversion from guest play → signup → returning learner.
* A thriving creator ecosystem publishing high‑quality, on‑brand quizzes.

### 1.2 Product Pillars

1. **Frictionless Play** – Anyone can play from a link without creating an account.
2. **Share‑Native** – Every quiz looks great on social and invites friendly competition.
3. **Fair & Safe** – Clear rules, light anti‑cheat, and responsive moderation.
4. **Creator‑Centric** – Publishing is fast, analytics are understandable, incentives are tangible.
5. **Global‑ready** – Copy and flows localized for PT‑BR and EN at launch; extensible later.

### 1.3 Personas

* **Independent Teachers/Influencers (Creators)**: Quick, engaging content and visibility.
* **Casual Social Learners (Participants)**: Try and share light challenges while scrolling.
* **Competitive Learners (Existing Users)**: Chase ranks and return weekly to improve.
* **Schools/Clubs (Later)**: Run friendly competitions and homework challenges.

### 1.4 Value Proposition & Differentiators

* **From static posts to playable content**: Teachers already share vocabulary images; we turn those into tappable quizzes with leaderboards.
* **Zero‑friction trial**: First play is anonymous; excitement drives the signup.
* **Creator attribution**: Every shared quiz reinforces the creator’s brand and impact.
* **Lightweight, fair competition**: Weekly resets and simple rules make leaderboards approachable, not elitist.

### 1.5 Feature Set (End‑State)

**Creator Experience**

* Publish public quiz from a wordlist: **Title**, **Description**, **Difficulty**, **Time Limit** (1–15 min), **Question Count** (5–15)
* Share‑ready output: social link + preview image.
* Entitlements: Premium creators unlimited; Free creators **one** active public quiz.
* Analytics (Lite at MVP): Views, Plays, Completion rate, Average score, Top countries (7‑day + All‑time).

**Participant Experience**

* Landing: title, difficulty, estimated time, **Start** button.
* Quiz: timed sequence of questions; consistent feedback pattern (MVP: show results at end).
* Results: **Score (%)**, **Accuracy (%)**, **Completion time**, **Rank preview**.
* Conversion: Clear **Signup CTA** to save progress and compete weekly.

**Social & Viral Mechanics**

* Smart link & preview: each quiz renders a rich preview with current title/description.
* Result share: one‑tap message like “I scored 92%—can you beat me?”
* Weekly rhythm: weekly leaderboard and highlights to prompt re‑engagement.

**Leaderboards & Gamification**

* Two tabs: **All‑time Top‑50**, **This Week Top‑50** (resets **Mondays 00:00**, America/Fortaleza).
* Sorting: **Score (desc)** → **Completion time (asc)** for tie‑breaks.
* Guest display names: optional; profanity‑filtered.
* Post‑MVP: badges, streaks, friend/group leaderboards, seasonal challenges.
*

---

## 2) Business Rules & Policies

### 2.1 Roles & Permissions

* **Creator (Premium)**: publish unlimited public quizzes.
* **Creator (Free)**: limited to **one** active public quiz (can unpublish/swap which one is active).
* **Participant (Guest)**: play any public quiz; optionally submit a display name for leaderboards.
* **Participant (Logged‑in)**: same as guest, plus results saved to account; eligible for streaks/badges post‑MVP.
* **Moderator/Admin**: unpublish reported quizzes; view aggregated activity.

### 2.2 Publishing Rules

* **Immutable slug (MVP)** to preserve link integrity.
* Preview assets (image + texts) should reflect current title/description.

### 2.3 Participation & Attempts

* Anyone with the link can play without signup.
* If a quiz is **unpublished/suspended**, public access shows a friendly unavailable message.

### 2.4 Scoring & Ranking

* Score (%) = **correct answers / total questions × 100**.
* Accuracy (%) mirrors score for single‑choice; for mixed types, accuracy reflects all graded interactions.
* Ranking: **Score (desc)**; ties broken by **shorter completion time (asc)**.

###

### 2.8 Localization & Copy

* All user‑facing strings in **PT‑BR** and **EN** at MVP.
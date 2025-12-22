# Decorebator Web Application

The Decorebator web app is a Next.js 15 marketing site plus public-quiz experience. It consumes the Go API for public quizzes and generates localized, SEO-optimized pages.

## 🌟 Current Features

- **Marketing site** with landing, features, pricing, and testimonials.
- **Public quizzes** at `/q/[slug]` with quiz play + leaderboard (fetched from the API).
- **Internationalization** via `next-intl` (7 locales).
- **SEO + sitemaps** with dynamic inclusion of public quizzes at build time.

## 🧱 Tech Stack

- Next.js 15 App Router
- Tailwind CSS v4
- next-intl for i18n
- Chart.js for analytics previews

## 🚀 Getting Started

```bash
npm install
npm run dev
```

The dev server runs on port 4000.

## 🔧 Environment Variables

- `NEXT_PUBLIC_API_URL` — API base URL (used to fetch public quizzes)
- `STATIC_AUTHENTICATION` — optional token used to list public quizzes for the sitemap
- `SITE_URL` — base URL for sitemap generation

## 🗂️ Structure

```
web/
├── src/
│   ├── app/[locale]/...     # localized routes
│   ├── components/          # landing + quiz UI
│   └── lib/api.ts           # public quiz API client
├── messages/                # i18n translations
├── public/                  # static assets
└── next-sitemap.config.js   # sitemap config (public quizzes)
```

## 🧪 Commands

```bash
npm run dev
npm run build
npm run start
npm run lint
npm run format:check
```

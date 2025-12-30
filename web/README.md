# Decorebator Web Application

The Decorebator web app is a Next.js 15 marketing site with localized, SEO-optimized pages.

## 🌟 Current Features

- **Marketing site** with landing, features, pricing, and testimonials.
- **Internationalization** via `next-intl` (7 locales).
- **SEO + sitemaps** for marketing pages.

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

- `NEXT_PUBLIC_API_URL` — API base URL
- `SITE_URL` — base URL for sitemap generation

## 🗂️ Structure

```
web/
├── src/
│   ├── app/[locale]/...     # localized routes
│   ├── components/          # landing UI
├── messages/                # i18n translations
├── public/                  # static assets
└── next-sitemap.config.js   # sitemap config
```

## 🧪 Commands

```bash
npm run dev
npm run build
npm run start
npm run lint
npm run format:check
```

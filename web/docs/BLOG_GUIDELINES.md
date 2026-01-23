# Blog Post Guidelines

This document explains how to add new blog posts to the Decorebator marketing website.

## Architecture Overview

The blog uses a **file-based CMS** approach:

```
web/src/
├── content/
│   └── blog.ts              # All blog content defined here
└── app/[locale]/blog/
    ├── page.tsx             # Blog list page (auto-renders all posts)
    └── [slug]/page.tsx      # Blog post page (auto-renders by slug)
```

**Key principle:** Add content to `blog.ts` only. The pages handle rendering automatically.

## Adding a New Blog Post

### Step 1: Edit `web/src/content/blog.ts`

Add a new object to the `blogPosts` array:

```typescript
export const blogPosts: BlogPost[] = [
  // New post goes here (newest first)
  {
    slug: 'your-post-slug', // URL-friendly, lowercase, hyphens
    date: '2026-01-22', // ISO date format YYYY-MM-DD
    readingMinutes: 3, // Estimated read time
    tags: ['Feature', 'Learning'], // 1-4 short tags
    author: {
      name: 'Decorebator Team', // Author display name
      role: 'Product', // Author role/title
    },
    cover: {
      eyebrow: 'New feature', // Short label above title
      gradient: 'from-primary-500 via-amber-500 to-rose-500', // Tailwind gradient
      image: '/images/blog/your-image.jpg', // Optional cover image
    },
    content: {
      en: {
        title: 'Your Post Title',
        description: 'A brief summary shown in cards and meta tags.',
        excerpt: 'Short excerpt for sidebar widgets.',
        sections: [
          // Content sections go here
        ],
      },
    },
  },
  // ... existing posts
]
```

### Step 2: Write Content Sections

The `sections` array supports these types:

#### Paragraph

```typescript
{ type: 'paragraph', text: 'Your paragraph text here.' }
```

#### Heading (h2)

```typescript
{ type: 'heading', text: 'Section Heading' }
```

#### Bullet List

```typescript
{
  type: 'list',
  items: [
    'First bullet point',
    'Second bullet point',
    'Third bullet point',
  ]
}
```

#### Callout Box

```typescript
{
  type: 'callout',
  title: 'Tip',           // Short title like "Tip", "Note", "Warning"
  text: 'Callout content here.'
}
```

### Step 3: Add Cover Image (Optional)

1. Place image in `web/public/images/blog/`
2. Recommended size: 1200x630px (OG image ratio)
3. Formats: JPG, PNG, WebP
4. Reference as `/images/blog/filename.jpg`

If no image is provided, a gradient background is used instead.

## Content Guidelines

### Slug

- Lowercase letters, numbers, and hyphens only
- Keep under 50 characters
- Be descriptive: `new-audio-quiz-mode` not `post-1`

### Title

- Clear and specific
- 60 characters max for SEO
- Include key feature/topic name

### Description

- 1-2 sentences summarizing the post
- 160 characters max for SEO meta description
- Appears on blog cards and social shares

### Tags

- Use 2-4 tags per post
- Capitalize first letter: `Learning`, `Quiz modes`, `Audio`
- Reuse existing tags when possible

### Sections Structure

- Start with an intro paragraph
- Use headings to break up content
- Lists for feature benefits or steps
- End with a callout for tips or CTAs

## Example Post Structure

```typescript
sections: [
  { type: 'paragraph', text: 'Intro explaining what the post is about...' },
  { type: 'heading', text: 'Why this matters' },
  { type: 'list', items: ['Benefit 1', 'Benefit 2', 'Benefit 3'] },
  { type: 'heading', text: 'How it works' },
  { type: 'paragraph', text: 'Explanation of the feature...' },
  { type: 'list', items: ['Step 1', 'Step 2', 'Step 3'] },
  { type: 'callout', title: 'Tip', text: 'Pro tip for users...' },
]
```

## Automatic Features

When you add a post to `blog.ts`, these features work automatically:

- **Blog list page** displays the new post card
- **Static generation** pre-renders the post page
- **SEO metadata** (title, description, OG tags)
- **Article structured data** (JSON-LD for Google)
- **Social share buttons** (Twitter, LinkedIn, Email)
- **CTA card** at end of post
- **Sitemap** includes the new post URL

## Localization (Future)

Currently English-only. To add translations later:

```typescript
content: {
  en: { title: 'English Title', ... },
  es: { title: 'Spanish Title', ... },
  fr: { title: 'French Title', ... },
}
```

The system automatically falls back to English if a locale is missing.

## Testing Your Post

1. Run `npm run dev` in the `web` directory
2. Visit `http://localhost:4000/en/blog`
3. Verify your post card appears
4. Click through to the full post
5. Check all sections render correctly
6. Test share buttons work

## Checklist

Before publishing, verify:

- [ ] Slug is URL-friendly and unique
- [ ] Date is correct (YYYY-MM-DD format)
- [ ] Reading time is reasonable estimate
- [ ] Tags are relevant and properly capitalized
- [ ] Title is clear and under 60 characters
- [ ] Description summarizes the post well
- [ ] All sections render correctly
- [ ] Links work (if any)
- [ ] Cover image (if used) loads properly

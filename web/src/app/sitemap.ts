import type { MetadataRoute } from 'next'
import { routing } from '../i18n/routing'
import { blogPosts } from '../content/blog'

const baseUrl = 'https://decorebator.com'

type ChangeFrequency = 'always' | 'hourly' | 'daily' | 'weekly' | 'monthly' | 'yearly' | 'never'

function buildAlternates(path: string): Record<string, string> {
  const alternates: Record<string, string> = {
    'x-default': `${baseUrl}/en${path}`,
  }
  routing.locales.forEach((locale) => {
    alternates[locale] = `${baseUrl}/${locale}${path}`
  })
  return alternates
}

function createEntry(
  locale: string,
  path: string,
  changeFrequency: ChangeFrequency,
  priority: number,
  lastModified: Date = new Date()
): MetadataRoute.Sitemap[number] {
  return {
    url: `${baseUrl}/${locale}${path}`,
    lastModified,
    changeFrequency,
    priority,
    alternates: {
      languages: buildAlternates(path),
    },
  }
}

export default function sitemap(): MetadataRoute.Sitemap {
  const locales = routing.locales
  const urls: MetadataRoute.Sitemap = []

  // Root URL with alternates pointing to localized versions
  urls.push({
    url: baseUrl,
    lastModified: new Date(),
    changeFrequency: 'weekly',
    priority: 1,
    alternates: {
      languages: buildAlternates(''),
    },
  })

  locales.forEach((locale) => {
    // Home page
    urls.push(createEntry(locale, '', 'weekly', 1))

    // Static pages
    urls.push(createEntry(locale, '/help', 'monthly', 0.4))
    urls.push(createEntry(locale, '/blog', 'weekly', 0.6))
    urls.push(createEntry(locale, '/privacy', 'yearly', 0.3))
    urls.push(createEntry(locale, '/terms', 'yearly', 0.3))
    urls.push(createEntry(locale, '/delete-account', 'yearly', 0.4))

    // Blog posts
    blogPosts.forEach((post) => {
      urls.push(createEntry(locale, `/blog/${post.slug}`, 'monthly', 0.6, new Date(post.date)))
    })
  })

  return urls
}

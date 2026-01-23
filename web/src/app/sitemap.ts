import type { MetadataRoute } from 'next'
import { routing } from '../i18n/routing'
import { blogPosts } from '../content/blog'

const baseUrl = 'https://decorebator.com'

export default function sitemap(): MetadataRoute.Sitemap {
  const locales = routing.locales
  const urls: MetadataRoute.Sitemap = []

  urls.push({
    url: baseUrl,
    lastModified: new Date(),
    changeFrequency: 'weekly',
    priority: 1,
  })

  locales.forEach((locale) => {
    const localeBase = `${baseUrl}/${locale}`
    urls.push(
      {
        url: localeBase,
        lastModified: new Date(),
        changeFrequency: 'weekly',
        priority: 1,
      },
      {
        url: `${localeBase}/help`,
        lastModified: new Date(),
        changeFrequency: 'monthly',
        priority: 0.4,
      },
      {
        url: `${localeBase}/blog`,
        lastModified: new Date(),
        changeFrequency: 'weekly',
        priority: 0.6,
      },
      {
        url: `${localeBase}/privacy`,
        lastModified: new Date(),
        changeFrequency: 'yearly',
        priority: 0.3,
      },
      {
        url: `${localeBase}/terms`,
        lastModified: new Date(),
        changeFrequency: 'yearly',
        priority: 0.3,
      }
    )

    blogPosts.forEach((post) => {
      urls.push({
        url: `${localeBase}/blog/${post.slug}`,
        lastModified: new Date(post.date),
        changeFrequency: 'monthly',
        priority: 0.6,
      })
    })
  })

  return urls
}

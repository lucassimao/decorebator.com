import type { MetadataRoute } from 'next'
import { routing } from '../i18n/routing'

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
  })

  return urls
}

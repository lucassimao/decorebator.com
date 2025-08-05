/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: process.env.SITE_URL || 'https://decorebator.com',
  generateRobotsTxt: false, // We have a custom robots.txt
  exclude: ['/server-sitemap.xml'], // exclude dynamic sitemap from static sitemap
  generateIndexSitemap: false,
  changefreq: 'daily',
  priority: 0.7,
  sitemapSize: 5000,

  // Multi-language support
  alternateRefs: [
    {
      href: 'https://decorebator.com/en',
      hreflang: 'en',
    },
    {
      href: 'https://decorebator.com/es',
      hreflang: 'es',
    },
    {
      href: 'https://decorebator.com/fr',
      hreflang: 'fr',
    },
    {
      href: 'https://decorebator.com/de',
      hreflang: 'de',
    },
    {
      href: 'https://decorebator.com/it',
      hreflang: 'it',
    },
    {
      href: 'https://decorebator.com/pt',
      hreflang: 'pt',
    },
    {
      href: 'https://decorebator.com/ja',
      hreflang: 'ja',
    },
  ],

  // Custom transformation for better SEO
  transform: async (config, path) => {
    // Custom priority based on path importance
    let priority = 0.7
    let changefreq = 'weekly'

    if (path === '/') {
      priority = 1.0
      changefreq = 'daily'
    } else if (path.includes('/en') || path.includes('/es') || path.includes('/fr')) {
      priority = 0.9
      changefreq = 'daily'
    } else if (path.includes('/terms') || path.includes('/privacy')) {
      priority = 0.3
      changefreq = 'monthly'
    } else if (path.includes('/help')) {
      priority = 0.6
      changefreq = 'weekly'
    }

    return {
      loc: path,
      changefreq,
      priority,
      lastmod: new Date().toISOString(),
      alternateRefs: config.alternateRefs ?? [],
    }
  },

  robotsTxtOptions: {
    policies: [
      {
        userAgent: '*',
        allow: '/',
      },
    ],
    additionalSitemaps: ['https://decorebator.com/sitemap.xml'],
  },
}

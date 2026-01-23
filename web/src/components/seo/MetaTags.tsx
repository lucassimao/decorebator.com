import React from 'react'
import { useLocale } from 'next-intl'

interface MetaTagsProps {
  title?: string
  description?: string
  image?: string
  url?: string
  type?: 'website' | 'article'
  noindex?: boolean
}

const MetaTags: React.FC<MetaTagsProps> = ({
  title = 'Decorebator - AI-Powered Vocabulary Learning',
  description = 'Master any language with AI-powered vocabulary learning, spaced repetition, and 9 engaging quiz modes. Join learners all over the world mastering new languages effectively.',
  image = 'https://decorebator.com/social-share-image.jpg',
  url = 'https://decorebator.com',
  type = 'website',
  noindex = false,
}) => {
  const locale = useLocale()

  return (
    <>
      {/* Basic Meta Tags */}
      <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=5" />
      <meta name="description" content={description} />
      <meta
        name="keywords"
        content="vocabulary learning, language learning, AI learning, spaced repetition, flashcards, quiz app, language app, multilingual, educational app"
      />
      <meta name="author" content="Decorebator Team" />
      <meta name="language" content={locale} />

      {/* Canonical URL */}
      <link rel="canonical" href={url} />

      {/* Robots */}
      {noindex && <meta name="robots" content="noindex, nofollow" />}

      {/* Open Graph Meta Tags */}
      <meta property="og:type" content={type} />
      <meta property="og:title" content={title} />
      <meta property="og:description" content={description} />
      <meta property="og:image" content={image} />
      <meta property="og:image:width" content="1200" />
      <meta property="og:image:height" content="630" />
      <meta
        property="og:image:alt"
        content="Decorebator - AI-Powered Vocabulary Learning Platform"
      />
      <meta property="og:url" content={url} />
      <meta property="og:site_name" content="Decorebator" />
      <meta property="og:locale" content={locale.replace('-', '_')} />

      {/* Twitter Card Meta Tags */}
      <meta name="twitter:card" content="summary_large_image" />
      <meta name="twitter:site" content="@decorebator" />
      <meta name="twitter:creator" content="@decorebator" />
      <meta name="twitter:title" content={title} />
      <meta name="twitter:description" content={description} />
      <meta name="twitter:image" content={image} />
      <meta
        name="twitter:image:alt"
        content="Decorebator - AI-Powered Vocabulary Learning Platform"
      />

      {/* Apple Meta Tags */}
      <meta name="apple-mobile-web-app-capable" content="yes" />
      <meta name="apple-mobile-web-app-status-bar-style" content="default" />
      <meta name="apple-mobile-web-app-title" content="Decorebator" />

      {/* Microsoft Meta Tags */}
      <meta name="msapplication-TileColor" content="#FF7B54" />
      <meta name="msapplication-config" content="/browserconfig.xml" />

      {/* Theme Colors */}
      <meta name="theme-color" content="#FF7B54" />
      <meta name="color-scheme" content="light" />
    </>
  )
}

export default MetaTags

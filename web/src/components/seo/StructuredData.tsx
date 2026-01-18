import React from 'react'
import { statsConfig } from '@/config/statsConfig'

interface StructuredDataProps {
  type?: 'website' | 'softwareApplication' | 'mobileApplication'
  locale?: string
  siteUrl?: string
  baseUrl?: string
  description?: string
  faqEntries?: { question: string; answer: string }[]
}

const StructuredData: React.FC<StructuredDataProps> = ({
  type = 'website',
  locale = 'en',
  siteUrl = 'https://decorebator.com',
  baseUrl = 'https://decorebator.com',
  description = 'AI-powered vocabulary learning platform with spaced repetition, interactive quizzes, and comprehensive analytics',
  faqEntries = [],
}) => {
  const generateStructuredData = () => {
    const baseData = {
      '@context': 'https://schema.org',
      '@type': type === 'website' ? 'WebSite' : 'SoftwareApplication',
      name: 'Decorebator',
      description,
      url: baseUrl,
      inLanguage: locale,
      creator: {
        '@type': 'Organization',
        name: 'Decorebator Team',
      },
      publisher: {
        '@type': 'Organization',
        name: 'Decorebator',
      },
    }

    if (type === 'softwareApplication' || type === 'mobileApplication') {
      return {
        ...baseData,
        '@type': 'SoftwareApplication',
        applicationCategory: 'EducationalApplication',
        applicationSubCategory: 'Language Learning',
        operatingSystem: 'iOS, Android',
        offers: [
          {
            '@type': 'Offer',
            price: '0',
            priceCurrency: 'USD',
            name: 'Free Plan',
            description: '1 wordlist, up to 10 words, basic quiz modes',
          },
          {
            '@type': 'Offer',
            price: '6.99',
            priceCurrency: 'USD',
            name: 'Monthly Premium',
            description:
              'Unlimited wordlists and words, all 8 quiz modes, advanced analytics, offline support',
          },
          {
            '@type': 'Offer',
            price: '69.90',
            priceCurrency: 'USD',
            name: 'Annual Premium',
            description: 'Everything in Monthly Premium with annual savings',
          },
        ],
        ...(statsConfig.showStats.starRating && {
          aggregateRating: {
            '@type': 'AggregateRating',
            ratingValue: statsConfig.values.starRating.split('/')[0],
            ratingCount: statsConfig.showStats.userCount
              ? statsConfig.values.userCount.replace('+', '').replace(',', '')
              : '1000',
            bestRating: '5',
            worstRating: '1',
          },
        }),
        featureList: [
          'AI-powered content generation',
          '7-box Leitner spaced repetition system',
          '8 interactive quiz modes',
          ...(statsConfig.showStats.languageCount
            ? [`Multi-language support (${statsConfig.values.languageCount} languages)`]
            : []),
          'Visual learning with AI-generated images',
          'Audio pronunciation with TTS',
          'Comprehensive analytics and progress tracking',
          'Offline support for premium users',
          'Error reporting and content quality control',
          'Interactive flashcards with flip animations',
        ],
        screenshot: 'https://decorebator.com/app-screenshot.jpeg',
        softwareVersion: '2.0',
        datePublished: '2024-01-01',
        dateModified: '2026-01-01',
        inLanguage: ['en', 'es', 'fr', 'de', 'it', 'pt', 'ja'],
        installUrl: {
          iOS: 'https://apps.apple.com/app/decorebator',
          Android: 'https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator',
        },
      }
    }

    if (type === 'website') {
      return {
        ...baseData,
        mainEntity: {
          '@type': 'SoftwareApplication',
          name: 'Decorebator',
          applicationCategory: 'EducationalApplication',
        },
      }
    }

    return baseData
  }

  const faqStructuredData =
    faqEntries.length > 0
      ? {
          '@context': 'https://schema.org',
          '@type': 'FAQPage',
          mainEntity: faqEntries.map((entry) => ({
            '@type': 'Question',
            name: entry.question,
            acceptedAnswer: {
              '@type': 'Answer',
              text: entry.answer,
            },
          })),
        }
      : null

  const organizationStructuredData = {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: 'Decorebator',
    url: baseUrl,
    logo: 'https://decorebator.com/icon-512x512.png',
    description,
    foundingDate: '2024',
    contactPoint: {
      '@type': 'ContactPoint',
      contactType: 'customer service',
      email: 'support@decorebator.com',
    },
    sameAs: [
      'https://twitter.com/decorebator',
      'https://facebook.com/decorebator',
      'https://instagram.com/decorebator',
    ],
  }

  const breadcrumbStructuredData = {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: [
      {
        '@type': 'ListItem',
        position: 1,
        name: 'Home',
        item: siteUrl,
      },
      {
        '@type': 'ListItem',
        position: 2,
        name: 'Language Learning App',
        item: siteUrl,
      },
    ],
  }

  const productStructuredData = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: 'Decorebator - AI Vocabulary Learning App',
    description,
    brand: {
      '@type': 'Brand',
      name: 'Decorebator',
    },
    category: 'Educational Software',
    offers: [
      {
        '@type': 'Offer',
        name: 'Free Plan',
        price: '0',
        priceCurrency: 'USD',
        availability: 'https://schema.org/InStock',
        description: '1 wordlist, up to 10 words, basic quiz modes',
      },
      {
        '@type': 'Offer',
        name: 'Monthly Premium',
        price: '6.99',
        priceCurrency: 'USD',
        priceSpecification: {
          '@type': 'RecurringPaymentFrequency',
          frequency: 'Monthly',
        },
        availability: 'https://schema.org/InStock',
        description: 'Unlimited wordlists and words, all quiz modes, offline support',
      },
      {
        '@type': 'Offer',
        name: 'Annual Premium',
        price: '69.90',
        priceCurrency: 'USD',
        priceSpecification: {
          '@type': 'RecurringPaymentFrequency',
          frequency: 'Yearly',
        },
        availability: 'https://schema.org/InStock',
        description: 'Best value - includes everything in Monthly plus early access features',
      },
    ],
    hasFeatureList: [
      'AI-powered content generation',
      '7-box Leitner spaced repetition',
      '8 interactive quiz modes',
      'Multi-language support (7 languages)',
      'Offline learning capability',
      'Advanced progress analytics',
      'Visual learning with AI images',
      'Audio pronunciation training',
    ],
    operatingSystem: ['iOS', 'Android'],
    applicationCategory: 'EducationalApplication',
    isAccessibleForFree: true,
  }

  const videoObjectData = {
    '@context': 'https://schema.org',
    '@type': 'VideoObject',
    name: 'Decorebator App Demo - AI Vocabulary Learning',
    description,
    thumbnailUrl: 'https://decorebator.com/app-screenshot.jpeg',
    uploadDate: '2024-12-01',
    contentUrl: 'https://decorebator.com/hero-demo.mp4',
    publisher: {
      '@type': 'Organization',
      name: 'Decorebator',
    },
  }

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(generateStructuredData()),
        }}
      />
      {faqStructuredData && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(faqStructuredData),
          }}
        />
      )}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(organizationStructuredData),
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(breadcrumbStructuredData),
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(productStructuredData),
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(videoObjectData),
        }}
      />
    </>
  )
}

export default StructuredData

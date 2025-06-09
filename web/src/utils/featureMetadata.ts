import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';

export interface FeatureMetadataConfig {
  featureKey: string;
  locale: string;
  keywords: string;
  ogImage?: string;
}

// Map kebab-case URLs to camelCase JSON keys
const featureKeyMap: Record<string, string> = {
  'ai-content': 'aiContent',
  'spaced-repetition': 'spacedRepetition',
  'quiz-modes': 'quizModes',
  'visual-learning': 'visualLearning',
  'audio-learning': 'audioLearning',
  'analytics': 'analytics',
  'multi-language': 'multiLanguage',
  'flashcards': 'flashcards',
  'error-reporting': 'errorReporting',
  'offline-support': 'offlineSupport'
};

export async function generateFeatureMetadata({
  featureKey,
  locale,
  keywords,
  ogImage = 'https://decorebator.com/og-features.jpg'
}: FeatureMetadataConfig): Promise<Metadata> {
  const translationKey = featureKeyMap[featureKey] || featureKey;
  const t = await getTranslations({ locale, namespace: `featurePages.${translationKey}` });
  
  return {
    title: `${t('title')} | Decorebator`,
    description: t('description'),
    keywords: `${keywords}, vocabulary learning, language learning, Decorebator, AI-powered education`,
    authors: [{ name: 'Decorebator Team' }],
    creator: 'Decorebator',
    publisher: 'Decorebator',
    robots: 'index, follow',
    alternates: {
      canonical: `https://decorebator.com/${locale}/features/${featureKey}`,
      languages: {
        'en': `https://decorebator.com/en/features/${featureKey}`,
        'es': `https://decorebator.com/es/features/${featureKey}`,
        'fr': `https://decorebator.com/fr/features/${featureKey}`,
        'de': `https://decorebator.com/de/features/${featureKey}`,
        'it': `https://decorebator.com/it/features/${featureKey}`,
        'pt': `https://decorebator.com/pt/features/${featureKey}`,
        'ja': `https://decorebator.com/ja/features/${featureKey}`,
      },
    },
    openGraph: {
      title: t('title'),
      description: t('description'),
      type: 'website',
      url: `https://decorebator.com/${locale}/features/${featureKey}`,
      siteName: 'Decorebator',
      locale: locale,
      images: [
        {
          url: ogImage,
          width: 1200,
          height: 630,
          alt: t('title'),
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: t('title'),
      description: t('description'),
      images: [ogImage],
    },
  };
}

export function generateFeatureStructuredData(
  featureKey: string,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  t: any
) {
  return {
    "@context": "https://schema.org",
    "@type": "WebPage",
    "name": t('title'),
    "description": t('description'),
    "url": `https://decorebator.com/features/${featureKey}`,
    "isPartOf": {
      "@type": "WebSite",
      "name": "Decorebator",
      "url": "https://decorebator.com"
    },
    "about": {
      "@type": "SoftwareApplication",
      "name": "Decorebator",
      "applicationCategory": "EducationalApplication",
      "operatingSystem": "Web, iOS, Android",
      "description": "AI-powered vocabulary learning platform with spaced repetition and 8 quiz modes",
      "featureList": [
        "AI Content Generation",
        "Spaced Repetition Learning",
        "8 Interactive Quiz Modes",
        "Visual Learning with AI Images",
        "Multi-Language Audio",
        "Advanced Analytics",
        "Offline Support",
        "Error Reporting System"
      ]
    },
    "mainEntity": {
      "@type": "SoftwareFeature",
      "name": t('title'),
      "description": t('description'),
    }
  };
}
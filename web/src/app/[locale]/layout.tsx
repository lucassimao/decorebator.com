import { AppStoreModalProvider } from '@/components/common/AppStoreModalProvider'
import ScrollToTopButton from '@/components/common/ScrollToTopButton'
import StructuredData from '@/components/seo/StructuredData'
import type { Metadata } from 'next'
import { hasLocale, NextIntlClientProvider } from 'next-intl'
import { notFound } from 'next/navigation'
import { routing } from '../../../i18n'

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params

  const localeDescriptions: Record<string, string> = {
    en: 'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.',
    es: 'Domina cualquier idioma con aprendizaje de vocabulario con IA, repetición espaciada y 8 modos de quiz atractivos. Únete a estudiantes de todo el mundo que dominan nuevos idiomas eficazmente.',
    fr: "Maîtrisez n'importe quelle langue avec l'apprentissage de vocabulaire alimenté par l'IA, la répétition espacée et 8 modes de quiz engageants. Rejoignez des apprenants du monde entier qui maîtrisent efficacement de nouvelles langues.",
    de: 'Meistern Sie jede Sprache mit KI-gestütztem Vokabellernen, Wiederholung mit Abstand und 8 fesselnden Quiz-Modi. Schließen Sie sich Lernenden aus aller Welt an, die effektiv neue Sprachen meistern.',
    it: "Padroneggia qualsiasi lingua con l'apprendimento del vocabolario alimentato dall'IA, la ripetizione distanziata e 8 modalità quiz coinvolgenti. Unisciti a studenti di tutto il mondo che padroneggiano efficacemente nuove lingue.",
    pt: 'Domine qualquer idioma com aprendizado de vocabulário com IA, repetição espaçada e 8 modos de quiz envolventes. Junte-se a estudantes de todo o mundo dominando novos idiomas de forma eficaz.',
    ja: 'AIが定義・例文・画像・音声を自動生成。実証済みの間隔反復と8種類のクイズで、語彙がしっかり定着します。',
  }

  const localeNames: Record<string, string> = {
    en: 'English',
    es: 'Español',
    fr: 'Français',
    de: 'Deutsch',
    it: 'Italiano',
    pt: 'Português',
    ja: '日本語',
  }

  return {
    title: `Decorebator - AI-Powered Vocabulary Learning${locale !== 'en' ? ` | ${localeNames[locale]}` : ''}`,
    description: localeDescriptions[locale] || localeDescriptions.en,
    keywords: [
      'vocabulary learning',
      'spaced repetition',
      'language learning app',
      'AI education',
      'Leitner system',
      'multilingual learning',
      'flashcards',
      'pronunciation training',
      'educational technology',
      'interactive learning',
      'mobile learning',
      'offline learning',
      'progress tracking',
      'quiz modes',
      'visual learning',
      'audio learning',
      'memory retention',
      'language acquisition',
      'study app',
      'educational app',
    ],
    authors: [{ name: 'Decorebator Team' }],
    category: 'Education',
    classification: 'Educational Technology > Language Learning > Vocabulary',
    openGraph: {
      type: 'website',
      locale: locale.replace('-', '_'),
      url: `https://decorebator.com/${locale}`,
      siteName: 'Decorebator',
      title: `Decorebator - AI-Powered Vocabulary Learning${locale !== 'en' ? ` | ${localeNames[locale]}` : ''}`,
      description: localeDescriptions[locale] || localeDescriptions.en,
      images: [
        {
          url: 'https://decorebator.com/social-share-image.jpg',
          width: 1200,
          height: 630,
          alt: 'Decorebator - AI-Powered Vocabulary Learning Platform',
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      site: '@decorebator',
      creator: '@decorebator',
      title: `Decorebator - AI-Powered Vocabulary Learning${locale !== 'en' ? ` | ${localeNames[locale]}` : ''}`,
      description: localeDescriptions[locale] || localeDescriptions.en,
      images: ['https://decorebator.com/social-share-image.jpg'],
    },
    alternates: {
      canonical: `https://decorebator.com/${locale}`,
      languages: {
        en: 'https://decorebator.com/en',
        es: 'https://decorebator.com/es',
        fr: 'https://decorebator.com/fr',
        de: 'https://decorebator.com/de',
        it: 'https://decorebator.com/it',
        pt: 'https://decorebator.com/pt',
        ja: 'https://decorebator.com/ja',
      },
    },
    other: {
      'application-category': 'EducationalApplication',
      topic: 'language-learning,vocabulary,education,ai,spaced-repetition',
      'content-type': 'educational-platform',
      'platform-type': 'mobile-app',
      'learning-methodology': 'spaced-repetition,active-recall,multi-modal',
      'ai-features': 'content-generation,image-generation,audio-generation',
      'supported-languages': 'english,spanish,french,german,italian,portuguese,japanese',
      accessibility: 'screen-reader-compatible,keyboard-navigation',
      'pricing-model': 'freemium',
    },
  }
}

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }))
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params

  // Ensure that the incoming `locale` is valid
  if (!hasLocale(routing.locales, locale)) {
    notFound()
  }

  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider>
          <AppStoreModalProvider>
            <StructuredData type="website" />
            {children}
            <ScrollToTopButton />
          </AppStoreModalProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  )
}

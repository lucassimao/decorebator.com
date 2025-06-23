import type { Metadata } from "next";
import {NextIntlClientProvider} from 'next-intl';
import {routing} from '../../../i18n';
import {notFound} from 'next/navigation';
import { AppStoreModalProvider } from '@/components/common/AppStoreModalProvider';
import ScrollToTopButton from '@/components/common/ScrollToTopButton';
import StructuredData from '@/components/seo/StructuredData';

export async function generateMetadata({
  params
}: {
  params: Promise<{locale: string}>;
}): Promise<Metadata> {
  const {locale} = await params;

  const localeDescriptions: Record<string, string> = {
    en: "Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.",
    es: "Domina cualquier idioma con aprendizaje de vocabulario con IA, repetición espaciada y 8 modos de quiz atractivos. Únete a estudiantes de todo el mundo que dominan nuevos idiomas eficazmente.",
    fr: "Maîtrisez n'importe quelle langue avec l'apprentissage de vocabulaire alimenté par l'IA, la répétition espacée et 8 modes de quiz engageants. Rejoignez des apprenants du monde entier qui maîtrisent efficacement de nouvelles langues.",
    de: "Meistern Sie jede Sprache mit KI-gestütztem Vokabellernen, Wiederholung mit Abstand und 8 fesselnden Quiz-Modi. Schließen Sie sich Lernenden aus aller Welt an, die effektiv neue Sprachen meistern.",
    it: "Padroneggia qualsiasi lingua con l'apprendimento del vocabolario alimentato dall'IA, la ripetizione distanziata e 8 modalità quiz coinvolgenti. Unisciti a studenti di tutto il mondo che padroneggiano efficacemente nuove lingue.",
    pt: "Domine qualquer idioma com aprendizado de vocabulário com IA, repetição espaçada e 8 modos de quiz envolventes. Junte-se a estudantes de todo o mundo dominando novos idiomas de forma eficaz.",
    ja: "AI搭載語彙学習、科学的に証明された間隔反復、8つの魅力的なクイズモードで任意の言語をマスターしましょう。世界中の学習者と一緒に効果的に新しい言語を習得しましょう。"
  };

  const localeNames: Record<string, string> = {
    en: "English",
    es: "Español", 
    fr: "Français",
    de: "Deutsch",
    it: "Italiano",
    pt: "Português",
    ja: "日本語"
  };

  return {
    title: `Decorebator - AI-Powered Vocabulary Learning${locale !== 'en' ? ` | ${localeNames[locale]}` : ''}`,
    description: localeDescriptions[locale] || localeDescriptions.en,
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
        'en': 'https://decorebator.com/en',
        'es': 'https://decorebator.com/es',
        'fr': 'https://decorebator.com/fr',
        'de': 'https://decorebator.com/de',
        'it': 'https://decorebator.com/it',
        'pt': 'https://decorebator.com/pt',
        'ja': 'https://decorebator.com/ja',
      },
    },
  };
}

export function generateStaticParams() {
  return routing.locales.map((locale) => ({locale}));
}

export default async function LocaleLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: Promise<{locale: string}>;
}) {
  const {locale} = await params;
  
  // Validate locale
  if (!routing.locales.includes(locale as typeof routing.locales[number])) {
    notFound();
  }

  // Load messages directly
  const messages = (await import(`../../../messages/${locale}.json`)).default;

  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider messages={messages} locale={locale}>
          <AppStoreModalProvider>
            <StructuredData type="website" />
            {children}
            <ScrollToTopButton />
          </AppStoreModalProvider>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
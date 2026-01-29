import type { Metadata } from 'next'

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params

  const titles: Record<string, string> = {
    en: 'Terms of Service',
    es: 'Términos de Servicio',
    fr: "Conditions d'Utilisation",
    de: 'Nutzungsbedingungen',
    it: 'Termini di Servizio',
    pt: 'Termos de Serviço',
    ja: '利用規約',
  }

  const descriptions: Record<string, string> = {
    en: 'Read the terms and conditions for using Decorebator, the AI-powered vocabulary learning app.',
    es: 'Lee los términos y condiciones para usar Decorebator, la aplicación de aprendizaje de vocabulario con IA.',
    fr: "Lisez les termes et conditions d'utilisation de Decorebator, l'application d'apprentissage de vocabulaire alimentée par l'IA.",
    de: 'Lesen Sie die Nutzungsbedingungen für Decorebator, die KI-gestützte Vokabellern-App.',
    it: "Leggi i termini e le condizioni per l'utilizzo di Decorebator, l'app per l'apprendimento del vocabolario basata sull'IA.",
    pt: 'Leia os termos e condições de uso do Decorebator, o aplicativo de aprendizado de vocabulário com IA.',
    ja: 'AI搭載の語彙学習アプリ、Decorebatorの利用規約をお読みください。',
  }

  const title = `${titles[locale] || titles.en} | Decorebator`
  const description = descriptions[locale] || descriptions.en

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      url: `https://decorebator.com/${locale}/terms`,
      siteName: 'Decorebator',
      images: [
        {
          url: `https://decorebator.com/og?locale=${locale}&page=terms`,
          width: 1200,
          height: 630,
          alt: 'Decorebator Terms of Service',
        },
      ],
      type: 'website',
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [`https://decorebator.com/og?locale=${locale}&page=terms`],
    },
    alternates: {
      canonical: `https://decorebator.com/${locale}/terms`,
    },
  }
}

export default function TermsLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

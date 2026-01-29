import type { Metadata } from 'next'

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params

  const titles: Record<string, string> = {
    en: 'Privacy Policy',
    es: 'Política de Privacidad',
    fr: 'Politique de Confidentialité',
    de: 'Datenschutzrichtlinie',
    it: 'Informativa sulla Privacy',
    pt: 'Política de Privacidade',
    ja: 'プライバシーポリシー',
  }

  const descriptions: Record<string, string> = {
    en: 'Learn how Decorebator collects, uses, and protects your personal information. Your privacy is important to us.',
    es: 'Descubre cómo Decorebator recopila, usa y protege tu información personal. Tu privacidad es importante para nosotros.',
    fr: 'Découvrez comment Decorebator collecte, utilise et protège vos informations personnelles. Votre vie privée est importante pour nous.',
    de: 'Erfahren Sie, wie Decorebator Ihre persönlichen Daten sammelt, verwendet und schützt. Ihre Privatsphäre ist uns wichtig.',
    it: 'Scopri come Decorebator raccoglie, utilizza e protegge le tue informazioni personali. La tua privacy è importante per noi.',
    pt: 'Saiba como o Decorebator coleta, usa e protege suas informações pessoais. Sua privacidade é importante para nós.',
    ja: 'Decorebatorがお客様の個人情報をどのように収集、使用、保護するかをご覧ください。お客様のプライバシーは私たちにとって重要です。',
  }

  const title = `${titles[locale] || titles.en} | Decorebator`
  const description = descriptions[locale] || descriptions.en

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      url: `https://decorebator.com/${locale}/privacy`,
      siteName: 'Decorebator',
      images: [
        {
          url: `https://decorebator.com/og?locale=${locale}&page=privacy`,
          width: 1200,
          height: 630,
          alt: 'Decorebator Privacy Policy',
        },
      ],
      type: 'website',
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [`https://decorebator.com/og?locale=${locale}&page=privacy`],
    },
    alternates: {
      canonical: `https://decorebator.com/${locale}/privacy`,
    },
  }
}

export default function PrivacyLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

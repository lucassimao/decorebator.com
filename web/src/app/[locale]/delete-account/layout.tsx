import { routing } from '@/i18n/routing'
import type { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'

const baseUrl = 'https://decorebator.com'

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'accountDeletion' })
  const canonical = `${baseUrl}/${locale}/delete-account`
  const languages = Object.fromEntries(
    routing.locales.map((supportedLocale) => [
      supportedLocale,
      `${baseUrl}/${supportedLocale}/delete-account`,
    ])
  )

  return {
    title: `${t('metadataTitle')} | Decorebator`,
    description: t('metadataDescription'),
    alternates: {
      canonical,
      languages: {
        ...languages,
        'x-default': `${baseUrl}/en/delete-account`,
      },
    },
    openGraph: {
      title: `${t('metadataTitle')} | Decorebator`,
      description: t('metadataDescription'),
      url: canonical,
      siteName: 'Decorebator',
      type: 'website',
    },
  }
}

export default function DeleteAccountLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}

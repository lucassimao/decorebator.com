import type { Metadata } from "next";
import {NextIntlClientProvider} from 'next-intl';
import {routing} from '../../../i18n';
import {notFound} from 'next/navigation';

export const metadata: Metadata = {
  title: "Decorebator - AI-Powered Vocabulary Learning",
  description: "Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Track your progress with advanced analytics. Join 10,000+ learners today!",
};

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
    <NextIntlClientProvider messages={messages} locale={locale}>
      {children}
    </NextIntlClientProvider>
  );
}
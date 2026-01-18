import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import './globals.css'
import Script from 'next/script'
import { SpeedInsights } from '@vercel/speed-insights/next'
import { Analytics } from '@vercel/analytics/next'
import { appStoreConfig } from '@/config/appStoreConfig'
const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
})

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
})

export const metadata: Metadata = {
  title: 'Decorebator - AI-Powered Vocabulary Learning',
  description:
    'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.',
  keywords:
    'vocabulary learning, language learning, AI learning, spaced repetition, flashcards, quiz app, language app, multilingual, educational app',
  authors: [{ name: 'Decorebator Team' }],
  creator: 'Decorebator Team',
  publisher: 'Decorebator',
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://decorebator.com',
    siteName: 'Decorebator',
    title: 'Decorebator - AI-Powered Vocabulary Learning',
    description:
      'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.',
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
    title: 'Decorebator - AI-Powered Vocabulary Learning',
    description:
      'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes.',
    images: ['https://decorebator.com/social-share-image.jpg'],
  },
  manifest: '/manifest.json',
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
      { url: '/favicon-96x96.png', sizes: '96x96', type: 'image/png' },
    ],
    apple: [{ url: '/apple-touch-icon.png', sizes: '180x180', type: 'image/png' }],
    other: [
      { rel: 'icon', url: '/icon-192x192.png', sizes: '192x192', type: 'image/png' },
      { rel: 'icon', url: '/icon-512x512.png', sizes: '512x512', type: 'image/png' },
    ],
  },
  metadataBase: new URL('https://decorebator.com'),
  alternates: {
    canonical: 'https://decorebator.com',
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
  appLinks: {
    ios: {
      url: appStoreConfig.appStoreUrl ?? 'https://decorebator.com',
      app_store_id: '6749329064',
      app_name: 'Decorebator',
    },
    android: {
      package: 'com.lsimaocosta.decorebator',
      url: appStoreConfig.playStoreUrl ?? 'https://decorebator.com',
      app_name: 'Decorebator',
    },
    web: {
      url: 'https://decorebator.com',
      should_fallback: true,
    },
  },
  verification: {
    google: process.env.GOOGLE_SITE_VERIFICATION,
    yandex: process.env.YANDEX_SITE_VERIFICATION,
    yahoo: process.env.YAHOO_SITE_VERIFICATION,
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* Preload LCP hero image to improve discovery/prioritization */}
        <link rel="preload" as="image" href="/app-screenshot.jpeg" fetchPriority="high" />
        <link rel="preconnect" href="https://static.ads-twitter.com" crossOrigin="" />
        <link rel="preconnect" href="https://upload.wikimedia.org" crossOrigin="" />
        {/* Twitter conversion tracking base code */}
        <Script id="twitter-conversion-tracking" strategy="lazyOnload">
          {`!function(e,t,n,s,u,a){e.twq||(s=e.twq=function(){s.exe?s.exe.apply(s,arguments):s.queue.push(arguments);},s.version='1.1',s.queue=[],u=t.createElement(n),u.async=!0,u.src='https://static.ads-twitter.com/uwt.js',a=t.getElementsByTagName(n)[0],a.parentNode.insertBefore(u,a))}(window,document,'script');twq('config','qpm75');`}
        </Script>
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  )
}

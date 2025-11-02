import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import './globals.css'
import { SpeedInsights } from '@vercel/speed-insights/next'
import { Analytics } from '@vercel/analytics/next'
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
  verification: {
    google: process.env.GOOGLE_SITE_VERIFICATION,
    yandex: process.env.YANDEX_SITE_VERIFICATION,
    yahoo: process.env.YAHOO_SITE_VERIFICATION,
  },
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <head>
        {/* Preload LCP hero image to improve discovery/prioritization */}
        <link rel="preload" as="image" href="/app-screenshot.jpeg" fetchPriority="high" />

        {/* Make Font Awesome non-blocking: preload + media=print swap */}
        <link
          rel="preload"
          as="style"
          href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css"
        />
        <link
          id="fa-css"
          rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css"
          media="print"
        />
        <script
          // Swap Font Awesome to media=all after it loads (non-blocking)
          dangerouslySetInnerHTML={{
            __html:
              "(function(){var l=document.getElementById('fa-css');if(!l)return;function s(){l.media='all';} if(l.addEventListener){l.addEventListener('load',s);} else {l.onload=s;}})();",
          }}
        />
        <noscript>
          <link
            rel="stylesheet"
            href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css"
          />
        </noscript>

        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link rel="dns-prefetch" href="https://cdnjs.cloudflare.com" />
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  )
}

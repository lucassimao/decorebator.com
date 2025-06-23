import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Analytics } from "@vercel/analytics/next"
const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Decorebator - AI-Powered Vocabulary Learning",
  description: "Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.",
  keywords: "vocabulary learning, language learning, AI learning, spaced repetition, flashcards, quiz app, language app, multilingual, educational app",
  authors: [{ name: "Decorebator Team" }],
  creator: "Decorebator Team",
  publisher: "Decorebator",
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
    description: 'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes. Join learners all over the world mastering new languages effectively.',
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
    description: 'Master any language with AI-powered vocabulary learning, spaced repetition, and 8 engaging quiz modes.',
    images: ['https://decorebator.com/social-share-image.jpg'],
  },
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
      { url: '/favicon-96x96.png', sizes: '96x96', type: 'image/png' },
    ],
    shortcut: '/favicon.ico',
    apple: '/apple-touch-icon.png',
  },
  manifest: '/manifest.json',
  metadataBase: new URL('https://decorebator.com'),
  alternates: {
    canonical: 'https://decorebator.com',
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
  verification: {
    google: process.env.GOOGLE_SITE_VERIFICATION,
    yandex: process.env.YANDEX_SITE_VERIFICATION,
    yahoo: process.env.YAHOO_SITE_VERIFICATION,
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <link 
          rel="stylesheet" 
          href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" 
        />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link rel="dns-prefetch" href="https://cdnjs.cloudflare.com" />
      </head>
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        {children}
        <Analytics/>
      </body>
    </html>
  );
}

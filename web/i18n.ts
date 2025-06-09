import {defineRouting} from 'next-intl/routing';

export const routing = defineRouting({
  // A list of all locales that are supported
  locales: ['en', 'es', 'fr', 'de', 'it', 'pt', 'ja'],

  // Used when no locale matches
  defaultLocale: 'en',

  // The prefix for all pages
  pathnames: {
    '/': '/',
    '/signup': '/signup',
    '/help': '/help',
    '/privacy': '/privacy',
    '/terms': '/terms',
    '/reset-password': '/reset-password'
  }
});

export type Locale = (typeof routing.locales)[number];
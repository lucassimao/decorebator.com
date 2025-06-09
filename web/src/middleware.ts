import createMiddleware from 'next-intl/middleware';
import {routing} from '../i18n';

export default createMiddleware(routing);

export const config = {
  // Match only internationalized pathnames
  matcher: ['/', '/(de|en|es|fr|it|pt|ja)/:path*', '/((?!api|_next|_vercel|.*\\..*).*)']
};
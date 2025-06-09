import {getRequestConfig} from 'next-intl/server';
import {routing} from '../../i18n';

export default getRequestConfig(async ({locale}) => {
  // Validate that the incoming `locale` parameter is valid
  if (!routing.locales.includes(locale as typeof routing.locales[number])) {
    locale = routing.defaultLocale;
  }

  try {
    const messages = (await import(`../../messages/${locale}.json`)).default;
    return {
      locale,
      messages
    };
  } catch (error) {
    console.error('Failed to load messages for locale:', locale, error);
    // Fallback to English
    const fallbackMessages = (await import(`../../messages/en.json`)).default;
    return {
      locale: 'en',
      messages: fallbackMessages
    };
  }
});
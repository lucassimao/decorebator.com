'use client';

import React from 'react';
import { useTranslations, useLocale } from 'next-intl';
import { FacebookIcon, TwitterIcon, InstagramIcon, LinkedInIcon } from './icons';
import Link from 'next/link';

const FooterSection: React.FC = () => {
  const t = useTranslations('footer');
  const tNav = useTranslations('navigation');
  const locale = useLocale();
  return (
    <footer className="bg-slate-800 text-slate-400 py-12" aria-labelledby="footer-heading">
      <h2 id="footer-heading" className="sr-only">{t('heading')}</h2>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">Decorebator</h5>
            <ul className="space-y-2">
              <li><a href={`/${locale}/#features`} className="hover:text-slate-100">{tNav('features')}</a></li>
              <li><a href={`/${locale}/#how-it-works`} className="hover:text-slate-100">{tNav('howItWorks')}</a></li>
              <li><a href={`/${locale}/#pricing`} className="hover:text-slate-100">{tNav('pricing')}</a></li>
              <li><a href={`/${locale}/#faq`} className="hover:text-slate-100">{tNav('faq')}</a></li>
            </ul>
          </div>
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">{t('support')}</h5>
            <ul className="space-y-2">
              <li><Link href={`/${locale}/help`} className="hover:text-slate-100">{tNav('helpCenter')}</Link></li>
              <li><Link href={`/${locale}/privacy`} className="hover:text-slate-100">{tNav('privacy')}</Link></li>
              <li><Link href={`/${locale}/terms`} className="hover:text-slate-100">{tNav('terms')}</Link></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-slate-700 pt-8 flex flex-col sm:flex-row justify-between items-center">
          <p className="text-sm">{t('copyright', { year: new Date().getFullYear() })}</p>
          <div className="flex space-x-4 mt-4 sm:mt-0">
            <a href="#" className="hover:text-slate-100" aria-label={t('social.facebook')}><FacebookIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label={t('social.twitter')}><TwitterIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label={t('social.instagram')}><InstagramIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label={t('social.linkedin')}><LinkedInIcon className="h-6 w-6" /></a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default FooterSection;
'use client'

import React from 'react'
import { useTranslations, useLocale } from 'next-intl'
import { FacebookIcon /*, TwitterIcon, InstagramIcon */ } from './icons'
import Link from 'next/link'

const FooterSection: React.FC = () => {
  const t = useTranslations('footer')
  const tNav = useTranslations('navigation')
  const locale = useLocale()
  return (
    <footer className="bg-slate-800 py-12 text-slate-400" aria-labelledby="footer-heading">
      <h2 id="footer-heading" className="sr-only">
        {t('heading')}
      </h2>
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-8 grid grid-cols-1 gap-8 md:grid-cols-2">
          <div>
            <h5 className="mb-3 font-semibold text-slate-200">Decorebator</h5>
            <ul className="space-y-2">
              <li>
                <a href={`/${locale}/#features`} className="hover:text-slate-100">
                  {tNav('features')}
                </a>
              </li>
              <li>
                <a href={`/${locale}/#how-it-works`} className="hover:text-slate-100">
                  {tNav('howItWorks')}
                </a>
              </li>
              <li>
                <a href={`/${locale}/#pricing`} className="hover:text-slate-100">
                  {tNav('pricing')}
                </a>
              </li>
              <li>
                <a href={`/${locale}/#faq`} className="hover:text-slate-100">
                  {tNav('faq')}
                </a>
              </li>
            </ul>
          </div>
          <div>
            <h5 className="mb-3 font-semibold text-slate-200">{t('support')}</h5>
            <ul className="space-y-2">
              <li>
                <Link href={`/${locale}/help`} className="hover:text-slate-100">
                  {tNav('helpCenter')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}/privacy`} className="hover:text-slate-100">
                  {tNav('privacy')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}/terms`} className="hover:text-slate-100">
                  {tNav('terms')}
                </Link>
              </li>
            </ul>
          </div>
        </div>
        <div className="flex flex-col items-center justify-between border-t border-slate-700 pt-8 sm:flex-row">
          <p className="text-sm">{t('copyright', { year: new Date().getFullYear() })}</p>
          <div className="mt-4 flex space-x-4 sm:mt-0">
            <a
              href="https://www.facebook.com/profile.php?id=61579066558219"
              className="hover:text-slate-100"
              aria-label={t('social.facebook')}
            >
              <FacebookIcon className="h-6 w-6" />
            </a>
            {/*
            <a href="#" className="hover:text-slate-100" aria-label={t('social.twitter')}><TwitterIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label={t('social.instagram')}><InstagramIcon className="h-6 w-6" /></a>
              */}
          </div>
        </div>
      </div>
    </footer>
  )
}

export default FooterSection

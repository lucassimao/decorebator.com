'use client'

import React from 'react'
import { useTranslations, useLocale } from 'next-intl'
import { FacebookIcon, TwitterIcon, InstagramIcon } from './icons'
import Link from 'next/link'
import Image from 'next/image'
import AppStoreButton from '../common/AppStoreButton'

const FooterSection: React.FC = () => {
  const t = useTranslations('footer')
  const tNav = useTranslations('navigation')
  const locale = useLocale()
  return (
    <footer
      className="relative overflow-hidden bg-gradient-to-b from-slate-800 to-slate-900 py-16 text-slate-400"
      aria-labelledby="footer-heading"
    >
      {/* Decorative gradient blob */}
      <div className="bg-primary-500/5 pointer-events-none absolute -top-40 left-1/2 h-80 w-80 -translate-x-1/2 rounded-full blur-3xl" />

      <h2 id="footer-heading" className="sr-only">
        {t('heading')}
      </h2>
      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Main footer grid */}
        <div className="mb-12 grid grid-cols-1 gap-10 md:grid-cols-4">
          {/* Brand column */}
          <div className="md:col-span-2">
            <Link href={`/${locale}`} className="group mb-4 inline-flex items-center gap-2">
              <Image
                src="/logo.png"
                alt=""
                width={36}
                height={36}
                className="transition-transform duration-200 group-hover:rotate-6"
              />
              <span className="text-xl font-bold text-white">Decorebator</span>
            </Link>
            <p className="mb-6 max-w-sm text-sm leading-relaxed text-slate-400">
              Master vocabulary faster with AI-powered flashcards, spaced repetition, and 9 engaging
              quiz modes. Join 10,000+ learners today.
            </p>
            {/* App store buttons */}
            <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
              <AppStoreButton
                store="apple"
                size="small"
                className="transition-opacity hover:opacity-80"
              />
              <AppStoreButton
                store="google"
                size="small"
                className="transition-opacity hover:opacity-80"
              />
            </div>
          </div>

          {/* Navigation column */}
          <div>
            <h3 className="mb-4 text-sm font-semibold tracking-wider text-slate-200 uppercase">
              Product
            </h3>
            <ul className="space-y-3">
              <li>
                <a
                  href={`/${locale}/#features`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('features')}
                </a>
              </li>
              <li>
                <a
                  href={`/${locale}/#how-it-works`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('howItWorks')}
                </a>
              </li>
              <li>
                <a
                  href={`/${locale}/#pricing`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('pricing')}
                </a>
              </li>
              <li>
                <a href={`/${locale}/#faq`} className="text-sm transition-colors hover:text-white">
                  {tNav('faq')}
                </a>
              </li>
              <li>
                <Link
                  href={`/${locale}/blog`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('blog')}
                </Link>
              </li>
              <li>
                <a
                  href={`/${locale}/#download`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('getStartedFree')}
                </a>
              </li>
            </ul>
          </div>

          {/* Support column */}
          <div>
            <h3 className="mb-4 text-sm font-semibold tracking-wider text-slate-200 uppercase">
              {t('support')}
            </h3>
            <ul className="space-y-3">
              <li>
                <Link
                  href={`/${locale}/help`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('helpCenter')}
                </Link>
              </li>
              <li>
                <Link
                  href={`/${locale}/privacy`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('privacy')}
                </Link>
              </li>
              <li>
                <Link
                  href={`/${locale}/terms`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('terms')}
                </Link>
              </li>
              <li>
                <Link
                  href={`/${locale}/delete-account`}
                  className="text-sm transition-colors hover:text-white"
                >
                  {tNav('deleteAccount')}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="flex flex-col items-center justify-between gap-4 border-t border-slate-700/50 pt-8 sm:flex-row">
          <p className="text-sm text-slate-500">
            {t('copyright', { year: new Date().getFullYear() })}
          </p>

          {/* Social links */}
          <div className="flex items-center gap-1">
            <a
              href="https://www.facebook.com/profile.php?id=61579066558219"
              className="rounded-lg p-2 transition-colors hover:bg-slate-700/50 hover:text-white"
              aria-label={t('social.facebook')}
              target="_blank"
              rel="noopener noreferrer"
            >
              <FacebookIcon className="h-5 w-5" />
            </a>
            <a
              href="https://twitter.com/decorebator"
              className="rounded-lg p-2 transition-colors hover:bg-slate-700/50 hover:text-white"
              aria-label={t('social.twitter')}
              target="_blank"
              rel="noopener noreferrer"
            >
              <TwitterIcon className="h-5 w-5" />
            </a>
            <a
              href="https://instagram.com/decorebator"
              className="rounded-lg p-2 transition-colors hover:bg-slate-700/50 hover:text-white"
              aria-label={t('social.instagram')}
              target="_blank"
              rel="noopener noreferrer"
            >
              <InstagramIcon className="h-5 w-5" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}

export default FooterSection

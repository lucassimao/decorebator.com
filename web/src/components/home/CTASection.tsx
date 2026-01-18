'use client'

import React from 'react'
import { ArrowRight } from 'lucide-react'
import { motion } from 'framer-motion'
import AppStoreButton from '../common/AppStoreButton'
import { statsConfig } from '@/config/statsConfig'
import { useTranslations } from 'next-intl'

const CTASection: React.FC = () => {
  const t = useTranslations('cta')

  return (
    <section
      id="download"
      className="from-secondary-600 via-accent-600 to-primary-600 relative scroll-mt-24 overflow-hidden bg-gradient-to-br py-20 sm:scroll-mt-28 sm:py-24 lg:scroll-mt-32 lg:py-32"
    >
      {/* Animated background gradient mesh */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute -top-40 -left-40 h-96 w-96 rounded-full bg-white/10 blur-3xl"></div>
        <div className="absolute -right-40 -bottom-40 h-96 w-96 rounded-full bg-white/10 blur-3xl"></div>
        <div className="absolute top-1/2 left-1/2 h-96 w-96 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white/5 blur-3xl"></div>
      </div>

      <div className="relative z-10 mx-auto max-w-5xl px-4 text-center sm:px-6 lg:px-8">
        {/* Headline */}
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="mb-6 text-4xl leading-[1.05] font-bold text-white sm:text-5xl lg:text-6xl"
        >
          {t.rich('title', {
            accent: (chunks) => (
              <span className="mt-2 block bg-gradient-to-r from-yellow-200 to-yellow-400 bg-clip-text text-transparent">
                {chunks}
              </span>
            ),
          })}
        </motion.h2>

        {/* Subheadline */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mx-auto mb-10 max-w-2xl text-lg leading-relaxed text-white/90 sm:text-xl"
        >
          {t('subtitle', { count: statsConfig.values.userCount })}
        </motion.p>

        {/* Primary CTA Button */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.2 }}
          className="mb-10 flex justify-center"
        >
          <a
            href="#download"
            className="group focus-visible:ring-offset-primary-700 inline-flex items-center gap-2 rounded-full bg-white px-8 py-4 text-lg font-semibold text-slate-900 shadow-xl transition-all duration-200 hover:scale-105 hover:shadow-2xl focus-visible:ring-2 focus-visible:ring-white/80 focus-visible:ring-offset-2 focus-visible:outline-none"
          >
            {t('primaryButton')}
            <ArrowRight
              className="h-5 w-5 transition-transform duration-200 group-hover:translate-x-1"
              aria-hidden="true"
              focusable="false"
            />
          </a>
        </motion.div>

        {/* Trust badges */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
          className="mb-10 flex flex-col items-center justify-center gap-4 text-sm text-white/80 sm:flex-row sm:gap-8"
        >
          <span className="flex items-center gap-2">
            <svg
              className="h-5 w-5"
              fill="currentColor"
              viewBox="0 0 20 20"
              aria-hidden="true"
              focusable="false"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
            {t('trust.noCard')}
          </span>
          <span className="flex items-center gap-2">
            <svg
              className="h-5 w-5"
              fill="currentColor"
              viewBox="0 0 20 20"
              aria-hidden="true"
              focusable="false"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
            {t('trust.trial')}
          </span>
          <span className="flex items-center gap-2">
            <svg
              className="h-5 w-5"
              fill="currentColor"
              viewBox="0 0 20 20"
              aria-hidden="true"
              focusable="false"
            >
              <path
                fillRule="evenodd"
                d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                clipRule="evenodd"
              />
            </svg>
            {t('trust.cancel')}
          </span>
        </motion.div>

        {/* App Store Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.4 }}
          className="flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4"
        >
          <AppStoreButton
            store="apple"
            className="h-14 opacity-95 transition-opacity hover:opacity-100"
          />
          <AppStoreButton
            store="google"
            className="h-14 opacity-95 transition-opacity hover:opacity-100"
          />
        </motion.div>
      </div>
    </section>
  )
}

export default CTASection

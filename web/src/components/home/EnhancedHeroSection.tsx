'use client'

import React from 'react'
import Image from 'next/image'
import { Star, Zap } from 'lucide-react'
import { motion } from 'framer-motion'
import AppStoreButton from '../common/AppStoreButton'
import { statsConfig } from '@/config/statsConfig'
import { useTranslations } from 'next-intl'

type EnhancedHeroSectionProps = Record<string, never>

const EnhancedHeroSection: React.FC<EnhancedHeroSectionProps> = () => {
  const t = useTranslations('hero')

  return (
    <section className="relative overflow-hidden bg-gradient-to-b from-white via-white to-slate-50 pt-20 pb-16 sm:pt-28 sm:pb-20 lg:pt-32 lg:pb-24">
      <div className="pointer-events-none absolute inset-0">
        <div className="bg-primary-200/40 absolute -top-32 left-1/2 h-[28rem] w-[28rem] -translate-x-1/2 rounded-full blur-[120px]" />
        <div className="bg-accent-200/30 absolute right-[-10%] -bottom-40 h-[26rem] w-[26rem] rounded-full blur-[120px]" />
      </div>

      <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Center-aligned content */}
        <div className="mx-auto max-w-4xl text-center">
          {/* Trust badge */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="border-primary-200 bg-primary-50/50 mb-6 inline-flex items-center gap-2 rounded-full border px-4 py-2 backdrop-blur-sm"
          >
            <Zap className="text-primary-500 h-4 w-4" aria-hidden="true" focusable="false" />
            <span className="text-primary-700 text-sm font-semibold">
              {t('trustBadge', { count: statsConfig.values.userCount })}
            </span>
          </motion.div>

          {/* Main headline */}
          <motion.h1
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="mb-6 text-4xl leading-[1.05] font-bold tracking-tight text-slate-900 sm:text-5xl lg:text-6xl"
          >
            {t.rich('title', {
              highlight: (chunks) => (
                <span className="from-primary-500 via-primary-600 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
                  {chunks}
                </span>
              ),
            })}
          </motion.h1>

          {/* Subheadline */}
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="mb-8 text-lg leading-relaxed text-slate-700 sm:text-xl lg:text-2xl"
          >
            {t.rich('subtitle', {
              strong: (chunks) => (
                <strong className="font-semibold text-slate-900">{chunks}</strong>
              ),
              accent: (chunks) => <span className="text-slate-700">{chunks}</span>,
            })}
          </motion.p>

          {/* Trust row */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
            className="mb-12 flex flex-col items-center justify-center gap-6 text-sm text-slate-600 sm:flex-row sm:gap-8"
          >
            {statsConfig.showStats.starRating && (
              <div
                className="flex items-center gap-2"
                role="img"
                aria-label={t('rating.aria', { rating: statsConfig.values.starRating })}
              >
                <div className="flex" aria-hidden="true">
                  {[1, 2, 3, 4, 5].map((star) => (
                    <Star
                      key={star}
                      className="h-5 w-5 fill-yellow-400 text-yellow-400"
                      strokeWidth={0}
                      aria-hidden="true"
                      focusable="false"
                    />
                  ))}
                </div>
                <span className="font-semibold text-slate-700">
                  {statsConfig.values.starRating}/5
                </span>
                <span className="text-slate-500">{t('rating.reviews', { count: '2,500+' })}</span>
              </div>
            )}
            <div className="hidden h-4 w-px bg-slate-300 sm:block" aria-hidden="true"></div>
            <span className="font-medium text-slate-600">{t('trust.freePlan')}</span>
            <span className="font-medium text-slate-600">{t('trust.noCard')}</span>
          </motion.div>

          {/* App Store Buttons */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="flex flex-col items-center justify-center gap-3 sm:flex-row sm:gap-4"
          >
            <AppStoreButton
              store="apple"
              className="h-12 opacity-90 transition-opacity hover:opacity-100"
            />
            <AppStoreButton
              store="google"
              className="h-12 opacity-90 transition-opacity hover:opacity-100"
            />
          </motion.div>
        </div>

        {/* Hero Visual - Bento Grid Style */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.7, delay: 0.3 }}
          className="relative mx-auto mt-16 max-w-5xl lg:mt-20"
        >
          <div className="grid grid-cols-1 gap-4 md:grid-cols-12 md:gap-6">
            {/* Main phone mockup - larger column */}
            <div className="relative md:col-span-7">
              <div className="group relative overflow-hidden rounded-3xl bg-gradient-to-br from-slate-900 to-slate-800 p-3 shadow-2xl transition-all duration-300 hover:shadow-[0_20px_70px_-15px_rgba(0,0,0,0.3)]">
                <div className="max-h-[460px] overflow-hidden rounded-2xl bg-white sm:max-h-[560px] lg:max-h-[620px]">
                  <video
                    className="h-full w-full object-cover"
                    autoPlay
                    loop
                    muted
                    playsInline
                    preload="metadata"
                    poster="/app-screenshot.jpeg"
                    aria-label={t('video.aria')}
                  >
                    <source src="/hero-demo.webm" type="video/webm" />
                    <source src="/hero-demo.mp4" type="video/mp4" />
                    <Image
                      src="/app-screenshot.jpeg"
                      alt={t('imageAlt')}
                      width={600}
                      height={1200}
                      className="h-full w-full object-cover"
                      priority
                      sizes="(max-width: 768px) 100vw, 600px"
                    />
                  </video>
                </div>
                {/* Glow effect */}
                <div className="from-primary-500 to-accent-500 pointer-events-none absolute -inset-0.5 rounded-3xl bg-gradient-to-r opacity-0 blur transition-opacity duration-300 group-hover:opacity-20"></div>
              </div>
            </div>

            {/* Side cards - smaller columns */}
            <div className="flex flex-col gap-4 md:col-span-5 md:gap-6">
              {/* Analytics preview card */}
              <div className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50/50 p-6 shadow-lg transition-all duration-300 hover:shadow-xl">
                <div className="bg-success-100 text-success-700 mb-3 inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-semibold">
                  <span className="relative flex h-2 w-2">
                    <span className="bg-success-400 absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"></span>
                    <span className="bg-success-500 relative inline-flex h-2 w-2 rounded-full"></span>
                  </span>
                  {t('cards.analytics.badge')}
                </div>
                <h3 className="mb-2 text-lg font-bold text-slate-900">
                  {t('cards.analytics.title')}
                </h3>
                <p className="mb-4 text-sm text-slate-600">{t('cards.analytics.description')}</p>
                {/* Mini progress bars preview */}
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="w-16 text-xs text-slate-500">
                      {t('cards.analytics.languages.french')}
                    </span>
                    <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                      <div className="from-success-400 to-success-500 h-full w-[85%] rounded-full bg-gradient-to-r"></div>
                    </div>
                    <span className="text-success-600 w-8 text-xs font-semibold">85%</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-16 text-xs text-slate-500">
                      {t('cards.analytics.languages.spanish')}
                    </span>
                    <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                      <div className="from-primary-400 to-primary-500 h-full w-[72%] rounded-full bg-gradient-to-r"></div>
                    </div>
                    <span className="text-primary-600 w-8 text-xs font-semibold">72%</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-16 text-xs text-slate-500">
                      {t('cards.analytics.languages.japanese')}
                    </span>
                    <div className="h-2 flex-1 overflow-hidden rounded-full bg-slate-100">
                      <div className="from-accent-400 to-accent-500 h-full w-[45%] rounded-full bg-gradient-to-r"></div>
                    </div>
                    <span className="text-accent-600 w-8 text-xs font-semibold">45%</span>
                  </div>
                </div>
              </div>

              {/* AI-generated content card */}
              <div className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50/50 p-6 shadow-lg transition-all duration-300 hover:shadow-xl">
                <div className="bg-primary-100 text-primary-700 mb-3 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-semibold">
                  <Zap className="h-3 w-3" aria-hidden="true" focusable="false" />
                  {t('cards.ai.badge')}
                </div>
                <h3 className="mb-2 text-lg font-bold text-slate-900">{t('cards.ai.title')}</h3>
                <p className="mb-4 text-sm text-slate-600">{t('cards.ai.description')}</p>
                {/* Feature icons row */}
                <div className="flex items-center gap-3">
                  <div className="bg-primary-50 flex h-10 w-10 items-center justify-center rounded-lg">
                    <span className="text-lg" aria-hidden="true">
                      📖
                    </span>
                  </div>
                  <div className="bg-accent-50 flex h-10 w-10 items-center justify-center rounded-lg">
                    <span className="text-lg" aria-hidden="true">
                      🖼️
                    </span>
                  </div>
                  <div className="bg-success-50 flex h-10 w-10 items-center justify-center rounded-lg">
                    <span className="text-lg" aria-hidden="true">
                      🔊
                    </span>
                  </div>
                  <div className="bg-secondary-50 flex h-10 w-10 items-center justify-center rounded-lg">
                    <span className="text-lg" aria-hidden="true">
                      💬
                    </span>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

export default EnhancedHeroSection

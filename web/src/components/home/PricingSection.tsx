'use client'

import React, { useState } from 'react'
import { Check, Sparkles } from 'lucide-react'
import { motion } from 'framer-motion'
import { useTranslations } from 'next-intl'

const PricingSection: React.FC = () => {
  const t = useTranslations('pricing')
  const [isAnnual, setIsAnnual] = useState(false)

  const monthlyPrice = 6.99
  const annualPrice = 69.9
  const annualMonthlyEquivalent = (annualPrice / 12).toFixed(2)
  const savings = (((monthlyPrice * 12 - annualPrice) / (monthlyPrice * 12)) * 100).toFixed(0)

  const freePlanFeatures = t.raw('plans.free.features') as string[]

  const premiumFeatures = [
    { text: t('plans.premium.features.unlimitedWordlists'), highlight: true },
    { text: t('plans.premium.features.allQuizModes'), highlight: true },
    { text: t('plans.premium.features.advancedAnalytics'), highlight: false },
    { text: t('plans.premium.features.offlineSupport'), highlight: false },
    { text: t('plans.premium.features.realTimeSpeaking'), highlight: true, badge: t('badges.new') },
    { text: t('plans.premium.features.publicQuizSharing'), highlight: false },
    { text: t('plans.premium.features.errorReporting'), highlight: false },
    { text: t('plans.premium.features.prioritySupport'), highlight: false },
  ]

  return (
    <section
      id="pricing"
      className="scroll-mt-24 bg-gradient-to-b from-slate-50 via-white to-slate-50 py-16 sm:scroll-mt-28 sm:py-20 lg:scroll-mt-32 lg:py-24"
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="mb-12 text-center">
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl"
          >
            {t.rich('title', {
              highlight: (chunks) => (
                <span className="from-primary-500 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
                  {chunks}
                </span>
              ),
            })}
          </motion.h2>
          <motion.p
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="mx-auto max-w-2xl text-lg text-slate-600"
          >
            {t('subtitle')}
          </motion.p>

          {/* Billing Toggle */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.2 }}
            className="mt-8 flex items-center justify-center gap-4"
          >
            <span
              className={`text-sm font-medium ${!isAnnual ? 'text-slate-900' : 'text-slate-500'}`}
            >
              {t('billing.monthly')}
            </span>
            <button
              type="button"
              onClick={() => setIsAnnual(!isAnnual)}
              className={`relative inline-flex h-11 w-16 items-center rounded-full transition-colors ${
                isAnnual ? 'bg-primary-500' : 'bg-slate-300'
              } focus-visible:ring-primary-500 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none`}
              role="switch"
              aria-checked={isAnnual}
              aria-label={t('billing.aria', {
                cycle: t(isAnnual ? 'billing.annualCycle' : 'billing.monthlyCycle'),
              })}
            >
              <span
                className={`inline-block h-7 w-7 transform rounded-full bg-white transition-transform ${
                  isAnnual ? 'translate-x-8' : 'translate-x-1'
                }`}
              />
            </button>
            <span
              className={`text-sm font-medium ${isAnnual ? 'text-slate-900' : 'text-slate-500'}`}
            >
              {t('billing.annual')}
              <span className="bg-success-100 text-success-700 ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-semibold">
                {t('billing.save', { savings })}
              </span>
            </span>
          </motion.div>
        </div>

        {/* Pricing Cards */}
        <div className="mx-auto grid max-w-5xl grid-cols-1 gap-8 lg:grid-cols-5">
          {/* Free Plan - 40% width */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.3 }}
            className="lg:col-span-2"
          >
            <div className="flex h-full flex-col rounded-3xl border-2 border-slate-200/80 bg-white p-8 shadow-sm transition-transform duration-200 hover:-translate-y-0.5">
              <div className="mb-6">
                <h3 className="mb-2 text-2xl font-bold text-slate-900">{t('plans.free.name')}</h3>
                <p className="text-sm text-slate-600">{t('plans.free.description')}</p>
              </div>

              <div className="mb-6">
                <span className="text-5xl font-bold text-slate-900">{t('plans.free.price')}</span>
                <span className="text-slate-600">{t('plans.free.period')}</span>
              </div>

              <a
                href="#download"
                className="hover:border-primary-500 focus-visible:ring-primary-500 mb-8 block w-full rounded-full border-2 border-slate-300 bg-white px-6 py-3 text-center font-semibold text-slate-700 transition-all duration-200 hover:bg-slate-50 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
              >
                {t('plans.free.cta')}
              </a>

              <ul className="flex-1 space-y-3">
                {freePlanFeatures.map((feature) => (
                  <li key={feature} className="flex items-start gap-3 text-sm">
                    <Check
                      className="mt-0.5 h-5 w-5 flex-shrink-0 text-slate-400"
                      strokeWidth={2.5}
                      aria-hidden="true"
                      focusable="false"
                    />
                    <span className="text-slate-600">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
          </motion.div>

          {/* Premium Plan - 60% width */}
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: 0.4 }}
            className="lg:col-span-3"
          >
            <div className="border-primary-500/90 relative flex h-full flex-col rounded-3xl border-2 bg-white p-8 shadow-xl transition-transform duration-200 hover:-translate-y-0.5">
              {/* Most Popular Badge */}
              <div className="absolute -top-4 right-8 z-10">
                <div className="from-primary-500 to-accent-500 flex items-center gap-1.5 rounded-full bg-gradient-to-r px-4 py-1.5 text-xs font-bold text-white shadow-lg">
                  <Sparkles className="h-3.5 w-3.5" aria-hidden="true" focusable="false" />
                  {t('badges.mostPopular')}
                </div>
              </div>

              <div className="mb-6">
                <h3 className="mb-2 text-2xl font-bold text-slate-900">
                  {t('plans.premium.name')}
                </h3>
                <p className="text-sm text-slate-600">
                  {t('plans.premium.description', {
                    cycle: t(isAnnual ? 'billing.annualCycle' : 'billing.monthlyCycle'),
                  })}
                </p>
              </div>

              <div className="mb-6">
                {isAnnual ? (
                  <>
                    <span className="text-5xl font-bold text-slate-900">
                      ${annualMonthlyEquivalent}
                    </span>
                    <span className="text-slate-600">{t('plans.premium.monthlyLabel')}</span>
                    <div className="mt-1 text-sm text-slate-500">
                      {t('plans.premium.annualNote', { price: annualPrice })}
                    </div>
                  </>
                ) : (
                  <>
                    <span className="text-5xl font-bold text-slate-900">${monthlyPrice}</span>
                    <span className="text-slate-600">{t('plans.premium.monthlyLabel')}</span>
                  </>
                )}
              </div>

              <a
                href="#download"
                className="from-primary-500 to-primary-600 focus-visible:ring-primary-500 mb-8 block w-full rounded-full bg-gradient-to-r px-6 py-3 text-center font-semibold text-white shadow-lg transition-all duration-200 hover:scale-[1.02] hover:shadow-xl focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
              >
                {t('plans.premium.cta')}
              </a>

              <ul className="mb-6 flex-1 space-y-3">
                {premiumFeatures.map((feature) => (
                  <li key={feature.text} className="flex items-start gap-3 text-sm">
                    <Check
                      className={`mt-0.5 h-5 w-5 flex-shrink-0 ${feature.highlight ? 'text-primary-500' : 'text-success-500'}`}
                      strokeWidth={2.5}
                      aria-hidden="true"
                      focusable="false"
                    />
                    <span
                      className={
                        feature.highlight ? 'font-semibold text-slate-900' : 'text-slate-600'
                      }
                    >
                      {feature.text}
                      {feature.badge && (
                        <span className="bg-accent-100 text-accent-700 ml-2 inline-flex items-center rounded-full px-2 py-0.5 text-xs font-bold">
                          {feature.badge}
                        </span>
                      )}
                    </span>
                  </li>
                ))}
              </ul>

              {/* Money-back guarantee */}
              <div className="border-success-200 bg-success-50/50 text-success-700 rounded-xl border px-4 py-3 text-center text-sm font-medium">
                {t('guarantee')}
              </div>
            </div>
          </motion.div>
        </div>

        {/* Trust badges */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.5 }}
          className="mt-12 text-center text-sm text-slate-600"
        >
          <div className="flex flex-wrap items-center justify-center gap-6">
            <span>{t('trust.noCard')}</span>
            <span className="hidden sm:inline">•</span>
            <span>{t('trust.cancelAnytime')}</span>
            <span className="hidden sm:inline">•</span>
            <span>{t('trust.secureStripe')}</span>
          </div>
        </motion.div>
      </div>
    </section>
  )
}

export default PricingSection

'use client'

import React from 'react'
import { Sparkles, Brain, Gamepad, Volume2, Cloud, LineChart, AlertCircle } from 'lucide-react'
import { motion } from 'framer-motion'
import { useTranslations } from 'next-intl'

const NewFeaturesSection: React.FC = () => {
  const t = useTranslations('newFeatures')

  // 3 Hero Features
  const heroFeatures = [
    {
      icon: Sparkles,
      title: t('items.aiContent.title'),
      description: t('items.aiContent.description'),
      color: 'from-primary-500 to-primary-600',
      bgColor: 'from-primary-50 to-primary-100',
      stats: t('items.aiContent.stats'),
    },
    {
      icon: Brain,
      title: t('items.spacedRepetition.title'),
      description: t('items.spacedRepetition.description'),
      color: 'from-secondary-500 to-secondary-600',
      bgColor: 'from-secondary-50 to-secondary-100',
      stats: t('items.spacedRepetition.stats'),
    },
    {
      icon: Gamepad,
      title: t('items.quizModes.title'),
      description: t('items.quizModes.description'),
      color: 'from-accent-500 to-accent-600',
      bgColor: 'from-accent-50 to-accent-100',
      stats: t('items.quizModes.stats'),
    },
  ]

  // Additional compact features
  const compactFeatures = [
    {
      icon: Volume2,
      title: t('items.multiLanguageAudio.title'),
      description: t('items.multiLanguageAudio.description'),
    },
    {
      icon: Cloud,
      title: t('items.offlineSupport.title'),
      description: t('items.offlineSupport.description'),
    },
    {
      icon: LineChart,
      title: t('items.analytics.title'),
      description: t('items.analytics.description'),
    },
    {
      icon: AlertCircle,
      title: t('items.errorReporting.title'),
      description: t('items.errorReporting.description'),
    },
  ]

  return (
    <section id="features" className="bg-white py-16 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="mb-12 text-center lg:mb-16">
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl"
          >
            <span>{t('title.part1')}</span>{' '}
            <span className="from-primary-500 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
              {t('title.part2')}
            </span>
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
        </div>

        {/* 3 Hero Features */}
        <div className="mb-12 grid grid-cols-1 gap-6 md:grid-cols-3 lg:gap-8">
          {heroFeatures.map((feature, index) => {
            const IconComponent = feature.icon
            return (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="card-shine group relative overflow-hidden rounded-3xl border border-slate-200 bg-white p-8 shadow-sm transition-all duration-300 hover:-translate-y-1 hover:shadow-xl"
              >
                {/* Icon */}
                <div
                  className={`mb-6 inline-flex rounded-2xl bg-gradient-to-br ${feature.bgColor} p-4`}
                >
                  <IconComponent className="h-8 w-8 text-slate-700" strokeWidth={2} />
                </div>

                {/* Title */}
                <h3 className="mb-3 text-xl font-bold text-slate-900">{feature.title}</h3>

                {/* Description */}
                <p className="mb-4 text-slate-600">{feature.description}</p>

                {/* Stats */}
                <div
                  className={`inline-flex items-center gap-1.5 rounded-full bg-gradient-to-br ${feature.bgColor} px-3 py-1.5 text-xs font-semibold text-slate-700`}
                >
                  <span
                    className={`h-1.5 w-1.5 rounded-full bg-gradient-to-r ${feature.color}`}
                  ></span>
                  {feature.stats}
                </div>

                {/* Hover gradient effect */}
                <div
                  className={`pointer-events-none absolute inset-0 rounded-3xl bg-gradient-to-br ${feature.bgColor} opacity-0 transition-opacity duration-300 group-hover:opacity-10`}
                ></div>
              </motion.div>
            )
          })}
        </div>

        {/* Compact Features Grid */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
          className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4"
        >
          {compactFeatures.map((feature, index) => {
            const IconComponent = feature.icon
            return (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: 0.4 + index * 0.05 }}
                className="group hover:border-primary-300 rounded-2xl border border-slate-200 bg-slate-50/50 p-6 transition-all duration-300 hover:bg-white hover:shadow-md"
              >
                <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-white shadow-sm transition-all duration-300 group-hover:shadow-md">
                  <IconComponent
                    className="group-hover:text-primary-500 h-6 w-6 text-slate-700 transition-colors duration-300"
                    strokeWidth={2}
                  />
                </div>
                <h4 className="mb-2 font-semibold text-slate-900">{feature.title}</h4>
                <p className="text-sm text-slate-600">{feature.description}</p>
              </motion.div>
            )
          })}
        </motion.div>
      </div>
    </section>
  )
}

export default NewFeaturesSection

'use client'

import React from 'react'
import { Users, Globe, BookOpen, Star } from 'lucide-react'
import { motion } from 'framer-motion'
import { statsConfig } from '@/config/statsConfig'
import { useTranslations } from 'next-intl'

const SocialProofSection: React.FC = () => {
  const t = useTranslations('socialProof')
  const tTestimonials = useTranslations('testimonials')

  const stats = [
    {
      icon: Users,
      value: statsConfig.values.userCount,
      label: t('stats.activeLearners'),
      color: 'primary',
    },
    {
      icon: Globe,
      value: statsConfig.values.languageCount,
      label: t('stats.languagesSupported'),
      color: 'secondary',
    },
    {
      icon: BookOpen,
      value: t('stats.wordsLearnedValue'),
      label: t('stats.wordsLearned'),
      color: 'accent',
    },
    {
      icon: Star,
      value: statsConfig.values.starRating,
      label: t('stats.averageRating'),
      color: 'success',
    },
  ]

  const testimonials = [
    {
      name: tTestimonials('testimonial1.author'),
      role: tTestimonials('testimonial1.role'),
      image: '🇪🇸',
      quote: tTestimonials('testimonial1.text'),
      rating: 5,
    },
    {
      name: tTestimonials('testimonial2.author'),
      role: tTestimonials('testimonial2.role'),
      image: '🇯🇵',
      quote: tTestimonials('testimonial2.text'),
      rating: 5,
    },
    {
      name: tTestimonials('testimonial3.author'),
      role: tTestimonials('testimonial3.role'),
      image: '🇩🇪',
      quote: tTestimonials('testimonial3.text'),
      rating: 5,
    },
  ]

  const colorMap: Record<string, string> = {
    primary: 'text-primary-500',
    secondary: 'text-secondary-500',
    accent: 'text-accent-500',
    success: 'text-success-500',
  }

  return (
    <section className="border-y border-slate-200 bg-slate-50 py-16 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="mb-12 text-center">
          <motion.h2
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl"
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
        </div>

        {/* Stats Grid */}
        <div className="mb-16 grid grid-cols-2 gap-6 sm:gap-8 lg:grid-cols-4">
          {stats.map((stat, index) => {
            const IconComponent = stat.icon
            return (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 text-center shadow-sm transition-all duration-300 hover:shadow-lg"
              >
                <div className="mb-3 flex justify-center">
                  <div className={`rounded-full bg-slate-100 p-3 ${colorMap[stat.color]}`}>
                    <IconComponent className="h-6 w-6" strokeWidth={2.5} />
                  </div>
                </div>
                <div className={`mb-1 text-3xl font-bold ${colorMap[stat.color]}`}>
                  {stat.value}
                </div>
                <div className="text-sm font-medium text-slate-600">{stat.label}</div>
                {/* Hover gradient effect */}
                <div
                  className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 transition-opacity duration-300 group-hover:opacity-5"
                  style={{
                    background: `linear-gradient(135deg, ${
                      stat.color === 'primary'
                        ? '#FF6B35'
                        : stat.color === 'secondary'
                          ? '#004E89'
                          : stat.color === 'accent'
                            ? '#7B2CBF'
                            : '#06D6A0'
                    } 0%, transparent 100%)`,
                  }}
                ></div>
              </motion.div>
            )
          })}
        </div>

        {/* Testimonials Carousel */}
        <div className="relative">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="grid grid-cols-1 gap-6 md:grid-cols-3"
          >
            {testimonials.map((testimonial, index) => (
              <motion.div
                key={testimonial.name}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: 0.2 + index * 0.1 }}
                className="group relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg"
              >
                {/* Stars */}
                <div className="mb-4 flex">
                  {Array.from({ length: testimonial.rating }).map((_, i) => (
                    <Star
                      key={i}
                      className="h-4 w-4 fill-yellow-400 text-yellow-400"
                      strokeWidth={0}
                    />
                  ))}
                </div>

                {/* Quote */}
                <p className="mb-6 text-sm leading-relaxed text-slate-600">
                  &ldquo;{testimonial.quote}&rdquo;
                </p>

                {/* Author */}
                <div className="flex items-center gap-3">
                  <div className="flex h-12 w-12 items-center justify-center rounded-full bg-slate-100 text-2xl">
                    {testimonial.image}
                  </div>
                  <div>
                    <div className="font-semibold text-slate-900">{testimonial.name}</div>
                    <div className="text-sm text-slate-500">{testimonial.role}</div>
                  </div>
                </div>

                {/* Hover effect */}
                <div className="from-primary-500/5 pointer-events-none absolute inset-0 rounded-2xl bg-gradient-to-br to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100"></div>
              </motion.div>
            ))}
          </motion.div>
        </div>
      </div>
    </section>
  )
}

export default SocialProofSection

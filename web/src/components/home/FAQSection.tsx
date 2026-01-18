'use client'

import React, { useState } from 'react'
import { useTranslations } from 'next-intl'

const FAQSection: React.FC = () => {
  const [openFAQ, setOpenFAQ] = useState<number | null>(null)
  const t = useTranslations('faq')

  const faqs = [
    {
      id: 1,
      question: t('questions.1.question'),
      answer: t('questions.1.answer'),
    },
    {
      id: 2,
      question: t('questions.2.question'),
      answer: t('questions.2.answer'),
    },
    {
      id: 3,
      question: t('questions.3.question'),
      answer: t('questions.3.answer'),
    },
    {
      id: 4,
      question: t('questions.4.question'),
      answer: t('questions.4.answer'),
    },
    {
      id: 5,
      question: t('questions.5.question'),
      answer: t('questions.5.answer'),
    },
    {
      id: 6,
      question: t('questions.6.question'),
      answer: t('questions.6.answer'),
    },
    {
      id: 7,
      question: t('questions.7.question'),
      answer: t('questions.7.answer'),
    },
    {
      id: 8,
      question: t('questions.8.question'),
      answer: t('questions.8.answer'),
    },
    {
      id: 9,
      question: t('questions.9.question'),
      answer: t('questions.9.answer'),
    },
    {
      id: 10,
      question: t('questions.10.question'),
      answer: t('questions.10.answer'),
    },
  ]

  const toggleFAQ = (id: number) => {
    setOpenFAQ(openFAQ === id ? null : id)
  }

  return (
    <section id="faq" className="bg-gradient-to-b from-white to-slate-50 py-16 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
        <div className="mb-12 text-center lg:mb-16">
          <h2 className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl">
            {t('title.part1')}
            <span className="from-primary-500 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
              {' '}
              {t('title.part2')}
            </span>
          </h2>
          <p className="text-lg text-slate-600">{t('subtitle')}</p>
        </div>

        <div className="space-y-3">
          {faqs.map((faq, index) => (
            <div
              key={faq.id}
              className={`group overflow-hidden rounded-2xl border bg-white transition-all duration-300 ${
                openFAQ === faq.id
                  ? 'border-primary-200 shadow-lg shadow-primary-500/5'
                  : 'border-slate-200 hover:border-primary-200/50 hover:shadow-md'
              }`}
              style={{ animationDelay: `${index * 0.05}s` }}
            >
              <button
                className={`flex w-full items-center justify-between px-6 py-5 text-left transition-colors ${
                  openFAQ === faq.id ? 'bg-primary-50/50' : 'hover:bg-slate-50'
                }`}
                onClick={() => toggleFAQ(faq.id)}
                aria-expanded={openFAQ === faq.id}
              >
                <span className="pr-4 text-base font-semibold text-slate-900 sm:text-lg">{faq.question}</span>
                <span
                  className={`flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full transition-all duration-300 ${
                    openFAQ === faq.id
                      ? 'bg-primary-500 text-white rotate-180'
                      : 'bg-slate-100 text-slate-500 group-hover:bg-primary-100 group-hover:text-primary-600'
                  }`}
                >
                  <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
                  </svg>
                </span>
              </button>
              <div
                className={`transition-all duration-300 ease-in-out ${openFAQ === faq.id ? 'max-h-96' : 'max-h-0 overflow-hidden'}`}
              >
                <p className="px-6 pb-5 leading-relaxed text-slate-600">{faq.answer}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

export default FAQSection

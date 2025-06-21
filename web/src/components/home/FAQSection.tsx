'use client';

import React, { useState } from 'react';
import { useTranslations } from 'next-intl';

const FAQSection: React.FC = () => {
  const [openFAQ, setOpenFAQ] = useState<number | null>(null);
  const t = useTranslations('faq');

  const faqs = [
    {
      id: 1,
      question: t('questions.1.question'),
      answer: t('questions.1.answer')
    },
    {
      id: 2,
      question: t('questions.2.question'),
      answer: t('questions.2.answer')
    },
    {
      id: 3,
      question: t('questions.3.question'),
      answer: t('questions.3.answer')
    },
    {
      id: 4,
      question: t('questions.4.question'),
      answer: t('questions.4.answer')
    },
    {
      id: 5,
      question: t('questions.5.question'),
      answer: t('questions.5.answer')
    },
    {
      id: 6,
      question: t('questions.6.question'),
      answer: t('questions.6.answer')
    },
    {
      id: 7,
      question: t('questions.7.question'),
      answer: t('questions.7.answer')
    },
    {
      id: 8,
      question: t('questions.8.question'),
      answer: t('questions.8.answer')
    },
    {
      id: 9,
      question: t('questions.9.question'),
      answer: t('questions.9.answer')
    },
    {
      id: 10,
      question: t('questions.10.question'),
      answer: t('questions.10.answer')
    }
  ];

  const toggleFAQ = (id: number) => {
    setOpenFAQ(openFAQ === id ? null : id);
  };

  return (
    <section id="faq" className="py-20 bg-white">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            {t('title.part1')} 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> {t('title.part2')}</span>
          </h2>
          <p className="text-xl text-[#636E72]">
            {t('subtitle')}
          </p>
        </div>

        <div className="space-y-4">
          {faqs.map((faq) => (
            <div key={faq.id} className="border border-gray-200 rounded-2xl overflow-hidden hover:border-orange-200 transition-colors">
              <button 
                className="w-full px-6 py-4 text-left flex items-center justify-between hover:bg-orange-50 transition-colors"
                onClick={() => toggleFAQ(faq.id)}
              >
                <span className="text-lg font-semibold">{faq.question}</span>
                <i className={`fas fa-chevron-down text-[#636E72] transition-transform ${openFAQ === faq.id ? 'rotate-180' : ''}`}></i>
              </button>
              <div className={`px-6 transition-all duration-300 ease-in-out ${openFAQ === faq.id ? 'pb-4 max-h-96' : 'max-h-0 overflow-hidden'}`}>
                <p className="text-[#636E72] leading-relaxed">
                  {faq.answer}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default FAQSection;
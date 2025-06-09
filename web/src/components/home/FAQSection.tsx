'use client';

import React, { useState } from 'react';

const FAQSection: React.FC = () => {
  const [openFAQ, setOpenFAQ] = useState<number | null>(null);

  const faqs = [
    {
      id: 1,
      question: "How does the AI content generation work?",
      answer: "Our AI analyzes each word and automatically generates contextually relevant definitions, example sentences, images, and audio pronunciations. This ensures you get comprehensive learning materials for every word you add."
    },
    {
      id: 2,
      question: "What is the Leitner spaced repetition system?",
      answer: "The Leitner system is a scientifically proven method that uses 7 boxes to schedule word reviews based on how well you know them. Words move through boxes with increasing review intervals (1 hour, 1 day, 3 days, 1 week, 2 weeks, 1 month), optimizing your learning efficiency."
    },
    {
      id: 3,
      question: "Can I use Decorebator offline?",
      answer: "Yes! Premium users can download their wordlists and continue learning offline. Your progress syncs automatically when you reconnect to the internet, ensuring you never miss a learning session."
    },
    {
      id: 4,
      question: "What languages are supported?",
      answer: "Decorebator features native AI support for 7 languages: English, Spanish, French, German, Italian, Portuguese, and Japanese. Our AI generates language-specific content with proper grammar rules, cultural context, and optimized voices for each language."
    },
    {
      id: 5,
      question: "How do the different quiz modes work?",
      answer: "We offer 8 quiz modes: meaning matching, word from meaning, image association, audio comprehension, sentence completion, write from definition, audio-to-meaning, and example audio recognition. Each mode targets different aspects of vocabulary learning and adapts based on your current Leitner box level."
    },
    {
      id: 6,
      question: "What makes the multi-language AI special?",
      answer: "Our AI doesn't just translate - it understands each language natively. Spanish definitions include gender-aware grammar, German definitions use the four-case system, Japanese content handles kanji/hiragana appropriately, and each language gets culturally relevant images and optimized voice selection."
    }
  ];

  const toggleFAQ = (id: number) => {
    setOpenFAQ(openFAQ === id ? null : id);
  };

  return (
    <section className="py-20 bg-white">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Frequently Asked 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> Questions</span>
          </h2>
          <p className="text-xl text-[#636E72]">
            Everything you need to know about Decorebator.
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
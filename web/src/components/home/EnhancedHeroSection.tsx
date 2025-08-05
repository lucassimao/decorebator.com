'use client'

import React, { useEffect, useState } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import Image from 'next/image'
import VideoModal from '../common/VideoModal'
import QuizDemoModal from '../quiz/QuizDemoModal'
import { Quiz } from '@/lib/quiz-data'
import AppStoreButton from '../common/AppStoreButton'
import { statsConfig } from '@/config/statsConfig'

interface EnhancedHeroSectionProps {
  demoQuizzes: Quiz[]
}

const EnhancedHeroSection: React.FC<EnhancedHeroSectionProps> = ({ demoQuizzes }) => {
  const [currentWordIndex, setCurrentWordIndex] = useState(0)
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false)
  const [isQuizModalOpen, setIsQuizModalOpen] = useState(false)
  const t = useTranslations('hero')
  const locale = useLocale()

  const words = [
    { key: 'aiIntelligence', text: t('rotatingWords.aiIntelligence') || 'AI Intelligence' },
    { key: 'spacedRepetition', text: t('rotatingWords.spacedRepetition') || 'Spaced Repetition' },
    { key: 'visualLearning', text: t('rotatingWords.visualLearning') || 'Visual Learning' },
    { key: 'smartQuizzes', text: t('rotatingWords.smartQuizzes') || 'Smart Quizzes' },
  ]

  // Font size mapping for each language and word combination
  const getFontSizeClass = (wordKey: string, locale: string) => {
    const fontSizeMap: Record<string, Record<string, string>> = {
      en: {
        aiIntelligence: 'text-4xl sm:text-5xl lg:text-6xl',
        spacedRepetition: 'text-3xl sm:text-4xl lg:text-5xl', // Longer text
        visualLearning: 'text-4xl sm:text-5xl lg:text-6xl',
        smartQuizzes: 'text-4xl sm:text-5xl lg:text-6xl',
      },
      es: {
        aiIntelligence: 'text-3xl sm:text-4xl lg:text-5xl', // "Intelligence Artificial"
        spacedRepetition: 'text-2xl sm:text-3xl lg:text-4xl', // "Repetición Espaciada"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "Aprendizaje Visual"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "Quiz Inteligentes"
      },
      fr: {
        aiIntelligence: 'text-3xl sm:text-4xl lg:text-5xl', // "Intelligence Artificielle"
        spacedRepetition: 'text-2xl sm:text-3xl lg:text-4xl', // "Répétition Espacée"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "Apprentissage Visuel"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "Quiz Intelligents"
      },
      de: {
        aiIntelligence: 'text-3xl sm:text-4xl lg:text-5xl', // "KI-Intelligenz"
        spacedRepetition: 'text-2xl sm:text-3xl lg:text-4xl', // "Wiederholung mit Abstand"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "Visuelles Lernen"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "Intelligente Quiz"
      },
      it: {
        aiIntelligence: 'text-3xl sm:text-4xl lg:text-5xl', // "Intelligence Artificiale"
        spacedRepetition: 'text-2xl sm:text-3xl lg:text-4xl', // "Ripetizione Distanziata"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "Apprendimento Visivo"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "Quiz Intelligenti"
      },
      pt: {
        aiIntelligence: 'text-3xl sm:text-4xl lg:text-5xl', // "Inteligência Artificial"
        spacedRepetition: 'text-2xl sm:text-3xl lg:text-4xl', // "Repetição Espaçada"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "Aprendizado Visual"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "Quiz Inteligentes"
      },
      ja: {
        aiIntelligence: 'text-4xl sm:text-5xl lg:text-6xl', // "AI知能"
        spacedRepetition: 'text-3xl sm:text-4xl lg:text-5xl', // "間隔反復"
        visualLearning: 'text-3xl sm:text-4xl lg:text-5xl', // "視覚学習"
        smartQuizzes: 'text-3xl sm:text-4xl lg:text-5xl', // "スマートクイズ"
      },
    }

    return fontSizeMap[locale]?.[wordKey] || 'text-3xl sm:text-4xl lg:text-5xl'
  }

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentWordIndex((prev) => (prev + 1) % words.length)
    }, 3000)

    return () => clearInterval(interval)
  }, [words.length])

  return (
    <section className="relative overflow-hidden pt-24 pb-12 sm:pb-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="grid items-center gap-8 lg:grid-cols-2 lg:gap-12">
          <div className="slide-in-left order-1 space-y-6 sm:space-y-8 lg:order-1">
            <div className="glass inline-flex items-center rounded-full bg-orange-100/50 px-4 py-2 text-sm font-semibold text-[#FF7B54]">
              <i className="fas fa-zap mr-2"></i>
              {t('tagline')}
            </div>

            <h1 className="text-4xl leading-tight font-bold sm:text-5xl lg:text-6xl">
              {t('title')}
              <span className="mt-2 block flex h-16 items-center sm:h-20 lg:h-24">
                <span
                  className={`gradient-animation bg-clip-text leading-tight font-bold text-transparent transition-all duration-700 ease-in-out ${getFontSizeClass(words[currentWordIndex].key, locale)}`}
                  style={{ transitionProperty: 'opacity, transform, font-size' }}
                >
                  {words[currentWordIndex].text}
                </span>
              </span>
            </h1>

            <p className="-mt-2 text-lg leading-relaxed text-[#636E72] sm:-mt-3 sm:text-xl">
              {t('subtitle')}
            </p>
            {/*
            <div className="flex flex-col sm:flex-row gap-4">
              <button 
                onClick={() => setIsQuizModalOpen(true)}
                className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 flex items-center justify-center"
              >
                <i className="fas fa-brain mr-2 group-hover:scale-110 transition-transform"></i>
                <span>{t('startLearningFree')}</span>
                <i className="fas fa-arrow-right ml-2 group-hover:translate-x-2 transition-transform"></i>
              </button>
              
              <button 
                onClick={() => setIsVideoModalOpen(true)}
                className="group bg-white/80 backdrop-blur px-8 py-4 rounded-full font-semibold text-lg border-2 border-gray-200 hover:border-[#FF7B54] transition-all duration-300 flex items-center justify-center"
              >
                <i className="fas fa-play-circle mr-2 text-[#FF7B54] group-hover:scale-110 transition-transform"></i>
                {t('watchDemo')}
              </button>
            </div>
              */}

            {/* App Store Buttons */}
            <div className="flex flex-wrap gap-4 pt-4">
              <AppStoreButton store="apple" className="h-12" />
              <AppStoreButton store="google" className="h-12" />
            </div>

            {/* Social Proof */}
            {statsConfig.locations.heroSection.showSocialProof && (
              <div className="flex items-center space-x-8 text-sm text-[#636E72]">
                {statsConfig.showStats.starRating && (
                  <div className="flex items-center space-x-2">
                    <div className="flex -space-x-1">
                      <i className="fas fa-star text-yellow-400"></i>
                      <i className="fas fa-star text-yellow-400"></i>
                      <i className="fas fa-star text-yellow-400"></i>
                      <i className="fas fa-star text-yellow-400"></i>
                      <i className="fas fa-star text-yellow-400"></i>
                    </div>
                    <span className="font-semibold">{statsConfig.values.starRating}</span>
                  </div>
                )}
                {statsConfig.showStats.userCount && (
                  <div className="flex items-center space-x-2">
                    <i className="fas fa-users text-[#FF7B54]"></i>
                    <span>{statsConfig.values.userCount}+ active learners</span>
                  </div>
                )}
                {statsConfig.showStats.languageCount && (
                  <div className="flex items-center space-x-2">
                    <i className="fas fa-globe text-[#FF7B54]"></i>
                    <span>{statsConfig.values.totalLanguages}</span>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Hero Visual */}
          <div className="slide-in-right relative order-2 mt-8 lg:order-2 lg:mt-0">
            <div className="relative z-10">
              {/* Phone Mockup */}
              <div className="relative mx-auto w-64 max-w-sm sm:w-72 md:w-80 lg:w-96 lg:max-w-md xl:w-[400px]">
                <div className="card-3d rotate-3 transform rounded-[3rem] bg-gray-900 p-2 shadow-2xl transition-transform duration-700 hover:rotate-0 sm:rotate-6">
                  <div className="overflow-hidden rounded-[2.5rem] bg-white">
                    {/* App Demo Video */}
                    <div className="relative">
                      <video
                        className="h-auto w-full rounded-[2rem] object-contain"
                        autoPlay
                        loop
                        muted
                        playsInline
                        preload="metadata"
                        poster="/app-screenshot.jpeg"
                      >
                        <source src="/hero-demo.webm" type="video/webm" />
                        <source src="/hero-demo.mp4" type="video/mp4" />
                        {/* Fallback for browsers without video support */}
                        <Image
                          src="/app-screenshot.jpeg"
                          alt={t('imageAlt')}
                          width={320}
                          height={678}
                          className="h-auto w-full rounded-[2rem] object-contain"
                          priority
                          sizes="(max-width: 640px) 256px, (max-width: 768px) 288px, (max-width: 1024px) 320px, (max-width: 1280px) 352px, 384px"
                        />
                      </video>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Floating Elements */}
            <div className="pulse-glow absolute -top-4 -right-4 h-20 w-20 rounded-full bg-gradient-to-br from-yellow-400 to-[#FFD700] opacity-80"></div>
            <div className="float-animation absolute -bottom-8 -left-8 h-16 w-16 rounded-full bg-gradient-to-br from-[#4CAF50] to-green-600 opacity-70"></div>
            <div
              className="float-animation absolute top-1/2 -right-12 h-12 w-12 rounded-full bg-gradient-to-br from-purple-400 to-[#9C27B0] opacity-60"
              style={{ animationDelay: '2s' }}
            ></div>
          </div>
        </div>
      </div>

      {/* Video Modal */}
      <VideoModal
        isOpen={isVideoModalOpen}
        onClose={() => setIsVideoModalOpen(false)}
        videoId="dQw4w9WgXcQ" // Replace with actual demo video ID
        title={t('videoTitle')}
      />

      {/* Quiz Demo Modal */}
      <QuizDemoModal
        isOpen={isQuizModalOpen}
        onClose={() => setIsQuizModalOpen(false)}
        demoQuizzes={demoQuizzes}
      />
    </section>
  )
}

export default EnhancedHeroSection

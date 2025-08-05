'use client'

import React, { useState } from 'react'
import { useTranslations } from 'next-intl'
import VideoModal from '../common/VideoModal'
import AppStoreButton from '../common/AppStoreButton'
import { statsConfig } from '@/config/statsConfig'

const CTASection: React.FC = () => {
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false)
  const t = useTranslations('cta')
  // const tCommon = useTranslations('common');

  return (
    <section
      id="download"
      className="relative overflow-hidden bg-gradient-to-r from-[#FF7B54] to-orange-600 py-20"
    >
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="float-animation absolute -top-10 -left-10 h-40 w-40 rounded-full bg-white/10 blur-2xl"></div>
        <div
          className="float-animation absolute -right-10 -bottom-10 h-60 w-60 rounded-full bg-white/10 blur-2xl"
          style={{ animationDelay: '3s' }}
        ></div>
      </div>

      <div className="relative z-10 mx-auto max-w-4xl px-4 text-center sm:px-6 lg:px-8">
        <h2 className="mb-6 text-4xl font-bold text-white lg:text-5xl">{t('title')}</h2>
        <p className="mx-auto mb-8 max-w-2xl text-xl text-white/90">
          {statsConfig.locations.ctaSection.showUserCount
            ? t('subtitleWithCount', { count: statsConfig.values.userCount })
            : t('subtitle')}
        </p>
        {/*
        <div className="flex flex-col sm:flex-row gap-4 justify-center mb-8">
          <button 
            onClick={() => setIsVideoModalOpen(true)}
            className="group bg-white/20 backdrop-blur text-white px-8 py-4 rounded-full font-semibold text-lg border-2 border-white/50 hover:bg-white/30 transition-all duration-300"
          >
            <i className="fas fa-play-circle mr-2"></i>
            <span>{tCommon('watchDemo')}</span>
          </button>
        </div>
         */}
        {/* App Store Buttons */}
        <div className="flex flex-wrap justify-center gap-4">
          <AppStoreButton store="apple" />
          <AppStoreButton store="google" />
        </div>
      </div>

      {/* Video Modal */}
      <VideoModal
        isOpen={isVideoModalOpen}
        onClose={() => setIsVideoModalOpen(false)}
        videoId="dQw4w9WgXcQ" // Replace with actual demo video ID
        title="Decorebator Demo - AI-Powered Vocabulary Learning"
      />
    </section>
  )
}

export default CTASection

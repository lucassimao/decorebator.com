'use client';

import React, { useState } from 'react';
import { useTranslations } from 'next-intl';
import VideoModal from '../common/VideoModal';
import AppStoreButton from '../common/AppStoreButton';
import { statsConfig } from '@/config/statsConfig';

const CTASection: React.FC = () => {
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const t = useTranslations('cta');
  // const tCommon = useTranslations('common');

  return (
    <section id="download" className="py-20 bg-gradient-to-r from-[#FF7B54] to-orange-600 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-10 -left-10 w-40 h-40 bg-white/10 rounded-full blur-2xl float-animation"></div>
        <div className="absolute -bottom-10 -right-10 w-60 h-60 bg-white/10 rounded-full blur-2xl float-animation" style={{animationDelay: '3s'}}></div>
      </div>
      
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center relative z-10">
        <h2 className="text-4xl lg:text-5xl font-bold text-white mb-6">
          {t('title')}
        </h2>
        <p className="text-xl text-white/90 mb-8 max-w-2xl mx-auto">
          {statsConfig.locations.ctaSection.showUserCount 
            ? t('subtitleWithCount', { count: statsConfig.values.userCount })
            : t('subtitle')
          }
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
        <div className="flex flex-wrap gap-4 justify-center">
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
  );
};

export default CTASection;
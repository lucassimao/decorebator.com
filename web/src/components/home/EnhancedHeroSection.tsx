'use client';

import React, { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import Image from 'next/image';
import VideoModal from '../common/VideoModal';
import QuizDemoModal from '../quiz/QuizDemoModal';
import { Quiz } from '@/lib/quiz-data';

interface EnhancedHeroSectionProps {
  demoQuizzes: Quiz[];
}

const EnhancedHeroSection: React.FC<EnhancedHeroSectionProps> = ({ demoQuizzes }) => {
  const [currentWordIndex, setCurrentWordIndex] = useState(0);
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const [isQuizModalOpen, setIsQuizModalOpen] = useState(false);
  const [showAppStoreAlert, setShowAppStoreAlert] = useState(false);
  const t = useTranslations('hero');
  const tCommon = useTranslations('common');
  
  const words = [
    t('rotatingWords.aiIntelligence') || 'AI Intelligence',
    t('rotatingWords.spacedRepetition') || 'Spaced Repetition', 
    t('rotatingWords.visualLearning') || 'Visual Learning',
    t('rotatingWords.smartQuizzes') || 'Smart Quizzes'
  ];

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentWordIndex((prev) => (prev + 1) % words.length);
    }, 3000);

    return () => clearInterval(interval);
  }, [words.length]);

  return (
    <section className="pt-24 pb-12 sm:pb-20 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-8 lg:gap-12 items-center">
          <div className="space-y-6 sm:space-y-8 slide-in-left order-1 lg:order-1">
            <div className="inline-flex items-center px-4 py-2 rounded-full glass bg-orange-100/50 text-[#FF7B54] text-sm font-semibold">
              <i className="fas fa-zap mr-2"></i>
              AI-Powered Learning • 7 Languages • Multi-Platform
            </div>
            
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold leading-tight">
              {t('title')}
              <span className="block mt-2">
                <span className="gradient-animation bg-clip-text text-transparent font-bold transition-all duration-500">
                  {words[currentWordIndex]}
                </span>
              </span>
            </h1>
            
            <p className="text-lg sm:text-xl text-[#636E72] leading-relaxed">
              {t('subtitle', { count: '10,000' })}
            </p>

            <div className="flex flex-col sm:flex-row gap-4">
              <button 
                onClick={() => setIsQuizModalOpen(true)}
                className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 flex items-center justify-center"
              >
                <i className="fas fa-brain mr-2 group-hover:scale-110 transition-transform"></i>
                <span>Try a Quick Quiz</span>
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

            {/* App Store Buttons */}
            <div className="flex flex-wrap gap-4 pt-4">
              <button 
                onClick={() => setShowAppStoreAlert(true)} 
                className="group"
              >
                <Image 
                  src="https://upload.wikimedia.org/wikipedia/commons/3/3c/Download_on_the_App_Store_Badge.svg" 
                  alt="Download on App Store" 
                  width={144}
                  height={48}
                  className="h-12 transform group-hover:scale-105 transition-transform duration-300"
                />
              </button>
              <button 
                onClick={() => setShowAppStoreAlert(true)} 
                className="group"
              >
                <Image 
                  src="https://upload.wikimedia.org/wikipedia/commons/7/78/Google_Play_Store_badge_EN.svg" 
                  alt="Get it on Google Play" 
                  width={144}
                  height={48}
                  className="h-12 transform group-hover:scale-105 transition-transform duration-300"
                />
              </button>
            </div>

            {/* Social Proof */}
            <div className="flex items-center space-x-8 text-sm text-[#636E72]">
              <div className="flex items-center space-x-2">
                <div className="flex -space-x-1">
                  <i className="fas fa-star text-yellow-400"></i>
                  <i className="fas fa-star text-yellow-400"></i>
                  <i className="fas fa-star text-yellow-400"></i>
                  <i className="fas fa-star text-yellow-400"></i>
                  <i className="fas fa-star text-yellow-400"></i>
                </div>
                <span className="font-semibold">{t('stats.rating')}</span>
              </div>
              <div className="flex items-center space-x-2">
                <i className="fas fa-users text-[#FF7B54]"></i>
                <span>{t('stats.activeLearnersCount')}</span>
              </div>
              <div className="flex items-center space-x-2">
                <i className="fas fa-globe text-[#FF7B54]"></i>
                <span>{t('stats.languagesCount')}</span>
              </div>
            </div>
          </div>

          {/* Hero Visual */}
          <div className="relative slide-in-right mt-8 lg:mt-0 order-2 lg:order-2">
            <div className="relative z-10">
              {/* Phone Mockup */}
              <div className="relative mx-auto w-64 sm:w-72 md:w-80 lg:w-96 xl:w-[400px] max-w-sm lg:max-w-md">
                <div className="bg-gray-900 rounded-[3rem] p-2 shadow-2xl transform rotate-3 sm:rotate-6 hover:rotate-0 transition-transform duration-700 card-3d">
                  <div className="bg-white rounded-[2.5rem] overflow-hidden">
                    {/* App Demo Video */}
                    <div className="relative">
                      <video
                        className="w-full h-auto object-contain rounded-[2rem]"
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
                          alt="Decorebator App Demo showing interactive learning features, quiz modes, and progress tracking"
                          width={320}
                          height={678}
                          className="w-full h-auto object-contain rounded-[2rem]"
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
            <div className="absolute -top-4 -right-4 w-20 h-20 bg-gradient-to-br from-yellow-400 to-[#FFD700] rounded-full opacity-80 pulse-glow"></div>
            <div className="absolute -bottom-8 -left-8 w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-full opacity-70 float-animation"></div>
            <div className="absolute top-1/2 -right-12 w-12 h-12 bg-gradient-to-br from-purple-400 to-[#9C27B0] rounded-full opacity-60 float-animation" style={{animationDelay: '2s'}}></div>
          </div>
        </div>
      </div>

      {/* Video Modal */}
      <VideoModal
        isOpen={isVideoModalOpen}
        onClose={() => setIsVideoModalOpen(false)}
        videoId="dQw4w9WgXcQ" // Replace with actual demo video ID
        title="Decorebator Demo - AI-Powered Vocabulary Learning"
      />

      {/* Quiz Demo Modal */}
      <QuizDemoModal
        isOpen={isQuizModalOpen}
        onClose={() => setIsQuizModalOpen(false)}
        demoQuizzes={demoQuizzes}
      />

      {/* App Store Alert Modal */}
      {showAppStoreAlert && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="fixed inset-0 bg-black/50" onClick={() => setShowAppStoreAlert(false)} />
          <div className="bg-white rounded-2xl p-8 max-w-sm w-full relative z-10 shadow-2xl animate-scale-in">
            <div className="text-center">
              <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-4">
                <i className="fas fa-rocket text-white text-2xl"></i>
              </div>
              <h3 className="text-2xl font-bold text-[#2D3436] mb-3">
                {tCommon('appStorePending.title')}
              </h3>
              <p className="text-[#636E72] mb-6">
                {tCommon('appStorePending.message')}
              </p>
              <button
                onClick={() => setShowAppStoreAlert(false)}
                className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-3 rounded-full font-semibold hover:shadow-lg transition-all duration-300"
              >
                {tCommon('appStorePending.okButton')}
              </button>
            </div>
          </div>
        </div>
      )}
    </section>
  );
};

export default EnhancedHeroSection;
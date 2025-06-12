'use client';

import React, { useEffect, useState } from 'react';
import { useTranslations, useLocale } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';
import VideoModal from '../common/VideoModal';

const EnhancedHeroSection: React.FC = () => {
  const [currentWordIndex, setCurrentWordIndex] = useState(0);
  const [isVideoModalOpen, setIsVideoModalOpen] = useState(false);
  const t = useTranslations('hero');
  const locale = useLocale();
  
  const words = [
    t('rotatingWords.aiIntelligence'),
    t('rotatingWords.spacedRepetition'),
    t('rotatingWords.visualLearning'),
    t('rotatingWords.smartQuizzes')
  ];

  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentWordIndex((prev) => (prev + 1) % words.length);
    }, 3000);

    return () => clearInterval(interval);
  }, [words.length]);

  return (
    <section className="pt-24 pb-20 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div className="space-y-8 slide-in-left">
            <div className="inline-flex items-center px-4 py-2 rounded-full glass bg-orange-100/50 text-[#FF7B54] text-sm font-semibold">
              <i className="fas fa-zap mr-2"></i>
              AI-Powered Learning • 7 Languages • Multi-Platform
            </div>
            
            <h1 className="text-5xl lg:text-6xl font-bold leading-tight">
              {t('title')}
              <span className="block mt-2 gradient-animation bg-clip-text text-transparent">
                <div className="word-rotate inline-block">
                  {words.map((word, index) => (
                    <span
                      key={index}
                      className={`${index === currentWordIndex ? 'opacity-100' : 'opacity-0'} transition-opacity duration-1000 absolute`}
                    >
                      {word}
                    </span>
                  ))}
                </div>
              </span>
            </h1>
            
            <p className="text-xl text-[#636E72] leading-relaxed">
              {t('subtitle', { count: '10,000' })}
            </p>

            <div className="flex flex-col sm:flex-row gap-4">
              <Link href={`/${locale}/signup?plan=free`} className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 flex items-center justify-center">
                <span>{t('startLearningFree')}</span>
                <i className="fas fa-arrow-right ml-2 group-hover:translate-x-2 transition-transform"></i>
              </Link>
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
              <a href="#" className="group">
                <Image 
                  src="https://upload.wikimedia.org/wikipedia/commons/3/3c/Download_on_the_App_Store_Badge.svg" 
                  alt="Download on App Store" 
                  width={144}
                  height={48}
                  className="h-12 transform group-hover:scale-105 transition-transform duration-300"
                />
              </a>
              <a href="#" className="group">
                <Image 
                  src="https://upload.wikimedia.org/wikipedia/commons/7/78/Google_Play_Store_badge_EN.svg" 
                  alt="Get it on Google Play" 
                  width={144}
                  height={48}
                  className="h-12 transform group-hover:scale-105 transition-transform duration-300"
                />
              </a>
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
          <div className="relative slide-in-right">
            <div className="relative z-10">
              {/* Phone Mockup */}
              <div className="relative mx-auto w-72 lg:w-80">
                <div className="bg-gray-900 rounded-[3rem] p-3 shadow-2xl transform rotate-6 hover:rotate-0 transition-transform duration-700 card-3d">
                  <div className="bg-white rounded-[2.5rem] overflow-hidden">
                    {/* Status bar */}
                    <div className="bg-white px-6 py-3 flex justify-between items-center text-xs">
                      <span className="font-medium">9:41</span>
                      <div className="flex space-x-1">
                        <i className="fas fa-signal"></i>
                        <i className="fas fa-wifi"></i>
                        <i className="fas fa-battery-full"></i>
                      </div>
                    </div>
                    
                    {/* Quiz Screen */}
                    <div className="p-6 bg-gradient-to-br from-[#FF7B54] to-orange-600 h-96">
                      <div className="text-white">
                        <h3 className="text-lg font-bold mb-2">Today&apos;s Challenge</h3>
                        <div className="bg-white/20 backdrop-blur rounded-2xl p-4 mb-4">
                          <p className="text-sm mb-1 text-orange-100">Question 3 of 10</p>
                          <p className="font-semibold">Which word means &quot;serendipity&quot;?</p>
                        </div>
                        
                        <div className="space-y-3">
                          <button className="w-full p-3 bg-white/90 text-gray-800 rounded-xl font-medium hover:scale-105 transition-transform">
                            A pleasant surprise
                          </button>
                          <button className="w-full p-3 bg-white/20 backdrop-blur text-white rounded-xl hover:bg-white/30 transition-colors">
                            Careful planning
                          </button>
                          <button className="w-full p-3 bg-white/20 backdrop-blur text-white rounded-xl hover:bg-white/30 transition-colors">
                            Deep sadness
                          </button>
                          <button className="w-full p-3 bg-white/20 backdrop-blur text-white rounded-xl hover:bg-white/30 transition-colors">
                            Quick movement
                          </button>
                        </div>
                        
                        <div className="mt-6 flex justify-between items-center">
                          <div className="flex space-x-1">
                            <div className="w-8 h-1 bg-white rounded-full"></div>
                            <div className="w-8 h-1 bg-white rounded-full"></div>
                            <div className="w-8 h-1 bg-white rounded-full"></div>
                            <div className="w-8 h-1 bg-white/40 rounded-full"></div>
                            <div className="w-8 h-1 bg-white/40 rounded-full"></div>
                          </div>
                          <div className="text-sm font-medium">
                            <i className="fas fa-fire mr-1"></i> 15 day streak!
                          </div>
                        </div>
                      </div>
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
    </section>
  );
};

export default EnhancedHeroSection;
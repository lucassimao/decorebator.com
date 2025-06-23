import React from 'react';
import Image from 'next/image';
import { getTranslations } from 'next-intl/server';

const AppShowcaseSection: React.FC = async () => {
  const t = await getTranslations('appShowcase');

  return (
    <section className="py-20 bg-gradient-to-br from-orange-50 via-amber-50 to-yellow-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            <span>{t('title.part1')}</span>
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">{t('title.part2')}</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            {t('subtitle')}
          </p>
        </div>

        {/* Interactive Demo */}
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Feature Tabs */}
          <div className="space-y-6">
            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-orange-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-chart-line text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">{t('features.progressTracking.title')}</h3>
                  <p className="text-[#636E72] mb-3">{t('features.progressTracking.description')}</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-orange-100 text-orange-700 rounded-full text-sm font-medium">{t('features.progressTracking.tags.wordMastery')}</span>
                    <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm font-medium">{t('features.progressTracking.tags.boxDistribution')}</span>
                    <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">{t('features.progressTracking.tags.dailyStreaks')}</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-green-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-robot text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">{t('features.aiContent.title')}</h3>
                  <p className="text-[#636E72] mb-3">{t('features.aiContent.description')}</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm font-medium">{t('features.aiContent.tags.nativeDefinitions')}</span>
                    <span className="px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm font-medium">{t('features.aiContent.tags.culturalImages')}</span>
                    <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">{t('features.aiContent.tags.optimizedAudio')}</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-purple-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-box text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">{t('features.leitnerSystem.title')}</h3>
                  <p className="text-[#636E72] mb-3">{t('features.leitnerSystem.description')}</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm font-medium">{t('features.leitnerSystem.tags.newWords')}</span>
                    <span className="px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm font-medium">{t('features.leitnerSystem.tags.learning')}</span>
                    <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm font-medium">{t('features.leitnerSystem.tags.mastered')}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* App Demo Video */}
          <div className="relative flex justify-center">
            <div className="relative">
              {/* Phone Frame with Video */}
              <div className="w-80 bg-gray-900 rounded-[3rem] p-3 shadow-2xl transform hover:scale-105 transition-transform duration-500">
                <div className="w-full bg-white rounded-[2.5rem] overflow-hidden">
                  <video
                    className="w-full h-auto object-contain rounded-[2rem]"
                    autoPlay
                    loop
                    muted
                    playsInline
                    preload="metadata"
                    poster="/app-screenshot.jpeg"
                  >
                    <source src="/app-demo.webm" type="video/webm" />
                    <source src="/app-demo.mp4" type="video/mp4" />
                    {/* Fallback for very old browsers */}
                    <Image
                      src="/app-demo.gif"
                      alt={t('demo.videoAlt')}
                      width={320}
                      height={678}
                      className="w-full h-auto object-contain rounded-[2rem]"
                      unoptimized
                    />
                  </video>
                </div>
              </div>
              
              {/* Floating UI Elements */}
              <div className="absolute -top-8 -right-8 bg-white rounded-2xl p-4 shadow-xl transform rotate-12 hover:rotate-0 transition-transform duration-300">
                <div className="flex items-center space-x-2">
                  <div className="w-10 h-10 bg-[#FFD700] rounded-full flex items-center justify-center">
                    <i className="fas fa-video text-white"></i>
                  </div>
                  <div>
                    <div className="text-sm font-bold">{t('demo.liveDemo')}</div>
                    <div className="text-xs text-[#636E72]">{t('demo.seeInAction')}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default AppShowcaseSection;
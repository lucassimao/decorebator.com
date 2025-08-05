import React from 'react'
import Image from 'next/image'
import { getTranslations } from 'next-intl/server'

const AppShowcaseSection: React.FC = async () => {
  const t = await getTranslations('appShowcase')

  return (
    <section className="bg-gradient-to-br from-orange-50 via-amber-50 to-yellow-50 py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-16 text-center">
          <h2 className="mb-4 text-4xl font-bold lg:text-5xl">
            <span>{t('title.part1')}</span>
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
              {t('title.part2')}
            </span>
          </h2>
          <p className="mx-auto max-w-3xl text-xl text-[#636E72]">{t('subtitle')}</p>
        </div>

        {/* Interactive Demo */}
        <div className="grid items-center gap-12 lg:grid-cols-2">
          {/* Feature Tabs */}
          <div className="space-y-6">
            <div className="transform cursor-pointer rounded-2xl border border-orange-100 bg-white/80 p-6 shadow-lg backdrop-blur transition-transform duration-300 hover:scale-105">
              <div className="flex items-start space-x-4">
                <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-[#FF7B54] to-orange-600">
                  <i className="fas fa-chart-line text-white"></i>
                </div>
                <div>
                  <h3 className="mb-2 text-xl font-bold">{t('features.progressTracking.title')}</h3>
                  <p className="mb-3 text-[#636E72]">
                    {t('features.progressTracking.description')}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-orange-100 px-3 py-1 text-sm font-medium text-orange-700">
                      {t('features.progressTracking.tags.wordMastery')}
                    </span>
                    <span className="rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-700">
                      {t('features.progressTracking.tags.boxDistribution')}
                    </span>
                    <span className="rounded-full bg-purple-100 px-3 py-1 text-sm font-medium text-purple-700">
                      {t('features.progressTracking.tags.dailyStreaks')}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div className="transform cursor-pointer rounded-2xl border border-green-100 bg-white/80 p-6 shadow-lg backdrop-blur transition-transform duration-300 hover:scale-105">
              <div className="flex items-start space-x-4">
                <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-[#4CAF50] to-green-600">
                  <i className="fas fa-robot text-white"></i>
                </div>
                <div>
                  <h3 className="mb-2 text-xl font-bold">{t('features.aiContent.title')}</h3>
                  <p className="mb-3 text-[#636E72]">{t('features.aiContent.description')}</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-blue-100 px-3 py-1 text-sm font-medium text-blue-700">
                      {t('features.aiContent.tags.nativeDefinitions')}
                    </span>
                    <span className="rounded-full bg-yellow-100 px-3 py-1 text-sm font-medium text-yellow-700">
                      {t('features.aiContent.tags.culturalImages')}
                    </span>
                    <span className="rounded-full bg-purple-100 px-3 py-1 text-sm font-medium text-purple-700">
                      {t('features.aiContent.tags.optimizedAudio')}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <div className="transform cursor-pointer rounded-2xl border border-purple-100 bg-white/80 p-6 shadow-lg backdrop-blur transition-transform duration-300 hover:scale-105">
              <div className="flex items-start space-x-4">
                <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-[#9C27B0] to-purple-600">
                  <i className="fas fa-box text-white"></i>
                </div>
                <div>
                  <h3 className="mb-2 text-xl font-bold">{t('features.leitnerSystem.title')}</h3>
                  <p className="mb-3 text-[#636E72]">{t('features.leitnerSystem.description')}</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-red-100 px-3 py-1 text-sm font-medium text-red-700">
                      {t('features.leitnerSystem.tags.newWords')}
                    </span>
                    <span className="rounded-full bg-yellow-100 px-3 py-1 text-sm font-medium text-yellow-700">
                      {t('features.leitnerSystem.tags.learning')}
                    </span>
                    <span className="rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-700">
                      {t('features.leitnerSystem.tags.mastered')}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* App Demo Video */}
          <div className="relative flex justify-center">
            <div className="relative">
              {/* Phone Frame with Video */}
              <div className="w-80 transform rounded-[3rem] bg-gray-900 p-3 shadow-2xl transition-transform duration-500 hover:scale-105">
                <div className="w-full overflow-hidden rounded-[2.5rem] bg-white">
                  <video
                    className="h-auto w-full rounded-[2rem] object-contain"
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
                      className="h-auto w-full rounded-[2rem] object-contain"
                      unoptimized
                    />
                  </video>
                </div>
              </div>

              {/* Floating UI Elements */}
              <div className="absolute -top-8 -right-8 rotate-12 transform rounded-2xl bg-white p-4 shadow-xl transition-transform duration-300 hover:rotate-0">
                <div className="flex items-center space-x-2">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[#FFD700]">
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
  )
}

export default AppShowcaseSection

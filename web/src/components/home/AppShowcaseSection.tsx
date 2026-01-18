import React from 'react'
import Image from 'next/image'
import { getTranslations } from 'next-intl/server'
import { TrendingUp, Sparkles, Boxes } from 'lucide-react'

const AppShowcaseSection: React.FC = async () => {
  const t = await getTranslations('appShowcase')

  return (
    <section className="bg-gradient-to-b from-slate-50 to-white py-16 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mb-12 text-center lg:mb-16">
          <h2 className="mb-4 text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl lg:text-5xl">
            <span>{t('title.part1')}</span>
            <span className="from-primary-500 to-accent-500 bg-gradient-to-r bg-clip-text text-transparent">
              {t('title.part2')}
            </span>
          </h2>
          <p className="mx-auto max-w-2xl text-lg text-slate-600">{t('subtitle')}</p>
        </div>

        {/* Interactive Demo */}
        <div className="grid items-center gap-8 lg:grid-cols-2 lg:gap-12">
          {/* Feature Cards */}
          <div className="space-y-4">
            {/* Progress Tracking Card */}
            <div className="group hover:border-primary-300 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg">
              <div className="flex items-start gap-4">
                <div className="from-primary-500 to-primary-600 shadow-primary-500/25 flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br shadow-lg">
                  <TrendingUp className="h-6 w-6 text-white" strokeWidth={2.5} />
                </div>
                <div className="flex-1">
                  <h3 className="mb-2 text-lg font-bold text-slate-900">
                    {t('features.progressTracking.title')}
                  </h3>
                  <p className="mb-3 text-sm text-slate-600">
                    {t('features.progressTracking.description')}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <span className="bg-primary-50 text-primary-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.progressTracking.tags.wordMastery')}
                    </span>
                    <span className="bg-success-50 text-success-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.progressTracking.tags.boxDistribution')}
                    </span>
                    <span className="bg-accent-50 text-accent-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.progressTracking.tags.dailyStreaks')}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* AI Content Card */}
            <div className="group hover:border-success-300 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg">
              <div className="flex items-start gap-4">
                <div className="from-success-500 to-success-600 shadow-success-500/25 flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br shadow-lg">
                  <Sparkles className="h-6 w-6 text-white" strokeWidth={2.5} />
                </div>
                <div className="flex-1">
                  <h3 className="mb-2 text-lg font-bold text-slate-900">
                    {t('features.aiContent.title')}
                  </h3>
                  <p className="mb-3 text-sm text-slate-600">
                    {t('features.aiContent.description')}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <span className="bg-secondary-50 text-secondary-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.aiContent.tags.nativeDefinitions')}
                    </span>
                    <span className="bg-primary-50 text-primary-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.aiContent.tags.culturalImages')}
                    </span>
                    <span className="bg-accent-50 text-accent-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.aiContent.tags.optimizedAudio')}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Leitner System Card */}
            <div className="group hover:border-accent-300 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm transition-all duration-300 hover:shadow-lg">
              <div className="flex items-start gap-4">
                <div className="from-accent-500 to-accent-600 shadow-accent-500/25 flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br shadow-lg">
                  <Boxes className="h-6 w-6 text-white" strokeWidth={2.5} />
                </div>
                <div className="flex-1">
                  <h3 className="mb-2 text-lg font-bold text-slate-900">
                    {t('features.leitnerSystem.title')}
                  </h3>
                  <p className="mb-3 text-sm text-slate-600">
                    {t('features.leitnerSystem.description')}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    <span className="rounded-full bg-red-50 px-3 py-1 text-xs font-medium text-red-700">
                      {t('features.leitnerSystem.tags.newWords')}
                    </span>
                    <span className="rounded-full bg-yellow-50 px-3 py-1 text-xs font-medium text-yellow-700">
                      {t('features.leitnerSystem.tags.learning')}
                    </span>
                    <span className="bg-success-50 text-success-700 rounded-full px-3 py-1 text-xs font-medium">
                      {t('features.leitnerSystem.tags.mastered')}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* App Demo Video */}
          <div className="relative flex justify-center lg:justify-end">
            <div className="relative">
              {/* Phone Frame with Video */}
              <div className="group w-72 transform rounded-[2.5rem] bg-gradient-to-br from-slate-900 to-slate-800 p-3 shadow-2xl transition-all duration-500 hover:shadow-[0_25px_80px_-15px_rgba(0,0,0,0.3)] sm:w-80">
                <div className="w-full overflow-hidden rounded-[2rem] bg-white">
                  <video
                    className="h-auto w-full object-contain"
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
                      className="h-auto w-full object-contain"
                      unoptimized
                    />
                  </video>
                </div>
                {/* Glow effect on hover */}
                <div className="from-primary-500 to-accent-500 pointer-events-none absolute -inset-0.5 rounded-[2.5rem] bg-gradient-to-r opacity-0 blur transition-opacity duration-300 group-hover:opacity-20"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

export default AppShowcaseSection

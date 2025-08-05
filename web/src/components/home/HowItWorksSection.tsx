import React from 'react'
import { getTranslations } from 'next-intl/server'

const HowItWorksSection: React.FC = async () => {
  const t = await getTranslations('howItWorks')
  const tCommon = await getTranslations('common')

  return (
    <section id="how-it-works" className="bg-white py-20">
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

        <div className="relative grid gap-8 lg:grid-cols-3">
          {/* Step 1 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600 text-xl font-bold text-white">
              1
            </div>
            <div className="h-full transform rounded-2xl border border-orange-100 bg-gradient-to-br from-orange-50 to-amber-50 p-8 transition-transform duration-300 hover:scale-105">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <i className="fas fa-plus-circle text-3xl text-[#FF7B54]"></i>
              </div>
              <h3 className="mb-4 text-2xl font-bold">{t('step1.title')}</h3>
              <p className="mb-6 text-[#636E72]">{t('step1.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step1.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <i className="fas fa-check-circle mt-1 mr-3 text-[#4CAF50]"></i>
                    <span className="text-sm">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
            {/* Connector */}
            <div className="absolute top-1/2 -right-4 z-0 hidden -translate-y-1/2 transform lg:block">
              <i className="fas fa-arrow-right text-4xl text-orange-300"></i>
            </div>
          </div>

          {/* Step 2 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-[#9C27B0] to-purple-600 text-xl font-bold text-white">
              2
            </div>
            <div className="h-full transform rounded-2xl border border-purple-100 bg-gradient-to-br from-purple-50 to-pink-50 p-8 transition-transform duration-300 hover:scale-105">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <i className="fas fa-gamepad text-3xl text-[#9C27B0]"></i>
              </div>
              <h3 className="mb-4 text-2xl font-bold">{t('step2.title')}</h3>
              <p className="mb-6 text-[#636E72]">{t('step2.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step2.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <i className="fas fa-check-circle mt-1 mr-3 text-[#4CAF50]"></i>
                    <span className="text-sm">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
            {/* Connector */}
            <div className="absolute top-1/2 -right-4 z-0 hidden -translate-y-1/2 transform lg:block">
              <i className="fas fa-arrow-right text-4xl text-purple-300"></i>
            </div>
          </div>

          {/* Step 3 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 z-10 flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-[#4CAF50] to-green-600 text-xl font-bold text-white">
              3
            </div>
            <div className="h-full transform rounded-2xl border border-green-100 bg-gradient-to-br from-green-50 to-emerald-50 p-8 transition-transform duration-300 hover:scale-105">
              <div className="mb-6 flex h-16 w-16 items-center justify-center rounded-2xl bg-white shadow-lg">
                <i className="fas fa-brain text-3xl text-[#4CAF50]"></i>
              </div>
              <h3 className="mb-4 text-2xl font-bold">{t('step3.title')}</h3>
              <p className="mb-6 text-[#636E72]">{t('step3.description')}</p>
              <ul className="space-y-3">
                {(t.raw('step3.features') as string[]).map((feature, index) => (
                  <li key={index} className="flex items-start">
                    <i className="fas fa-check-circle mt-1 mr-3 text-[#4CAF50]"></i>
                    <span className="text-sm">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* CTA */}
        <div className="mt-16 text-center">
          <a
            href="#download"
            className="group inline-block transform rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 px-10 py-4 text-lg font-semibold text-white transition-all duration-300 hover:scale-105 hover:shadow-2xl"
          >
            <span>{tCommon('downloadApp')}</span>
            <i className="fas fa-download ml-2 transition-transform group-hover:translate-x-2"></i>
          </a>
        </div>
      </div>
    </section>
  )
}

export default HowItWorksSection

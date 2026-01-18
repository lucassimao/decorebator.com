import React from 'react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import FooterSection from '../home/FooterSection'
import SmartDownloadButton from '../common/SmartDownloadButton'

interface FeaturePageLayoutProps {
  featureKey: string
  children?: React.ReactNode
}

const FeaturePageLayout: React.FC<FeaturePageLayoutProps> = ({ featureKey, children }) => {
  const t = useTranslations(`featurePages.${featureKey}`)

  return (
    <div className="relative min-h-screen overflow-hidden bg-white text-slate-800">
      {/* Hero Section */}
      <section className="relative z-10 px-4 pt-24 pb-20 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-7xl">
          {/* Breadcrumb */}
          <nav className="mb-8" aria-label="Breadcrumb">
            <ol className="flex items-center space-x-2 text-sm">
              <li>
                <Link
                  href="/"
                  className="hover:text-primary-500 text-slate-500 transition-colors duration-300"
                >
                  {t('common.home')}
                </Link>
              </li>
              <li className="text-slate-500">/</li>
              <li>
                <Link
                  href="/#features"
                  className="hover:text-primary-500 text-slate-500 transition-colors duration-300"
                >
                  {t('common.features')}
                </Link>
              </li>
              <li className="text-slate-500">/</li>
              <li className="font-medium text-slate-800">{t('title')}</li>
            </ol>
          </nav>

          <div className="slide-in-up text-center">
            <h1 className="mb-6 text-4xl font-bold text-slate-800 md:text-5xl lg:text-6xl">
              {t('hero.title')}
            </h1>
            <p className="mx-auto mb-8 max-w-4xl text-xl leading-relaxed text-slate-500 md:text-2xl">
              {t('hero.subtitle')}
            </p>
            <div className="flex flex-col justify-center gap-4 sm:flex-row">
              <SmartDownloadButton
                className="group from-primary-500 flex transform items-center justify-center rounded-full bg-gradient-to-r to-orange-600 px-8 py-4 text-lg font-semibold text-white transition-all duration-300 hover:scale-105 hover:shadow-2xl"
                size="large"
              >
                <span>{t('common.downloadApp')}</span>
                <i className="fas fa-arrow-right ml-2 transition-transform group-hover:translate-x-2"></i>
              </SmartDownloadButton>
              <Link
                href="/#features"
                className="group hover:border-primary-500 flex items-center justify-center rounded-full border-2 border-gray-200 bg-white/80 px-8 py-4 text-lg font-semibold text-slate-800 backdrop-blur transition-all duration-300"
              >
                <i className="fas fa-grid-3x3 text-primary-500 mr-2 transition-transform group-hover:scale-110"></i>
                {t('common.viewAllFeatures')}
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Custom content passed as children */}
      <div className="relative z-10">{children}</div>

      {/* CTA Section */}
      <section className="from-primary-500 relative z-10 bg-gradient-to-r to-orange-600 px-4 py-20 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-4xl text-center">
          <h2 className="mb-6 text-3xl font-bold text-white md:text-4xl">
            {t('common.readyToExperience')}
          </h2>
          <p className="mb-8 text-xl text-orange-100">{t('common.joinThousands')}</p>
          <div className="flex flex-col justify-center gap-4 sm:flex-row">
            <SmartDownloadButton
              className="group text-primary-500 flex transform items-center justify-center rounded-full bg-white px-8 py-4 text-lg font-semibold transition-all duration-300 hover:scale-105 hover:shadow-lg"
              size="large"
            >
              <span>{t('common.downloadApp')}</span>
              <i className="fas fa-rocket ml-2 transition-transform group-hover:translate-x-1"></i>
            </SmartDownloadButton>
            <Link
              href="/#pricing"
              className="group hover:text-primary-500 flex items-center justify-center rounded-full border-2 border-white px-8 py-4 text-lg font-semibold text-white transition-all duration-300 hover:bg-white"
            >
              <i className="fas fa-tag mr-2 transition-transform group-hover:scale-110"></i>
              {t('common.viewPricing')}
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <FooterSection />
    </div>
  )
}

export default FeaturePageLayout

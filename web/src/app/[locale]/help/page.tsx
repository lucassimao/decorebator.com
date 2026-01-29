import React from 'react'
import type { Metadata } from 'next'
import { getTranslations, setRequestLocale } from 'next-intl/server'
import PageLayout from '../../../components/layout/PageLayout'
import AppStoreButton from '../../../components/common/AppStoreButton'

interface HelpPageProps {
  params: Promise<{ locale: string }>
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>
}): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'help' })

  const title = `${t('title')} | Decorebator`
  const description = t('subtitle')

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      url: `https://decorebator.com/${locale}/help`,
      siteName: 'Decorebator',
      images: [
        {
          url: `https://decorebator.com/og?locale=${locale}&page=help`,
          width: 1200,
          height: 630,
          alt: 'Decorebator Help Center',
        },
      ],
      type: 'website',
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [`https://decorebator.com/og?locale=${locale}&page=help`],
    },
    alternates: {
      canonical: `https://decorebator.com/${locale}/help`,
    },
  }
}

const HelpCenterPage: React.FC<HelpPageProps> = async ({ params }) => {
  const { locale } = await params
  setRequestLocale(locale)
  const t = await getTranslations({ locale, namespace: 'help' })
  return (
    <PageLayout>
      <div className="min-h-screen pt-24 pb-20">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-16 text-center">
            <h1 className="mb-4 text-4xl font-bold text-[#2D3436] lg:text-5xl">{t('title')}</h1>
            <p className="mx-auto max-w-3xl text-xl text-[#636E72]">{t('subtitle')}</p>
          </div>

          {/* Quick Navigation */}
          <div className="mb-16 rounded-3xl bg-gradient-to-br from-orange-50 to-amber-50 p-8">
            <h2 className="mb-6 text-center text-2xl font-bold text-[#2D3436]">
              {t('navigation.title')}
            </h2>
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              <a
                href="#getting-started"
                className="flex flex-col items-center rounded-xl bg-white p-4 transition-shadow duration-300 hover:shadow-lg"
              >
                <i className="fas fa-rocket mb-2 text-2xl text-[#FF7B54]"></i>
                <span className="text-sm font-medium">{t('navigation.gettingStarted')}</span>
              </a>
              <a
                href="#features"
                className="flex flex-col items-center rounded-xl bg-white p-4 transition-shadow duration-300 hover:shadow-lg"
              >
                <i className="fas fa-star mb-2 text-2xl text-[#4CAF50]"></i>
                <span className="text-sm font-medium">{t('navigation.features')}</span>
              </a>
              <a
                href="#leitner-system"
                className="flex flex-col items-center rounded-xl bg-white p-4 transition-shadow duration-300 hover:shadow-lg"
              >
                <i className="fas fa-brain mb-2 text-2xl text-[#9C27B0]"></i>
                <span className="text-sm font-medium">{t('navigation.leitnerSystem')}</span>
              </a>
              <a
                href="#troubleshooting"
                className="flex flex-col items-center rounded-xl bg-white p-4 transition-shadow duration-300 hover:shadow-lg"
              >
                <i className="fas fa-tools mb-2 text-2xl text-[#14B8A6]"></i>
                <span className="text-sm font-medium">{t('navigation.troubleshooting')}</span>
              </a>
            </div>
          </div>

          {/* Getting Started Section */}
          <section id="getting-started" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-rocket mr-3 text-[#FF7B54]"></i>
                {t('gettingStarted.title')}
              </h2>

              <div className="grid gap-8 md:grid-cols-2">
                <div>
                  <h3 className="mb-4 text-xl font-bold">
                    {t('gettingStarted.createAccount.title')}
                  </h3>
                  <ul className="space-y-2 text-[#636E72]">
                    {(t.raw('gettingStarted.createAccount.steps') as string[]).map(
                      (step: string, index: number) => (
                        <li key={index}>• {step}</li>
                      )
                    )}
                  </ul>
                </div>

                <div>
                  <h3 className="mb-4 text-xl font-bold">
                    {t('gettingStarted.startLearning.title')}
                  </h3>
                  <ul className="space-y-2 text-[#636E72]">
                    {(t.raw('gettingStarted.startLearning.steps') as string[]).map(
                      (step: string, index: number) => (
                        <li key={index}>• {step}</li>
                      )
                    )}
                  </ul>
                </div>
              </div>

              {/* Download Buttons */}
              <div className="mt-8 text-center">
                <p className="mb-4 text-[#636E72]">{t('gettingStarted.downloadText')}</p>
                <div className="flex flex-wrap justify-center gap-4">
                  <AppStoreButton store="apple" className="h-12" />
                  <AppStoreButton store="google" className="h-12" />
                </div>
              </div>

              <div className="mt-8 rounded-xl border border-blue-200 bg-gradient-to-r from-blue-50 to-indigo-50 p-6">
                <h4 className="mb-2 flex items-center font-bold text-[#2D3436]">
                  <i className="fas fa-lightbulb mr-2 text-blue-600"></i>
                  {t('gettingStarted.proTip.title')}
                </h4>
                <p className="text-[#636E72]">{t('gettingStarted.proTip.text')}</p>
              </div>
            </div>
          </section>

          {/* Features Section */}
          <section id="features" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-star mr-3 text-[#4CAF50]"></i>
                {t('coreFeatures.title')}
              </h2>

              <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#FF7B54]">
                    <i className="fas fa-brain mr-2"></i>
                    {t('coreFeatures.aiPowered.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.aiPowered.description')}
                  </p>
                </div>

                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#4CAF50]">
                    <i className="fas fa-gamepad mr-2"></i>
                    {t('coreFeatures.quizModes.title')}
                  </h4>
                  <ul className="space-y-1 text-sm text-[#636E72]">
                    {(t.raw('coreFeatures.quizModes.modes') as string[]).map(
                      (mode: string, index: number) => (
                        <li key={index}>• {mode}</li>
                      )
                    )}
                  </ul>
                </div>

                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#9C27B0]">
                    <i className="fas fa-clock mr-2"></i>
                    {t('coreFeatures.spacedRepetition.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.spacedRepetition.description')}
                  </p>
                </div>

                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#14B8A6]">
                    <i className="fas fa-chart-line mr-2"></i>
                    {t('coreFeatures.analytics.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.analytics.description')}
                  </p>
                </div>

                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#6366F1]">
                    <i className="fas fa-globe mr-2"></i>
                    {t('coreFeatures.multiLanguage.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.multiLanguage.description')}
                  </p>
                </div>

                <div className="rounded-xl border border-gray-200 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-[#FF6B6B]">
                    <i className="fas fa-wifi-slash mr-2"></i>
                    {t('coreFeatures.offline.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">{t('coreFeatures.offline.description')}</p>
                </div>
              </div>
            </div>
          </section>

          {/* Leitner System Section */}
          <section id="leitner-system" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-brain mr-3 text-[#9C27B0]"></i>
                {t('leitnerSystem.title')}
              </h2>

              <p className="mb-8 text-[#636E72]">{t('leitnerSystem.description')}</p>

              <div className="mb-8 overflow-x-auto">
                <table className="w-full border-collapse rounded-lg border border-gray-200">
                  <thead className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white">
                    <tr>
                      <th className="border border-gray-200 p-4 text-left">
                        {t('leitnerSystem.table.box')}
                      </th>
                      <th className="border border-gray-200 p-4 text-left">
                        {t('leitnerSystem.table.interval')}
                      </th>
                      <th className="border border-gray-200 p-4 text-left">
                        {t('leitnerSystem.table.purpose')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {(
                      t.raw('leitnerSystem.table.boxes') as {
                        box: string
                        interval: string
                        purpose: string
                      }[]
                    ).map(
                      (
                        boxData: { box: string; interval: string; purpose: string },
                        index: number
                      ) => {
                        const colorClasses = [
                          'bg-red-50',
                          'bg-orange-50',
                          'bg-yellow-50',
                          'bg-blue-50',
                          'bg-purple-50',
                          'bg-indigo-50',
                          'bg-green-50',
                        ]
                        return (
                          <tr key={index} className={colorClasses[index]}>
                            <td className="border border-gray-200 p-4 font-bold">{boxData.box}</td>
                            <td className="border border-gray-200 p-4">{boxData.interval}</td>
                            <td className="border border-gray-200 p-4">{boxData.purpose}</td>
                          </tr>
                        )
                      }
                    )}
                  </tbody>
                </table>
              </div>

              <div className="grid gap-8 md:grid-cols-2">
                <div className="rounded-xl border border-green-200 bg-green-50 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-green-800">
                    <i className="fas fa-check-circle mr-2"></i>
                    {t('leitnerSystem.progressionRules.title')}
                  </h4>
                  <ul className="space-y-2 text-green-700">
                    <li>
                      • <strong>{t('leitnerSystem.progressionRules.correct')}</strong>
                    </li>
                    <li>
                      • <strong>{t('leitnerSystem.progressionRules.incorrect')}</strong>
                    </li>
                    <li>
                      • <strong>{t('leitnerSystem.progressionRules.mastered')}</strong>
                    </li>
                  </ul>
                </div>

                <div className="rounded-xl border border-blue-200 bg-blue-50 p-6">
                  <h4 className="mb-3 flex items-center font-bold text-blue-800">
                    <i className="fas fa-graduation-cap mr-2"></i>
                    {t('leitnerSystem.quizProgression.title')}
                  </h4>
                  <p className="text-sm text-blue-700">
                    {t('leitnerSystem.quizProgression.description')}
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Subscription Plans */}
          <section id="plans" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-crown mr-3 text-[#FFD700]"></i>
                {t('subscriptionPlans.title')}
              </h2>

              <div className="mb-6 rounded-lg border border-blue-200 bg-blue-50 p-4">
                <p className="text-sm text-blue-700">
                  <i className="fas fa-info-circle mr-2"></i>
                  {t('subscriptionPlans.appNotice')}
                </p>
              </div>

              <div className="grid gap-6 md:grid-cols-3">
                <div className="rounded-xl border border-gray-200 p-6 text-center">
                  <h3 className="mb-4 text-xl font-bold text-[#4CAF50]">
                    {t('subscriptionPlans.free.title')}
                  </h3>
                  <div className="mb-4 text-3xl font-bold">{t('subscriptionPlans.free.price')}</div>
                  <ul className="mb-6 space-y-2 text-sm text-[#636E72]">
                    {(t.raw('subscriptionPlans.free.features') as string[]).map(
                      (feature: string, index: number) => (
                        <li key={index}>• {feature}</li>
                      )
                    )}
                  </ul>
                  <div className="flex flex-wrap justify-center gap-3">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>

                <div className="relative rounded-xl border-2 border-[#FF7B54] p-6 text-center">
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 transform rounded-full bg-[#FF7B54] px-4 py-1 text-sm font-semibold text-white">
                    {t('subscriptionPlans.monthly.badge')}
                  </div>
                  <h3 className="mb-4 text-xl font-bold text-[#FF7B54]">
                    {t('subscriptionPlans.monthly.title')}
                  </h3>
                  <div className="mb-4 text-3xl font-bold">
                    {t('subscriptionPlans.monthly.price')}
                    <span className="text-sm text-[#636E72]">
                      {t('subscriptionPlans.monthly.period')}
                    </span>
                  </div>
                  <ul className="mb-6 space-y-2 text-sm text-[#636E72]">
                    {(t.raw('subscriptionPlans.monthly.features') as string[]).map(
                      (feature: string, index: number) => (
                        <li key={index}>• {feature}</li>
                      )
                    )}
                  </ul>
                  <div className="flex flex-wrap justify-center gap-3">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>

                <div className="rounded-xl border border-gray-200 p-6 text-center">
                  <h3 className="mb-4 text-xl font-bold text-[#9C27B0]">
                    {t('subscriptionPlans.annual.title')}
                  </h3>
                  <div className="mb-2 text-3xl font-bold">
                    {t('subscriptionPlans.annual.price')}
                    <span className="text-sm text-[#636E72]">
                      {t('subscriptionPlans.annual.period')}
                    </span>
                  </div>
                  <div className="mb-4 text-sm font-semibold text-[#4CAF50]">
                    {t('subscriptionPlans.annual.savings')}
                  </div>
                  <ul className="mb-6 space-y-2 text-sm text-[#636E72]">
                    {(t.raw('subscriptionPlans.annual.features') as string[]).map(
                      (feature: string, index: number) => (
                        <li key={index}>• {feature}</li>
                      )
                    )}
                  </ul>
                  <div className="flex flex-wrap justify-center gap-3">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Language Support */}
          <section id="languages" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-globe mr-3 text-[#6366F1]"></i>
                {t('languageSupport.title')}
              </h2>

              <p className="mb-8 text-[#636E72]">{t('languageSupport.description')}</p>

              <div className="grid gap-8 md:grid-cols-2">
                <div>
                  <h3 className="mb-4 text-xl font-bold">
                    {t('languageSupport.supportedLanguages')}
                  </h3>
                  <div className="grid grid-cols-2 gap-4">
                    {(
                      t.raw('languageSupport.languages') as {
                        flag: string
                        name: string
                        bgClass: string
                      }[]
                    ).map(
                      (lang: { flag: string; name: string; bgClass: string }, index: number) => (
                        <div
                          key={index}
                          className={`flex items-center space-x-3 p-3 ${lang.bgClass} rounded-lg`}
                        >
                          <span className="text-2xl">{lang.flag}</span>
                          <span className="font-semibold">{lang.name}</span>
                        </div>
                      )
                    )}
                  </div>
                </div>

                <div>
                  <h3 className="mb-4 text-xl font-bold">{t('languageSupport.aiFeatures')}</h3>
                  <ul className="space-y-3 text-[#636E72]">
                    {(t.raw('languageSupport.features') as string[]).map(
                      (feature: string, index: number) => (
                        <li key={index} className="flex items-start">
                          <i className="fas fa-check-circle mt-1 mr-3 text-[#4CAF50]"></i>
                          <span>{feature}</span>
                        </li>
                      )
                    )}
                  </ul>
                </div>
              </div>
            </div>
          </section>

          {/* Troubleshooting Section */}
          <section id="troubleshooting" className="mb-16">
            <div className="rounded-3xl bg-white p-8 shadow-xl">
              <h2 className="mb-6 flex items-center text-3xl font-bold text-[#2D3436]">
                <i className="fas fa-tools mr-3 text-[#14B8A6]"></i>
                {t('troubleshooting.title')}
              </h2>

              <div className="grid gap-8 md:grid-cols-2">
                <div>
                  <h3 className="mb-4 text-xl font-bold">{t('troubleshooting.commonIssues')}</h3>
                  <div className="space-y-4">
                    <div className="rounded-lg border border-gray-200 p-4">
                      <h4 className="mb-2 font-semibold text-[#2D3436]">
                        {t('troubleshooting.audioNotPlaying.title')}
                      </h4>
                      <p className="text-sm text-[#636E72]">
                        {t('troubleshooting.audioNotPlaying.solution')}
                      </p>
                    </div>

                    <div className="rounded-lg border border-gray-200 p-4">
                      <h4 className="mb-2 font-semibold text-[#2D3436]">
                        {t('troubleshooting.imagesNotLoading.title')}
                      </h4>
                      <p className="text-sm text-[#636E72]">
                        {t('troubleshooting.imagesNotLoading.solution')}
                      </p>
                    </div>

                    <div className="rounded-lg border border-gray-200 p-4">
                      <h4 className="mb-2 font-semibold text-[#2D3436]">
                        {t('troubleshooting.subscriptionNotUpdating.title')}
                      </h4>
                      <p className="text-sm text-[#636E72]">
                        {t('troubleshooting.subscriptionNotUpdating.solution')}
                      </p>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="mb-4 text-xl font-bold">
                    {t('troubleshooting.errorReporting.title')}
                  </h3>
                  <p className="mb-4 text-[#636E72]">
                    {t('troubleshooting.errorReporting.description')}
                  </p>
                  <ul className="mb-6 space-y-2 text-[#636E72]">
                    {(t.raw('troubleshooting.errorReporting.types') as string[]).map(
                      (type: string, index: number) => (
                        <li key={index}>• {type}</li>
                      )
                    )}
                  </ul>

                  <div className="rounded-lg border border-blue-200 bg-blue-50 p-4">
                    <h4 className="mb-2 font-semibold text-blue-800">
                      {t('troubleshooting.needHelp.title')}
                    </h4>
                    <p className="mb-3 text-sm text-blue-700">
                      {t('troubleshooting.needHelp.description')}
                    </p>
                    <a
                      href="mailto:support@decorebator.com"
                      className="text-sm font-semibold text-blue-600 hover:text-blue-800"
                    >
                      {t('troubleshooting.needHelp.email')}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Contact Support */}
          <section className="text-center">
            <div className="rounded-3xl bg-gradient-to-r from-[#FF7B54] to-orange-600 p-8 text-white">
              <h2 className="mb-4 text-3xl font-bold">{t('stillNeedHelp.title')}</h2>
              <p className="mb-6 text-xl opacity-90">{t('stillNeedHelp.subtitle')}</p>
              <a
                href="mailto:support@decorebator.com"
                className="inline-block rounded-full bg-white px-8 py-3 font-semibold text-[#FF7B54] transition-all duration-300 hover:shadow-lg"
              >
                <i className="fas fa-envelope mr-2"></i>
                {t('stillNeedHelp.emailSupport')}
              </a>
            </div>
          </section>
        </div>
      </div>
    </PageLayout>
  )
}

export default HelpCenterPage

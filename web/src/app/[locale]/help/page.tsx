import React from 'react';
import { getTranslations } from 'next-intl/server';
import PageLayout from '../../../components/layout/PageLayout';
import AppStoreButton from '../../../components/common/AppStoreButton';

const HelpCenterPage: React.FC = async () => {
  const t = await getTranslations('help');
  return (
    <PageLayout>
      <div className="pt-24 pb-20 min-h-screen">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="text-center mb-16">
            <h1 className="text-4xl lg:text-5xl font-bold text-[#2D3436] mb-4">
              {t('title')}
            </h1>
            <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
              {t('subtitle')}
            </p>
          </div>

          {/* Quick Navigation */}
          <div className="bg-gradient-to-br from-orange-50 to-amber-50 rounded-3xl p-8 mb-16">
            <h2 className="text-2xl font-bold text-[#2D3436] mb-6 text-center">{t('navigation.title')}</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <a href="#getting-started" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-rocket text-2xl text-[#FF7B54] mb-2"></i>
                <span className="text-sm font-medium">{t('navigation.gettingStarted')}</span>
              </a>
              <a href="#features" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-star text-2xl text-[#4CAF50] mb-2"></i>
                <span className="text-sm font-medium">{t('navigation.features')}</span>
              </a>
              <a href="#leitner-system" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-brain text-2xl text-[#9C27B0] mb-2"></i>
                <span className="text-sm font-medium">{t('navigation.leitnerSystem')}</span>
              </a>
              <a href="#troubleshooting" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-tools text-2xl text-[#14B8A6] mb-2"></i>
                <span className="text-sm font-medium">{t('navigation.troubleshooting')}</span>
              </a>
            </div>
          </div>

          {/* Getting Started Section */}
          <section id="getting-started" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-rocket text-[#FF7B54] mr-3"></i>
                {t('gettingStarted.title')}
              </h2>
              
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">{t('gettingStarted.createAccount.title')}</h3>
                  <ul className="space-y-2 text-[#636E72]">
                    {(t.raw('gettingStarted.createAccount.steps') as string[]).map((step: string, index: number) => (
                      <li key={index}>• {step}</li>
                    ))}
                  </ul>
                </div>
                
                <div>
                  <h3 className="text-xl font-bold mb-4">{t('gettingStarted.startLearning.title')}</h3>
                  <ul className="space-y-2 text-[#636E72]">
                    {(t.raw('gettingStarted.startLearning.steps') as string[]).map((step: string, index: number) => (
                      <li key={index}>• {step}</li>
                    ))}
                  </ul>
                </div>
              </div>
              
              {/* Download Buttons */}
              <div className="mt-8 text-center">
                <p className="text-[#636E72] mb-4">{t('gettingStarted.downloadText')}</p>
                <div className="flex flex-wrap gap-4 justify-center">
                  <AppStoreButton store="apple" className="h-12" />
                  <AppStoreButton store="google" className="h-12" />
                </div>
              </div>

              <div className="mt-8 p-6 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-200">
                <h4 className="font-bold text-[#2D3436] mb-2 flex items-center">
                  <i className="fas fa-lightbulb text-blue-600 mr-2"></i>
                  {t('gettingStarted.proTip.title')}
                </h4>
                <p className="text-[#636E72]">
                  {t('gettingStarted.proTip.text')}
                </p>
              </div>
            </div>
          </section>

          {/* Features Section */}
          <section id="features" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-star text-[#4CAF50] mr-3"></i>
                {t('coreFeatures.title')}
              </h2>

              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#FF7B54] mb-3 flex items-center">
                    <i className="fas fa-brain mr-2"></i>
                    {t('coreFeatures.aiPowered.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.aiPowered.description')}
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#4CAF50] mb-3 flex items-center">
                    <i className="fas fa-gamepad mr-2"></i>
                    {t('coreFeatures.quizModes.title')}
                  </h4>
                  <ul className="text-sm text-[#636E72] space-y-1">
                    {(t.raw('coreFeatures.quizModes.modes') as string[]).map((mode: string, index: number) => (
                      <li key={index}>• {mode}</li>
                    ))}
                  </ul>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#9C27B0] mb-3 flex items-center">
                    <i className="fas fa-clock mr-2"></i>
                    {t('coreFeatures.spacedRepetition.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.spacedRepetition.description')}
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#14B8A6] mb-3 flex items-center">
                    <i className="fas fa-chart-line mr-2"></i>
                    {t('coreFeatures.analytics.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.analytics.description')}
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#6366F1] mb-3 flex items-center">
                    <i className="fas fa-globe mr-2"></i>
                    {t('coreFeatures.multiLanguage.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.multiLanguage.description')}
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#FF6B6B] mb-3 flex items-center">
                    <i className="fas fa-wifi-slash mr-2"></i>
                    {t('coreFeatures.offline.title')}
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    {t('coreFeatures.offline.description')}
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Leitner System Section */}
          <section id="leitner-system" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-brain text-[#9C27B0] mr-3"></i>
                {t('leitnerSystem.title')}
              </h2>
              
              <p className="text-[#636E72] mb-8">
                {t('leitnerSystem.description')}
              </p>

              <div className="overflow-x-auto mb-8">
                <table className="w-full border-collapse border border-gray-200 rounded-lg">
                  <thead className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white">
                    <tr>
                      <th className="border border-gray-200 p-4 text-left">{t('leitnerSystem.table.box')}</th>
                      <th className="border border-gray-200 p-4 text-left">{t('leitnerSystem.table.interval')}</th>
                      <th className="border border-gray-200 p-4 text-left">{t('leitnerSystem.table.purpose')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(t.raw('leitnerSystem.table.boxes') as {box: string, interval: string, purpose: string}[]).map((boxData: {box: string, interval: string, purpose: string}, index: number) => {
                      const colorClasses = [
                        'bg-red-50', 'bg-orange-50', 'bg-yellow-50', 'bg-blue-50', 
                        'bg-purple-50', 'bg-indigo-50', 'bg-green-50'
                      ];
                      return (
                        <tr key={index} className={colorClasses[index]}>
                          <td className="border border-gray-200 p-4 font-bold">{boxData.box}</td>
                          <td className="border border-gray-200 p-4">{boxData.interval}</td>
                          <td className="border border-gray-200 p-4">{boxData.purpose}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>

              <div className="grid md:grid-cols-2 gap-8">
                <div className="bg-green-50 border border-green-200 rounded-xl p-6">
                  <h4 className="font-bold text-green-800 mb-3 flex items-center">
                    <i className="fas fa-check-circle mr-2"></i>
                    {t('leitnerSystem.progressionRules.title')}
                  </h4>
                  <ul className="text-green-700 space-y-2">
                    <li>• <strong>{t('leitnerSystem.progressionRules.correct')}</strong></li>
                    <li>• <strong>{t('leitnerSystem.progressionRules.incorrect')}</strong></li>
                    <li>• <strong>{t('leitnerSystem.progressionRules.mastered')}</strong></li>
                  </ul>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-xl p-6">
                  <h4 className="font-bold text-blue-800 mb-3 flex items-center">
                    <i className="fas fa-graduation-cap mr-2"></i>
                    {t('leitnerSystem.quizProgression.title')}
                  </h4>
                  <p className="text-blue-700 text-sm">
                    {t('leitnerSystem.quizProgression.description')}
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Subscription Plans */}
          <section id="plans" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-crown text-[#FFD700] mr-3"></i>
                {t('subscriptionPlans.title')}
              </h2>
              
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
                <p className="text-blue-700 text-sm">
                  <i className="fas fa-info-circle mr-2"></i>
                  {t('subscriptionPlans.appNotice')}
                </p>
              </div>

              <div className="grid md:grid-cols-3 gap-6">
                <div className="border border-gray-200 rounded-xl p-6 text-center">
                  <h3 className="text-xl font-bold text-[#4CAF50] mb-4">{t('subscriptionPlans.free.title')}</h3>
                  <div className="text-3xl font-bold mb-4">{t('subscriptionPlans.free.price')}</div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    {(t.raw('subscriptionPlans.free.features') as string[]).map((feature: string, index: number) => (
                      <li key={index}>• {feature}</li>
                    ))}
                  </ul>
                  <div className="flex flex-wrap gap-3 justify-center">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>

                <div className="border-2 border-[#FF7B54] rounded-xl p-6 text-center relative">
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2 bg-[#FF7B54] text-white px-4 py-1 rounded-full text-sm font-semibold">
                    {t('subscriptionPlans.monthly.badge')}
                  </div>
                  <h3 className="text-xl font-bold text-[#FF7B54] mb-4">{t('subscriptionPlans.monthly.title')}</h3>
                  <div className="text-3xl font-bold mb-4">{t('subscriptionPlans.monthly.price')}<span className="text-sm text-[#636E72]">{t('subscriptionPlans.monthly.period')}</span></div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    {(t.raw('subscriptionPlans.monthly.features') as string[]).map((feature: string, index: number) => (
                      <li key={index}>• {feature}</li>
                    ))}
                  </ul>
                  <div className="flex flex-wrap gap-3 justify-center">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>

                <div className="border border-gray-200 rounded-xl p-6 text-center">
                  <h3 className="text-xl font-bold text-[#9C27B0] mb-4">{t('subscriptionPlans.annual.title')}</h3>
                  <div className="text-3xl font-bold mb-2">{t('subscriptionPlans.annual.price')}<span className="text-sm text-[#636E72]">{t('subscriptionPlans.annual.period')}</span></div>
                  <div className="text-sm text-[#4CAF50] font-semibold mb-4">{t('subscriptionPlans.annual.savings')}</div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    {(t.raw('subscriptionPlans.annual.features') as string[]).map((feature: string, index: number) => (
                      <li key={index}>• {feature}</li>
                    ))}
                  </ul>
                  <div className="flex flex-wrap gap-3 justify-center">
                    <AppStoreButton store="apple" className="h-11" />
                    <AppStoreButton store="google" className="h-11" />
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Language Support */}
          <section id="languages" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-globe text-[#6366F1] mr-3"></i>
                {t('languageSupport.title')}
              </h2>

              <p className="text-[#636E72] mb-8">
                {t('languageSupport.description')}
              </p>

              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">{t('languageSupport.supportedLanguages')}</h3>
                  <div className="grid grid-cols-2 gap-4">
                    {(t.raw('languageSupport.languages') as {flag: string, name: string, bgClass: string}[]).map((lang: {flag: string, name: string, bgClass: string}, index: number) => (
                      <div key={index} className={`flex items-center space-x-3 p-3 ${lang.bgClass} rounded-lg`}>
                        <span className="text-2xl">{lang.flag}</span>
                        <span className="font-semibold">{lang.name}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-bold mb-4">{t('languageSupport.aiFeatures')}</h3>
                  <ul className="space-y-3 text-[#636E72]">
                    {(t.raw('languageSupport.features') as string[]).map((feature: string, index: number) => (
                      <li key={index} className="flex items-start">
                        <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                        <span>{feature}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          </section>

          {/* Troubleshooting Section */}
          <section id="troubleshooting" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-tools text-[#14B8A6] mr-3"></i>
                {t('troubleshooting.title')}
              </h2>

              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">{t('troubleshooting.commonIssues')}</h3>
                  <div className="space-y-4">
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">{t('troubleshooting.audioNotPlaying.title')}</h4>
                      <p className="text-sm text-[#636E72]">{t('troubleshooting.audioNotPlaying.solution')}</p>
                    </div>
                    
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">{t('troubleshooting.imagesNotLoading.title')}</h4>
                      <p className="text-sm text-[#636E72]">{t('troubleshooting.imagesNotLoading.solution')}</p>
                    </div>
                    
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">{t('troubleshooting.subscriptionNotUpdating.title')}</h4>
                      <p className="text-sm text-[#636E72]">{t('troubleshooting.subscriptionNotUpdating.solution')}</p>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-bold mb-4">{t('troubleshooting.errorReporting.title')}</h3>
                  <p className="text-[#636E72] mb-4">
                    {t('troubleshooting.errorReporting.description')}
                  </p>
                  <ul className="space-y-2 text-[#636E72] mb-6">
                    {(t.raw('troubleshooting.errorReporting.types') as string[]).map((type: string, index: number) => (
                      <li key={index}>• {type}</li>
                    ))}
                  </ul>
                  
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <h4 className="font-semibold text-blue-800 mb-2">{t('troubleshooting.needHelp.title')}</h4>
                    <p className="text-blue-700 text-sm mb-3">
                      {t('troubleshooting.needHelp.description')}
                    </p>
                    <a href="mailto:support@decorebator.com" className="text-blue-600 hover:text-blue-800 font-semibold text-sm">
                      {t('troubleshooting.needHelp.email')}
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Contact Support */}
          <section className="text-center">
            <div className="bg-gradient-to-r from-[#FF7B54] to-orange-600 rounded-3xl p-8 text-white">
              <h2 className="text-3xl font-bold mb-4">{t('stillNeedHelp.title')}</h2>
              <p className="text-xl mb-6 opacity-90">
                {t('stillNeedHelp.subtitle')}
              </p>
              <a
                href="mailto:support@decorebator.com"
                className="bg-white text-[#FF7B54] px-8 py-3 rounded-full font-semibold hover:shadow-lg transition-all duration-300 inline-block"
              >
                <i className="fas fa-envelope mr-2"></i>
                {t('stillNeedHelp.emailSupport')}
              </a>
            </div>
          </section>
        </div>
      </div>
    </PageLayout>
  );
};

export default HelpCenterPage;
import React from 'react';
import { Metadata } from 'next';
import { useTranslations } from 'next-intl';
import { generateFeatureMetadata, generateFeatureStructuredData } from '../../../../utils/featureMetadata';
import FeaturePageLayout from '../../../../components/features/FeaturePageLayout';
import PageLayout from '../../../../components/layout/PageLayout';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  return generateFeatureMetadata({
    featureKey: 'offlineSupport',
    locale,
    keywords: 'offline learning, offline flashcards, download wordlists, mobile learning, commute study',
    ogImage: 'https://decorebator.com/og-offline-support.jpg'
  });
}

const OfflineSupportFeaturePage: React.FC = () => {
  const t = useTranslations('featurePages.offlineSupport');
  const structuredData = generateFeatureStructuredData('offlineSupport', t);

  return (
    <PageLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <FeaturePageLayout featureKey="offlineSupport">
        {/* Offline Features Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                  {t('features.title')}
                </span>
              </h2>
            </div>
            
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Download Wordlists</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.download')}</p>
              </div>

              <div className="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-gray-800 dark:to-gray-700 p-6 rounded-xl">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Full Quiz Functionality</h3>
                <p className="text-[#636E72] leading-relaxed">{t('features.quizzes')}</p>
              </div>

              <div className="bg-gradient-to-br from-purple-50 to-pink-50 dark:from-gray-800 dark:to-gray-700 p-6 rounded-xl">
                <div className="w-16 h-16 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Offline Flashcards</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.flashcards')}</p>
              </div>

              <div className="bg-gradient-to-br from-orange-50 to-red-50 dark:from-gray-800 dark:to-gray-700 p-6 rounded-xl">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FFD700] to-yellow-500 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Downloaded Audio</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.audio')}</p>
              </div>

              <div className="bg-gradient-to-br from-yellow-50 to-amber-50 dark:from-gray-800 dark:to-gray-700 p-6 rounded-xl">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Local Progress Tracking</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.progress')}</p>
              </div>

              <div className="bg-gradient-to-br from-teal-50 to-cyan-50 dark:from-gray-800 dark:to-gray-700 p-6 rounded-xl">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Automatic Sync</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.sync')}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Benefits Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                  {t('benefits.title')}
                </span>
              </h2>
            </div>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              {t.raw('benefits.items').map((benefit: string, index: number) => (
                <div key={index} className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                  <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                  <p className="text-[#636E72] leading-relaxed text-center">{benefit}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* How It Works Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                  {t('howItWorks.title')}
                </span>
              </h2>
            </div>
            <div className="grid md:grid-cols-3 gap-8">
              {t.raw('howItWorks.steps').map((step: { title: string; description: string }, index: number) => (
                <div key={index} className="text-center">
                  <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6 shadow-lg">
                    <span className="text-2xl font-bold text-white">{index + 1}</span>
                  </div>
                  <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300">
                    {step.title}
                  </h3>
                  <p className="text-[#636E72] leading-relaxed">
                    {step.description}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Premium Feature Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
          <div className="max-w-4xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                {t('premium.title')}
              </span>
            </h2>
            <p className="text-xl text-[#636E72] leading-relaxed mb-8">
              {t('premium.description')}
            </p>
            
            {/* Premium Benefits Visual */}
            <div className="bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl border border-gray-100">
              <div className="grid md:grid-cols-2 gap-8 items-center">
                <div>
                  <h3 className="text-2xl font-bold text-[#2D3436] mb-4">Premium Offline Features</h3>
                  <ul className="text-left space-y-3">
                    <li className="flex items-center text-[#636E72] leading-relaxed text-center">
                      <div className="w-2 h-2 bg-blue-500 rounded-full mr-3"></div>
                      Unlimited wordlist downloads
                    </li>
                    <li className="flex items-center text-[#636E72] leading-relaxed text-center">
                      <div className="w-2 h-2 bg-purple-500 rounded-full mr-3"></div>
                      All quiz modes available offline
                    </li>
                    <li className="flex items-center text-[#636E72] leading-relaxed text-center">
                      <div className="w-2 h-2 bg-green-500 rounded-full mr-3"></div>
                      High-quality audio downloads
                    </li>
                    <li className="flex items-center text-[#636E72] leading-relaxed text-center">
                      <div className="w-2 h-2 bg-orange-500 rounded-full mr-3"></div>
                      Unlimited storage capacity
                    </li>
                  </ul>
                </div>
                <div className="text-center">
                  <div className="w-32 h-32 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-4">
                    <svg className="w-16 h-16 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    </svg>
                  </div>
                  <div className="text-sm text-[#636E72]">Learn Anywhere, Anytime</div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </FeaturePageLayout>
    </PageLayout>
  );
};

export default OfflineSupportFeaturePage;
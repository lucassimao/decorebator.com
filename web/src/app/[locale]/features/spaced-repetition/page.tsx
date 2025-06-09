import React from 'react';
import { Metadata } from 'next';
import { useTranslations } from 'next-intl';
import FeaturePageLayout from '../../../../components/features/FeaturePageLayout';
import PageLayout from '../../../../components/layout/PageLayout';
import { generateFeatureMetadata, generateFeatureStructuredData } from '../../../../utils/featureMetadata';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  return generateFeatureMetadata({
    featureKey: 'spaced-repetition',
    locale,
    keywords: 'spaced repetition, Leitner system, vocabulary retention, memory improvement, learning science, 7-box system',
    ogImage: 'https://decorebator.com/og-spaced-repetition.jpg'
  });
}

const SpacedRepetitionFeaturePage: React.FC = () => {
  const t = useTranslations('featurePages.spacedRepetition');

  const structuredData = generateFeatureStructuredData('spacedRepetition', t);

  return (
    <PageLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <FeaturePageLayout featureKey="spacedRepetition">
        {/* Benefits Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-6">
                {t('benefits.title')}
              </h2>
            </div>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              {t.raw('benefits.items').map((benefit: string, index: number) => (
                <div key={index} className="bg-white dark:bg-gray-800 p-6 rounded-xl shadow-lg">
                  <div className="w-12 h-12 bg-gradient-to-r from-green-500 to-emerald-500 rounded-lg flex items-center justify-center mb-4">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <p className="text-gray-700 dark:text-gray-300">{benefit}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Leitner System Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white dark:bg-gray-800">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-6">
                {t('system.title')}
              </h2>
              <p className="text-xl text-gray-600 dark:text-gray-400 max-w-4xl mx-auto">
                {t('system.description')}
              </p>
            </div>
            
            <div className="overflow-x-auto">
              <table className="w-full bg-gray-50 dark:bg-gray-700 rounded-xl overflow-hidden">
                <thead className="bg-gradient-to-r from-green-500 to-emerald-500 text-white">
                  <tr>
                    <th className="px-6 py-4 text-left font-semibold">Box</th>
                    <th className="px-6 py-4 text-left font-semibold">Review Interval</th>
                    <th className="px-6 py-4 text-left font-semibold">Purpose</th>
                  </tr>
                </thead>
                <tbody>
                  {t.raw('system.boxes').map((box: { box: string; interval: string; purpose: string }, index: number) => (
                    <tr key={index} className={index % 2 === 0 ? 'bg-white dark:bg-gray-800' : 'bg-gray-50 dark:bg-gray-700'}>
                      <td className="px-6 py-4 font-medium text-gray-900 dark:text-white">{box.box}</td>
                      <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{box.interval}</td>
                      <td className="px-6 py-4 text-gray-600 dark:text-gray-300">{box.purpose}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </section>

        {/* Progression Rules Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-6">
                {t('progression.title')}
              </h2>
            </div>
            
            <div className="grid md:grid-cols-3 gap-8">
              <div className="bg-gradient-to-br from-green-50 to-emerald-50 dark:from-gray-800 dark:to-gray-700 p-8 rounded-xl text-center">
                <div className="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center mx-auto mb-6">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Correct Answer</h3>
                <p className="text-gray-600 dark:text-gray-400">{t('progression.correct')}</p>
              </div>

              <div className="bg-gradient-to-br from-red-50 to-pink-50 dark:from-gray-800 dark:to-gray-700 p-8 rounded-xl text-center">
                <div className="w-16 h-16 bg-red-500 rounded-full flex items-center justify-center mx-auto mb-6">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Incorrect Answer</h3>
                <p className="text-gray-600 dark:text-gray-400">{t('progression.incorrect')}</p>
              </div>

              <div className="bg-gradient-to-br from-yellow-50 to-amber-50 dark:from-gray-800 dark:to-gray-700 p-8 rounded-xl text-center">
                <div className="w-16 h-16 bg-yellow-500 rounded-full flex items-center justify-center mx-auto mb-6">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                  </svg>
                </div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">Mastered Words</h3>
                <p className="text-gray-600 dark:text-gray-400">{t('progression.mastered')}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Visual Representation */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-br from-green-50 to-emerald-100 dark:from-gray-800 dark:to-gray-900">
          <div className="max-w-7xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-12">
              85% Improvement in Retention
            </h2>
            <div className="bg-white dark:bg-gray-800 p-8 rounded-2xl shadow-lg inline-block">
              <div className="flex items-end space-x-4 h-40">
                <div className="flex flex-col items-center">
                  <div className="w-16 h-8 bg-red-400 rounded-t"></div>
                  <span className="mt-2 text-sm text-gray-600 dark:text-gray-400">Without</span>
                  <span className="text-xs text-gray-500 dark:text-gray-500">Spaced Repetition</span>
                </div>
                <div className="flex flex-col items-center">
                  <div className="w-16 h-32 bg-green-500 rounded-t"></div>
                  <span className="mt-2 text-sm text-gray-600 dark:text-gray-400">With Leitner</span>
                  <span className="text-xs text-gray-500 dark:text-gray-500">System</span>
                </div>
              </div>
              <p className="mt-4 text-gray-600 dark:text-gray-400">Long-term Retention Rate</p>
            </div>
          </div>
        </section>
      </FeaturePageLayout>
    </PageLayout>
  );
};

export default SpacedRepetitionFeaturePage;
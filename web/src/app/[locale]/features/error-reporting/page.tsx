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
    featureKey: 'errorReporting',
    locale,
    keywords: 'error reporting, quality control, AI content improvement, user feedback, community driven',
    ogImage: 'https://decorebator.com/og-error-reporting.jpg'
  });
}

const ErrorReportingFeaturePage: React.FC = () => {
  const t = useTranslations('featurePages.errorReporting');
  const structuredData = generateFeatureStructuredData('errorReporting', t);

  return (
    <PageLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <FeaturePageLayout featureKey="errorReporting">
        {/* Error Types Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                  {t('types.title')}
                </span>
              </h2>
            </div>
            
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C20.832 18.477 19.246 18 17.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Definition Errors</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('types.definition')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FFD700] to-yellow-500 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Image Issues</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('types.image')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Audio Problems</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('types.audio')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Example Issues</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('types.example')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-8 h-8 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Loading Problems</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('types.loading')}</p>
              </div>
            </div>
          </div>
        </section>

        {/* How It Works Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-[#2D3436] mb-6">
                {t('process.title')}
              </h2>
            </div>
            <div className="grid md:grid-cols-3 gap-8">
              {t.raw('process.steps').map((step: { title: string; description: string }, index: number) => (
                <div key={index} className="text-center group">
                  <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300 shadow-lg">
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

        {/* Benefits Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
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
                <div key={index} className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100">
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

        {/* Community Impact Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
          <div className="max-w-4xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold text-[#2D3436] mb-8">
              Your Reports Make a Difference
            </h2>
            <p className="text-xl text-[#636E72] leading-relaxed mb-12">
              Every error report you submit helps improve the learning experience for thousands of users worldwide. Together, we&apos;re building the most accurate AI-powered language learning platform.
            </p>
            
            {/* Impact Statistics */}
            <div className="bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl border border-gray-100">
              <div className="grid md:grid-cols-3 gap-8">
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#FF7B54] mb-2">25,000+</div>
                  <div className="text-[#636E72]">Reports submitted</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#FFD700] mb-2">98%</div>
                  <div className="text-[#636E72]">Issues resolved</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#4CAF50] mb-2">15%</div>
                  <div className="text-[#636E72]">Content quality improvement</div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </FeaturePageLayout>
    </PageLayout>
  );
};

export default ErrorReportingFeaturePage;
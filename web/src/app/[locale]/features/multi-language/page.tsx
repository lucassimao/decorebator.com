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
    featureKey: 'multiLanguage',
    locale,
    keywords: 'multi language support, native AI, 7 languages, cultural context, grammar rules',
    ogImage: 'https://decorebator.com/og-multi-language.jpg'
  });
}

const MultiLanguageFeaturePage: React.FC = () => {
  const t = useTranslations('featurePages.multiLanguage');
  const structuredData = generateFeatureStructuredData('multiLanguage', t);

  return (
    <PageLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <FeaturePageLayout featureKey="multiLanguage">
        {/* Supported Languages Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold mb-4">
                <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                  {t('languages.title')}
                </span>
              </h2>
            </div>
            
            <div className="flex flex-wrap justify-center gap-6 max-w-6xl mx-auto">
              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">EN</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">English</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.english')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FFD700] to-yellow-500 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">ES</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">Español</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.spanish')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">FR</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">Français</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.french')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">DE</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">Deutsch</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.german')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">IT</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">Italiano</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.italian')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FFD700] to-yellow-500 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">PT</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">Português</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.portuguese')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-6 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 text-center card-3d w-64">
                <div className="w-16 h-16 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-full flex items-center justify-center mx-auto mb-4 group-hover:scale-110 transition-transform duration-300">
                  <span className="text-2xl font-bold text-white">日</span>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-2 group-hover:text-[#FF7B54] transition-colors duration-300">日本語</h3>
                <p className="text-[#636E72] leading-relaxed text-sm">{t('languages.japanese')}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Language-Specific Features Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
          <div className="max-w-7xl mx-auto">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-4xl font-bold text-[#2D3436] mb-6">
                {t('features.title')}
              </h2>
            </div>
            
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Native Grammar</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.grammar')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Cultural Context</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.culture')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.536 8.464a5 5 0 010 7.072m2.828-9.9a9 9 0 010 12.728M5.586 15H4a1 1 0 01-1-1v-4a1 1 0 011-1h1.586l4.707-4.707C10.923 3.663 12 4.109 12 5v14c0 .891-1.077 1.337-1.707.707L5.586 15z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Optimized Pronunciation</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.pronunciation')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FFD700] to-yellow-500 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C20.832 18.477 19.246 18 17.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Authentic Examples</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.examples')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 8h10m0 0V6a2 2 0 00-2-2H9a2 2 0 00-2 2v2m10 0v10a2 2 0 01-2 2H9a2 2 0 01-2-2V8m10 0H7m0 0v10a2 2 0 002 2h6a2 2 0 002-2V8" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Language Nuances</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.nuances')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="w-16 h-16 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-2xl flex items-center justify-center mb-6 mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 text-center group-hover:text-[#FF7B54] transition-colors duration-300">Regional Variations</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.variations')}</p>
              </div>
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
      </FeaturePageLayout>
    </PageLayout>
  );
};

export default MultiLanguageFeaturePage;
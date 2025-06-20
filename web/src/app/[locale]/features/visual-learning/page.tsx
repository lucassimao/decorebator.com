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
    featureKey: 'visualLearning',
    locale,
    keywords: 'visual learning, AI images, memory association, visual memory, DALL-E, image generation',
    ogImage: 'https://decorebator.com/og-visual-learning.jpg'
  });
}

const VisualLearningFeaturePage: React.FC = () => {
  const t = useTranslations('featurePages.visualLearning');
  const structuredData = generateFeatureStructuredData('visualLearning', t);

  return (
    <PageLayout>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
      />
      <FeaturePageLayout featureKey="visualLearning">
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
                  <div className="bg-gradient-to-br from-[#FF7B54] to-orange-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                    <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <p className="text-[#636E72] leading-relaxed text-center">{benefit}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* AI Image Features Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-white reveal">
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
                <div className="bg-gradient-to-br from-[#FF7B54] to-orange-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">Custom Generation</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.generation')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="bg-gradient-to-br from-[#4CAF50] to-green-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m0 0a9 9 0 01-9-9m9 9v-9m0 9l3-3m-3 3l-3-3m9-5V3" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">Contextual Relevance</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.context')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="bg-gradient-to-br from-[#9C27B0] to-purple-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">Cultural Appropriateness</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.culture')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="bg-gradient-to-br from-[#FF7B54] to-orange-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">High Quality</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.quality')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="bg-gradient-to-br from-[#FFD700] to-yellow-500 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">Memory Association</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.association')}</p>
              </div>

              <div className="group bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500 border border-gray-100 card-3d">
                <div className="bg-gradient-to-br from-[#4CAF50] to-green-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center mx-auto group-hover:scale-110 transition-transform duration-300">
                  <svg className="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
                <h3 className="text-xl font-bold text-[#2D3436] mb-4 group-hover:text-[#FF7B54] transition-colors duration-300 text-center">Visual Variety</h3>
                <p className="text-[#636E72] leading-relaxed text-center">{t('features.variety')}</p>
              </div>
            </div>
          </div>
        </section>

        {/* Science Section */}
        <section className="py-20 px-4 sm:px-6 lg:px-8 reveal">
          <div className="max-w-4xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">
                {t('science.title')}
              </span>
            </h2>
            <p className="text-xl text-[#636E72] leading-relaxed mb-8">
              {t('science.description')}
            </p>
            
            {/* Visual Statistics */}
            <div className="mt-16 bg-white/80 backdrop-blur p-8 rounded-3xl shadow-xl border border-gray-100">
              <div className="grid md:grid-cols-3 gap-8">
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#FF7B54] mb-2">6x</div>
                  <div className="text-[#636E72]">Stronger than text alone</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#FFD700] mb-2">65%</div>
                  <div className="text-[#636E72]">Better retention after 3 days</div>
                </div>
                <div className="text-center">
                  <div className="text-4xl font-bold text-[#FF7B54] mb-2">90%</div>
                  <div className="text-[#636E72]">Information transmitted visually</div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </FeaturePageLayout>
    </PageLayout>
  );
};

export default VisualLearningFeaturePage;
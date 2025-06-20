import React from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import FooterSection from '../home/FooterSection';

interface FeaturePageLayoutProps {
  featureKey: string;
  children?: React.ReactNode;
}

const FeaturePageLayout: React.FC<FeaturePageLayoutProps> = ({ featureKey, children }) => {
  const t = useTranslations(`featurePages.${featureKey}`);
  const tCommon = useTranslations('common');

  return (
    <div className="min-h-screen bg-[#FDF6E3] relative overflow-hidden">
      {/* Animated Background Elements */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-20 left-10 w-32 h-32 bg-orange-300 rounded-full opacity-3 blur-3xl float-animation"></div>
        <div className="absolute bottom-20 right-10 w-40 h-40 bg-yellow-300 rounded-full opacity-3 blur-3xl float-animation" style={{animationDelay: '3s'}}></div>
        <div className="absolute top-1/2 left-1/3 w-24 h-24 bg-purple-300 rounded-full opacity-2 blur-2xl float-animation" style={{animationDelay: '6s'}}></div>
      </div>

      {/* Hero Section */}
      <section className="relative pt-24 pb-20 px-4 sm:px-6 lg:px-8 z-10">
        <div className="max-w-7xl mx-auto">
          {/* Breadcrumb */}
          <nav className="mb-8" aria-label="Breadcrumb">
            <ol className="flex items-center space-x-2 text-sm">
              <li>
                <Link href="/" className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300">
                  Home
                </Link>
              </li>
              <li className="text-[#636E72]">/</li>
              <li>
                <Link href="/#features" className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300">
                  Features
                </Link>
              </li>
              <li className="text-[#636E72]">/</li>
              <li className="text-[#2D3436] font-medium">
                {t('title')}
              </li>
            </ol>
          </nav>

          <div className="text-center slide-in-up">
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-[#2D3436] mb-6">
              {t('hero.title')}
            </h1>
            <p className="text-xl md:text-2xl text-[#636E72] max-w-4xl mx-auto mb-8 leading-relaxed">
              {t('hero.subtitle')}
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link
                href="/signup?plan=free"
                className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 flex items-center justify-center"
              >
                <span>{tCommon('getStarted')}</span>
                <i className="fas fa-arrow-right ml-2 group-hover:translate-x-2 transition-transform"></i>
              </Link>
              <Link
                href="/#features"
                className="group bg-white/80 backdrop-blur px-8 py-4 rounded-full font-semibold text-lg border-2 border-gray-200 hover:border-[#FF7B54] transition-all duration-300 flex items-center justify-center text-[#2D3436]"
              >
                <i className="fas fa-grid-3x3 mr-2 text-[#FF7B54] group-hover:scale-110 transition-transform"></i>
                View All Features
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Custom content passed as children */}
      <div className="relative z-10">
        {children}
      </div>

      {/* CTA Section */}
      <section className="relative py-20 px-4 sm:px-6 lg:px-8 bg-gradient-to-r from-[#FF7B54] to-orange-600 z-10">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-6">
            Ready to Experience This Feature?
          </h2>
          <p className="text-xl text-orange-100 mb-8">
            Join thousands of learners already mastering vocabulary with Decorebator
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/signup?plan=free"
              className="group bg-white text-[#FF7B54] px-8 py-4 rounded-full font-semibold text-lg hover:shadow-lg transform hover:scale-105 transition-all duration-300 flex items-center justify-center"
            >
              <span>Start Free Trial</span>
              <i className="fas fa-rocket ml-2 group-hover:translate-x-1 transition-transform"></i>
            </Link>
            <Link
              href="/#pricing"
              className="group border-2 border-white text-white px-8 py-4 rounded-full font-semibold text-lg hover:bg-white hover:text-[#FF7B54] transition-all duration-300 flex items-center justify-center"
            >
              <i className="fas fa-tag mr-2 group-hover:scale-110 transition-transform"></i>
              View Pricing
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <FooterSection />
    </div>
  );
};

export default FeaturePageLayout;
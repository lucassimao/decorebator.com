'use client';

import React, { useState, useEffect } from 'react';
import { useTranslations, useLocale } from 'next-intl';
import Link from 'next/link';
import LanguageSwitcher from '../common/LanguageSwitcher';

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
  const [isFeaturesDropdownOpen, setIsFeaturesDropdownOpen] = useState(false);
  const t = useTranslations('navigation');
  const locale = useLocale();

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 0);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const featureLinks = [
    { title: 'AI Content Generation', href: `/features/ai-content` },
    { title: 'Spaced Repetition', href: `/features/spaced-repetition` },
    { title: 'Quiz Modes', href: `/features/quiz-modes` },
    { title: 'Visual Learning', href: `/features/visual-learning` },
    { title: 'Audio Learning', href: `/features/audio-learning` },
    { title: 'Analytics', href: `/features/analytics` },
    { title: 'Multi-Language', href: `/features/multi-language` },
    { title: 'Flashcards', href: `/features/flashcards` },
    { title: 'Error Reporting', href: `/features/error-reporting` },
    { title: 'Offline Support', href: `/features/offline-support` }
  ];

  return (
    <header className={`fixed top-0 left-0 right-0 z-50 glass transition-all duration-300 ${isScrolled ? 'backdrop-blur-lg' : ''}`}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <Link href="/" className="flex items-center space-x-3 group cursor-pointer">
            <div className="w-10 h-10 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-xl flex items-center justify-center transform group-hover:rotate-12 transition-transform duration-300">
              <i className="fas fa-book-open text-white text-lg"></i>
            </div>
            <span className="text-2xl font-bold bg-gradient-to-r from-[#FF7B54] to-orange-600 bg-clip-text text-transparent">
              Decorebator
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center space-x-6">
            <a href={`/${locale}`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('home')}
            </a>
            
            {/* Features Dropdown */}
            <div 
              className="relative"
              onMouseEnter={() => setIsFeaturesDropdownOpen(true)}
              onMouseLeave={() => setIsFeaturesDropdownOpen(false)}
            >
              <button className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium flex items-center">
                {t('features')}
                <i className={`fas fa-chevron-down ml-1 text-xs transition-transform duration-200 ${isFeaturesDropdownOpen ? 'rotate-180' : ''}`}></i>
              </button>
              
              {isFeaturesDropdownOpen && (
                <div className="absolute top-full left-0 mt-2 w-64 bg-white/95 backdrop-blur-lg shadow-xl rounded-xl border border-gray-100 py-2 z-50">
                  <Link href={`/${locale}/#features`} className="block px-4 py-2 text-sm text-[#636E72] hover:text-[#FF7B54] hover:bg-orange-50 transition-colors">
                    View All Features
                  </Link>
                  <div className="border-t border-gray-100 my-2"></div>
                  {featureLinks.map((feature) => (
                    <Link
                      key={feature.href}
                      href={`/${locale}${feature.href}`}
                      className="block px-4 py-2 text-sm text-[#636E72] hover:text-[#FF7B54] hover:bg-orange-50 transition-colors"
                    >
                      {feature.title}
                    </Link>
                  ))}
                </div>
              )}
            </div>
            
            <a href={`/${locale}/#how-it-works`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('howItWorks')}
            </a>
            <a href={`/${locale}/#pricing`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('pricing')}
            </a>
            <a href={`/${locale}/#contact`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('contact')}
            </a>
            <LanguageSwitcher />
            <Link href={`/${locale}/signup?plan=free`} className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-2.5 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300 inline-block">
              {t('getStartedFree')}
            </Link>
          </nav>

          {/* Mobile Menu Button */}
          <button className="md:hidden p-2" onClick={toggleMobileMenu}>
            <i className="fas fa-bars text-2xl text-[#2D3436]"></i>
          </button>
        </div>

        {/* Mobile Navigation */}
        <div className={`${isMobileMenuOpen ? 'block' : 'hidden'} md:hidden absolute top-16 left-0 right-0 bg-white/95 backdrop-blur-lg shadow-xl rounded-b-2xl max-h-96 overflow-y-auto`}>
          <nav className="flex flex-col p-6 space-y-4">
            <a href={`/${locale}`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('home')}
            </a>
            
            {/* Features Section */}
            <div>
              <a href={`/${locale}/#features`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium block mb-2" onClick={() => setIsMobileMenuOpen(false)}>
                {t('features')}
              </a>
              <div className="ml-4 space-y-2">
                {featureLinks.slice(0, 5).map((feature) => (
                  <Link
                    key={feature.href}
                    href={`/${locale}${feature.href}`}
                    className="block text-sm text-[#636E72] hover:text-[#FF7B54] transition-colors"
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    {feature.title}
                  </Link>
                ))}
                <details className="text-sm">
                  <summary className="text-[#636E72] hover:text-[#FF7B54] cursor-pointer">More Features...</summary>
                  <div className="ml-2 mt-2 space-y-2">
                    {featureLinks.slice(5).map((feature) => (
                      <Link
                        key={feature.href}
                        href={`/${locale}${feature.href}`}
                        className="block text-sm text-[#636E72] hover:text-[#FF7B54] transition-colors"
                        onClick={() => setIsMobileMenuOpen(false)}
                      >
                        {feature.title}
                      </Link>
                    ))}
                  </div>
                </details>
              </div>
            </div>
            
            <a href={`/${locale}/#how-it-works`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('howItWorks')}
            </a>
            <a href={`/${locale}/#pricing`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('pricing')}
            </a>
            <a href={`/${locale}/#contact`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('contact')}
            </a>
            <div className="pt-2 border-t border-gray-200">
              <LanguageSwitcher />
            </div>
            <Link href={`/${locale}/signup?plan=free`} className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-full font-semibold inline-block text-center" onClick={() => setIsMobileMenuOpen(false)}>
              {t('getStartedFree')}
            </Link>
          </nav>
        </div>
      </div>
    </header>
  );
};

export default Header;
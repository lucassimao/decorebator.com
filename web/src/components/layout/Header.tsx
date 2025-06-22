'use client';

import React, { useState, useEffect } from 'react';
import { useTranslations, useLocale } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';
import LanguageSwitcher from '../common/LanguageSwitcher';
import DownloadAppButton from '../common/DownloadAppButton';

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isScrolled, setIsScrolled] = useState(false);
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


  return (
    <header className={`fixed top-0 left-0 right-0 z-50 glass transition-all duration-300 ${isScrolled ? 'backdrop-blur-lg' : ''}`}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <Link href="/" className="flex items-center space-x-3 group cursor-pointer">
            <div className="w-10 h-10 flex items-center justify-center transform group-hover:rotate-12 transition-transform duration-300">
              <Image
                src="/logo.png"
                alt="Decorebator Logo"
                width={40}
                height={40}
                className="w-10 h-10"
                priority
              />
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
            <a href={`/${locale}/#features`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('features')}
            </a>
            <a href={`/${locale}/#how-it-works`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('howItWorks')}
            </a>
            <a href={`/${locale}/#pricing`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              {t('pricing')}
            </a>
            <a href={`/${locale}/#faq`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300 font-medium">
              FAQ
            </a>
            <LanguageSwitcher />
            <DownloadAppButton className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-2.5 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300 inline-block">
              Download App
            </DownloadAppButton>
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
            <a href={`/${locale}/#features`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('features')}
            </a>
            <a href={`/${locale}/#how-it-works`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('howItWorks')}
            </a>
            <a href={`/${locale}/#pricing`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              {t('pricing')}
            </a>
            <a href={`/${locale}/#faq`} className="text-[#636E72] hover:text-[#FF7B54] transition-colors font-medium" onClick={() => setIsMobileMenuOpen(false)}>
              FAQ
            </a>
            <div className="pt-2 border-t border-gray-200">
              <LanguageSwitcher />
            </div>
            <DownloadAppButton 
              className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-full font-semibold inline-block text-center" 
              onClick={() => setIsMobileMenuOpen(false)}
            >
              Download App
            </DownloadAppButton>
          </nav>
        </div>
      </div>
    </header>
  );
};

export default Header;
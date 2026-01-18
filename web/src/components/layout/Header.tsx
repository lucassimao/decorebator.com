'use client'

import React, { useState, useEffect } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import Link from 'next/link'
import Image from 'next/image'
import { Menu, X } from 'lucide-react'
import LanguageSwitcher from '../common/LanguageSwitcher'
import DownloadAppButton from '../common/DownloadAppButton'

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isScrolled, setIsScrolled] = useState(false)
  const t = useTranslations('navigation')
  const locale = useLocale()

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 20)
    }

    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen)
  }

  return (
    <header
      className={`fixed top-0 right-0 left-0 z-50 transition-all duration-300 ${
        isScrolled
          ? 'border-b border-slate-200 bg-white/80 shadow-sm backdrop-blur-xl'
          : 'bg-white/60 backdrop-blur-md'
      }`}
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between sm:h-20">
          {/* Logo */}
          <Link
            href={`/${locale}`}
            className="group flex cursor-pointer items-center space-x-2.5 transition-transform duration-200 hover:scale-[1.02]"
          >
            <div className="flex h-9 w-9 items-center justify-center">
              <Image
                src="/logo.png"
                alt="Decorebator Logo"
                width={36}
                height={36}
                className="h-9 w-9 transition-transform duration-200 group-hover:rotate-6"
                priority
              />
            </div>
            <span className="from-primary-500 to-primary-600 bg-gradient-to-r bg-clip-text text-xl font-bold text-transparent sm:text-2xl">
              Decorebator
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden items-center space-x-8 md:flex">
            <a
              href={`/${locale}/#features`}
              className="hover:text-primary-500 focus-visible:ring-primary-500 text-sm font-medium text-slate-600 transition-colors duration-200 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              {t('features')}
            </a>
            <a
              href={`/${locale}/#how-it-works`}
              className="hover:text-primary-500 focus-visible:ring-primary-500 text-sm font-medium text-slate-600 transition-colors duration-200 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              {t('howItWorks')}
            </a>
            <a
              href={`/${locale}/#pricing`}
              className="hover:text-primary-500 focus-visible:ring-primary-500 text-sm font-medium text-slate-600 transition-colors duration-200 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              {t('pricing')}
            </a>
            <a
              href={`/${locale}/#faq`}
              className="hover:text-primary-500 focus-visible:ring-primary-500 text-sm font-medium text-slate-600 transition-colors duration-200 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              {t('faq')}
            </a>
            <LanguageSwitcher />
            <DownloadAppButton className="bg-primary-500 hover:bg-primary-600 rounded-full px-5 py-2 text-sm font-semibold text-white shadow-sm transition-all duration-200 hover:shadow-md">
              {t('getStartedFree')}
            </DownloadAppButton>
          </nav>

          {/* Mobile Menu Button */}
          <button
            className="focus-visible:ring-primary-500 rounded-lg p-2 text-slate-600 transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none md:hidden"
            onClick={toggleMobileMenu}
            aria-label="Toggle menu"
            aria-expanded={isMobileMenuOpen}
            aria-controls="mobile-navigation"
          >
            {isMobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
          </button>
        </div>

        {/* Mobile Navigation */}
        {isMobileMenuOpen && (
          <div id="mobile-navigation" className="border-t border-slate-200 bg-white md:hidden">
            <nav className="flex flex-col space-y-1 py-4">
              <a
                href={`/${locale}`}
                className="hover:text-primary-500 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {t('home')}
              </a>
              <a
                href={`/${locale}/#features`}
                className="hover:text-primary-500 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {t('features')}
              </a>
              <a
                href={`/${locale}/#how-it-works`}
                className="hover:text-primary-500 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {t('howItWorks')}
              </a>
              <a
                href={`/${locale}/#pricing`}
                className="hover:text-primary-500 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {t('pricing')}
              </a>
              <a
                href={`/${locale}/#faq`}
                className="hover:text-primary-500 rounded-lg px-4 py-2.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50"
                onClick={() => setIsMobileMenuOpen(false)}
              >
                {t('faq')}
              </a>
              <div className="border-t border-slate-200 px-4 pt-3">
                <LanguageSwitcher />
              </div>
              <div className="px-4 pt-2">
                <DownloadAppButton
                  className="bg-primary-500 w-full rounded-full px-5 py-3 text-center text-sm font-semibold text-white"
                  onClick={() => setIsMobileMenuOpen(false)}
                >
                  {t('getStartedFree')}
                </DownloadAppButton>
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  )
}

export default Header

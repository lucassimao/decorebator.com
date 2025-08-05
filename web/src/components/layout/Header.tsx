'use client'

import React, { useState, useEffect } from 'react'
import { useTranslations, useLocale } from 'next-intl'
import Link from 'next/link'
import Image from 'next/image'
import LanguageSwitcher from '../common/LanguageSwitcher'
import DownloadAppButton from '../common/DownloadAppButton'

const Header: React.FC = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const [isScrolled, setIsScrolled] = useState(false)
  const t = useTranslations('navigation')
  const locale = useLocale()

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 0)
    }

    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen)
  }

  return (
    <header
      className={`glass fixed top-0 right-0 left-0 z-50 transition-all duration-300 ${isScrolled ? 'backdrop-blur-lg' : ''}`}
    >
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <Link href="/" className="group flex cursor-pointer items-center space-x-3">
            <div className="flex h-10 w-10 transform items-center justify-center transition-transform duration-300 group-hover:rotate-12">
              <Image
                src="/logo.png"
                alt="Decorebator Logo"
                width={40}
                height={40}
                className="h-10 w-10"
                priority
              />
            </div>
            <span className="bg-gradient-to-r from-[#FF7B54] to-orange-600 bg-clip-text text-2xl font-bold text-transparent">
              Decorebator
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden items-center space-x-6 md:flex">
            <a
              href={`/${locale}`}
              className="font-medium text-[#636E72] transition-colors duration-300 hover:text-[#FF7B54]"
            >
              {t('home')}
            </a>
            <a
              href={`/${locale}/#features`}
              className="font-medium text-[#636E72] transition-colors duration-300 hover:text-[#FF7B54]"
            >
              {t('features')}
            </a>
            <a
              href={`/${locale}/#how-it-works`}
              className="font-medium text-[#636E72] transition-colors duration-300 hover:text-[#FF7B54]"
            >
              {t('howItWorks')}
            </a>
            <a
              href={`/${locale}/#pricing`}
              className="font-medium text-[#636E72] transition-colors duration-300 hover:text-[#FF7B54]"
            >
              {t('pricing')}
            </a>
            <a
              href={`/${locale}/#faq`}
              className="font-medium text-[#636E72] transition-colors duration-300 hover:text-[#FF7B54]"
            >
              {t('faq')}
            </a>
            <LanguageSwitcher />
            <DownloadAppButton className="inline-block transform rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-2.5 font-semibold text-white transition-all duration-300 hover:scale-105 hover:shadow-lg">
              {t('getStartedFree')}
            </DownloadAppButton>
          </nav>

          {/* Mobile Menu Button */}
          <button className="p-2 md:hidden" onClick={toggleMobileMenu}>
            <i className="fas fa-bars text-2xl text-[#2D3436]"></i>
          </button>
        </div>

        {/* Mobile Navigation */}
        <div
          className={`${isMobileMenuOpen ? 'block' : 'hidden'} absolute top-16 right-0 left-0 rounded-b-2xl bg-white/95 shadow-xl backdrop-blur-lg md:hidden`}
        >
          <nav className="flex flex-col space-y-4 p-6">
            <a
              href={`/${locale}`}
              className="font-medium text-[#636E72] transition-colors hover:text-[#FF7B54]"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('home')}
            </a>
            <a
              href={`/${locale}/#features`}
              className="font-medium text-[#636E72] transition-colors hover:text-[#FF7B54]"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('features')}
            </a>
            <a
              href={`/${locale}/#how-it-works`}
              className="font-medium text-[#636E72] transition-colors hover:text-[#FF7B54]"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('howItWorks')}
            </a>
            <a
              href={`/${locale}/#pricing`}
              className="font-medium text-[#636E72] transition-colors hover:text-[#FF7B54]"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('pricing')}
            </a>
            <a
              href={`/${locale}/#faq`}
              className="font-medium text-[#636E72] transition-colors hover:text-[#FF7B54]"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('faq')}
            </a>
            <div className="border-t border-gray-200 pt-2">
              <LanguageSwitcher />
            </div>
            <DownloadAppButton
              className="inline-block rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 text-center font-semibold text-white"
              onClick={() => setIsMobileMenuOpen(false)}
            >
              {t('getStartedFree')}
            </DownloadAppButton>
          </nav>
        </div>
      </div>
    </header>
  )
}

export default Header

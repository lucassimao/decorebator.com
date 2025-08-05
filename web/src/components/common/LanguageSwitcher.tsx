'use client'

import React, { useState, useEffect } from 'react'
import { useLocale } from 'next-intl'
import { useRouter, usePathname } from 'next/navigation'

const languages = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
  { code: 'it', name: 'Italiano', flag: '🇮🇹' },
  { code: 'pt', name: 'Português', flag: '🇵🇹' },
  { code: 'ja', name: '日本語', flag: '🇯🇵' },
]

const LanguageSwitcher: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false)
  const [isMobile, setIsMobile] = useState(false)
  const locale = useLocale()
  const router = useRouter()
  const pathname = usePathname()

  const currentLanguage = languages.find((lang) => lang.code === locale) || languages[0]

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }

    checkMobile()
    window.addEventListener('resize', checkMobile)
    return () => window.removeEventListener('resize', checkMobile)
  }, [])

  const handleLanguageChange = (newLocale: string) => {
    // Replace the current locale in the pathname with the new one
    const newPathname = pathname.replace(`/${locale}`, `/${newLocale}`)
    router.push(newPathname)
    setIsOpen(false)
  }

  // Mobile: show all languages inline
  if (isMobile) {
    return (
      <div className="space-y-2">
        <div className="mb-3 text-sm font-medium text-gray-600">Select Language:</div>
        <div className="grid grid-cols-2 gap-2">
          {languages.map((language) => (
            <button
              key={language.code}
              onClick={() => handleLanguageChange(language.code)}
              className={`flex items-center space-x-2 rounded-lg px-3 py-2 text-left text-sm transition-colors duration-200 ${
                language.code === locale
                  ? 'border border-[#FF7B54] bg-orange-50 text-[#FF7B54]'
                  : 'bg-gray-50 text-gray-700 hover:bg-gray-100'
              }`}
            >
              <span className="text-base">{language.flag}</span>
              <span className="truncate font-medium">{language.name}</span>
              {language.code === locale && <i className="fas fa-check ml-auto text-[#FF7B54]"></i>}
            </button>
          ))}
        </div>
      </div>
    )
  }

  // Desktop: dropdown behavior
  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center space-x-2 rounded-lg bg-white/10 px-3 py-2 backdrop-blur transition-colors duration-200 hover:bg-white/20"
        aria-label="Select language"
      >
        <span className="text-lg">{currentLanguage.flag}</span>
        <span className="hidden text-sm font-medium text-white sm:block">
          {currentLanguage.name}
        </span>
        <i
          className={`fas fa-chevron-down text-xs transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
        ></i>
      </button>

      {isOpen && (
        <>
          {/* Backdrop */}
          <div className="fixed inset-0 z-[60]" onClick={() => setIsOpen(false)} />

          {/* Dropdown */}
          <div className="absolute top-full right-0 z-[70] mt-2 min-w-48 overflow-hidden rounded-xl border border-gray-200 bg-white shadow-2xl">
            {languages.map((language) => (
              <button
                key={language.code}
                onClick={() => handleLanguageChange(language.code)}
                className={`flex w-full items-center space-x-3 px-4 py-3 text-left transition-colors duration-200 hover:bg-gray-50 ${
                  language.code === locale ? 'bg-orange-50 text-[#FF7B54]' : 'text-gray-700'
                }`}
              >
                <span className="text-lg">{language.flag}</span>
                <span className="font-medium">{language.name}</span>
                {language.code === locale && (
                  <i className="fas fa-check ml-auto text-[#FF7B54]"></i>
                )}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

export default LanguageSwitcher

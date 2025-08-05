'use client'

import React from 'react'
import { useTranslations } from 'next-intl'

interface AppStoreAlertProps {
  isOpen: boolean
  onClose: () => void
}

const AppStoreAlert: React.FC<AppStoreAlertProps> = ({ isOpen, onClose }) => {
  const t = useTranslations('common')

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className="animate-scale-in relative z-10 w-full max-w-sm rounded-2xl bg-white p-8 shadow-2xl">
        <div className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600">
            <i className="fas fa-rocket text-2xl text-white"></i>
          </div>
          <h3 className="mb-3 text-2xl font-bold text-[#2D3436]">{t('appStorePending.title')}</h3>
          <p className="mb-6 text-[#636E72]">{t('appStorePending.message')}</p>
          <button
            onClick={onClose}
            className="rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 px-8 py-3 font-semibold text-white transition-all duration-300 hover:shadow-lg"
          >
            {t('appStorePending.okButton')}
          </button>
        </div>
      </div>
    </div>
  )
}

export default AppStoreAlert

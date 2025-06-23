'use client';

import React from 'react';
import { useTranslations } from 'next-intl';

interface AppStoreAlertProps {
  isOpen: boolean;
  onClose: () => void;
}

const AppStoreAlert: React.FC<AppStoreAlertProps> = ({ isOpen, onClose }) => {
  const t = useTranslations('common');

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="fixed inset-0 bg-black/50" onClick={onClose} />
      <div className="bg-white rounded-2xl p-8 max-w-sm w-full relative z-10 shadow-2xl animate-scale-in">
        <div className="text-center">
          <div className="w-16 h-16 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <i className="fas fa-rocket text-white text-2xl"></i>
          </div>
          <h3 className="text-2xl font-bold text-[#2D3436] mb-3">
            {t('appStorePending.title')}
          </h3>
          <p className="text-[#636E72] mb-6">
            {t('appStorePending.message')}
          </p>
          <button
            onClick={onClose}
            className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-8 py-3 rounded-full font-semibold hover:shadow-lg transition-all duration-300"
          >
            {t('appStorePending.okButton')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default AppStoreAlert;
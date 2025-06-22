'use client';

import React, { createContext, useContext, useState, useCallback } from 'react';
import AppStoreAlert from './AppStoreAlert';

interface AppStoreModalContextType {
  showModal: () => void;
}

const AppStoreModalContext = createContext<AppStoreModalContextType | undefined>(undefined);

export const useAppStoreModal = () => {
  const context = useContext(AppStoreModalContext);
  if (!context) {
    throw new Error('useAppStoreModal must be used within AppStoreModalProvider');
  }
  return context;
};

interface AppStoreModalProviderProps {
  children: React.ReactNode;
}

export const AppStoreModalProvider: React.FC<AppStoreModalProviderProps> = ({ children }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const showModal = useCallback(() => {
    setIsModalOpen(true);
  }, []);

  const closeModal = useCallback(() => {
    setIsModalOpen(false);
  }, []);

  return (
    <AppStoreModalContext.Provider value={{ showModal }}>
      {children}
      <AppStoreAlert isOpen={isModalOpen} onClose={closeModal} />
    </AppStoreModalContext.Provider>
  );
};
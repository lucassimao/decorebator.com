'use client';

import React from 'react';
import { appStoreConfig } from '@/config/appStoreConfig';
import { useAppStoreModal } from './AppStoreModalProvider';

interface DownloadAppButtonProps {
  className?: string;
  children?: React.ReactNode;
  onClick?: () => void;
}

const DownloadAppButton: React.FC<DownloadAppButtonProps> = ({ 
  className = '', 
  children = 'Download App',
  onClick
}) => {
  const { showModal } = useAppStoreModal();

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault();
    
    // If custom onClick is provided, call it
    if (onClick) {
      onClick();
      return;
    }

    // Check if either store URL exists
    if (appStoreConfig.appStoreUrl || appStoreConfig.playStoreUrl) {
      // If we have at least one URL, navigate to download section
      // This assumes there's a download section on the page
      const downloadSection = document.getElementById('download');
      if (downloadSection) {
        downloadSection.scrollIntoView({ behavior: 'smooth' });
      }
    } else if (appStoreConfig.showPendingModal) {
      // If no URLs and modal is enabled, show the modal
      showModal();
    }
  };

  return (
    <a 
      href="#download" 
      onClick={handleClick}
      className={className}
    >
      {children}
    </a>
  );
};

export default DownloadAppButton;
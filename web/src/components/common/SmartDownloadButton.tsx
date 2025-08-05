'use client'

import React from 'react'
import { appStoreConfig } from '@/config/appStoreConfig'
import { useAppStoreModal } from './AppStoreModalProvider'

interface SmartDownloadButtonProps {
  className?: string
  children?: React.ReactNode
  onClick?: () => void
  size?: 'small' | 'medium' | 'large'
}

const SmartDownloadButton: React.FC<SmartDownloadButtonProps> = ({
  className = '',
  children = 'Download App',
  onClick,
  size = 'medium',
}) => {
  const { showModal } = useAppStoreModal()

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault()

    // If custom onClick is provided, call it first
    if (onClick) {
      onClick()
    }

    // Check if either store URL exists
    if (appStoreConfig.appStoreUrl || appStoreConfig.playStoreUrl) {
      // If we have at least one URL, navigate to download section
      // This assumes there's a download section on the page with app store buttons
      const downloadSection = document.getElementById('download')
      if (downloadSection) {
        downloadSection.scrollIntoView({ behavior: 'smooth' })
      } else {
        // Fallback: redirect to preferred store if available
        const preferredUrl = appStoreConfig.appStoreUrl || appStoreConfig.playStoreUrl
        if (preferredUrl) {
          window.open(preferredUrl, '_blank', 'noopener,noreferrer')
        }
      }
    } else if (appStoreConfig.showPendingModal) {
      // If no URLs and modal is enabled, show the modal
      showModal()
    }
    // If showPendingModal is false and no URLs, do nothing
  }

  // Size-based styling
  const sizeClasses = {
    small: 'px-4 py-2 text-sm',
    medium: 'px-6 py-3 text-base',
    large: 'px-8 py-4 text-lg',
  }

  const defaultClasses = `inline-flex items-center space-x-2 bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300 ${sizeClasses[size]}`

  return (
    <button onClick={handleClick} className={className || defaultClasses}>
      {typeof children === 'string' ? (
        <>
          <span>{children}</span>
          <i className="fas fa-download"></i>
        </>
      ) : (
        children
      )}
    </button>
  )
}

export default SmartDownloadButton

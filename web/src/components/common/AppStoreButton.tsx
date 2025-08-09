'use client'

import React from 'react'
import Image from 'next/image'
import { appStoreConfig } from '@/config/appStoreConfig'
import { useAppStoreModal } from './AppStoreModalProvider'

interface AppStoreButtonProps {
  store: 'apple' | 'google'
  className?: string
  size?: 'small' | 'medium' | 'large'
}

const AppStoreButton: React.FC<AppStoreButtonProps> = ({ store, className = '', size = 'medium' }) => {
  const { showModal } = useAppStoreModal()

  const handleClick = () => {
    const url = store === 'apple' ? appStoreConfig.appStoreUrl : appStoreConfig.playStoreUrl

    if (url) {
      // If URL exists, open it in a new tab
      window.open(url, '_blank', 'noopener,noreferrer')
    } else if (appStoreConfig.showPendingModal) {
      // If no URL and modal is enabled, show the modal
      showModal()
    }
    // If showPendingModal is false and no URL, do nothing
  }

  const imageProps =
    store === 'apple'
      ? {
          src: 'https://upload.wikimedia.org/wikipedia/commons/3/3c/Download_on_the_App_Store_Badge.svg',
          alt: 'Download on App Store',
        }
      : {
          src: 'https://upload.wikimedia.org/wikipedia/commons/7/78/Google_Play_Store_badge_EN.svg',
          alt: 'Get it on Google Play',
        }

  const sizeClass = size === 'small' ? 'h-10' : size === 'large' ? 'h-16' : 'h-14'

  return (
    <button
      onClick={handleClick}
      className={`transform transition-transform duration-300 hover:scale-105 ${className}`}
      aria-label={imageProps.alt}
    >
      <Image src={imageProps.src} alt={imageProps.alt} width={168} height={56} className={`${sizeClass} w-auto`} />
    </button>
  )
}

export default AppStoreButton

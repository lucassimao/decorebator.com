'use client'

import React from 'react'
import Header from './Header'
import BackgroundElements from './BackgroundElements'
import FooterSection from '../home/FooterSection'

interface PageLayoutProps {
  children: React.ReactNode
  className?: string
}

const PageLayout: React.FC<PageLayoutProps> = ({ children, className = '' }) => {
  return (
    <div className={`min-h-screen overflow-x-hidden bg-[#FDF6E3] text-[#2D3436] ${className}`}>
      <BackgroundElements />
      <Header />
      <main className="relative z-10">{children}</main>
      <FooterSection />
    </div>
  )
}

export default PageLayout

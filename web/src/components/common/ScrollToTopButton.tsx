'use client'

import React, { useState, useEffect } from 'react'

const ScrollToTopButton: React.FC = () => {
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    const toggleVisibility = () => {
      // Show button when user scrolls down 300px
      if (window.scrollY > 300) {
        setIsVisible(true)
      } else {
        setIsVisible(false)
      }
    }

    window.addEventListener('scroll', toggleVisibility)

    return () => {
      window.removeEventListener('scroll', toggleVisibility)
    }
  }, [])

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth',
    })
  }

  return (
    <button
      onClick={scrollToTop}
      className={`fixed right-6 bottom-6 z-40 flex h-12 w-12 transform items-center justify-center rounded-full bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white shadow-lg transition-all duration-300 hover:scale-110 hover:shadow-xl md:hidden ${isVisible ? 'translate-y-0 opacity-100' : 'pointer-events-none translate-y-16 opacity-0'} `}
      aria-label="Scroll to top"
    >
      <i className="fas fa-chevron-up text-sm"></i>
    </button>
  )
}

export default ScrollToTopButton

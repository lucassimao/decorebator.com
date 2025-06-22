'use client';

import React, { useState, useEffect } from 'react';

const ScrollToTopButton: React.FC = () => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const toggleVisibility = () => {
      // Show button when user scrolls down 300px
      if (window.scrollY > 300) {
        setIsVisible(true);
      } else {
        setIsVisible(false);
      }
    };

    window.addEventListener('scroll', toggleVisibility);

    return () => {
      window.removeEventListener('scroll', toggleVisibility);
    };
  }, []);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth'
    });
  };

  return (
    <button
      onClick={scrollToTop}
      className={`
        fixed bottom-6 right-6 z-40
        w-12 h-12 
        bg-gradient-to-r from-[#FF7B54] to-orange-600
        text-white
        rounded-full
        shadow-lg hover:shadow-xl
        flex items-center justify-center
        transform transition-all duration-300
        hover:scale-110
        md:hidden
        ${isVisible ? 'translate-y-0 opacity-100' : 'translate-y-16 opacity-0 pointer-events-none'}
      `}
      aria-label="Scroll to top"
    >
      <i className="fas fa-chevron-up text-sm"></i>
    </button>
  );
};

export default ScrollToTopButton;
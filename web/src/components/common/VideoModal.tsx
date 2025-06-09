'use client';

import React, { useEffect } from 'react';

interface VideoModalProps {
  isOpen: boolean;
  onClose: () => void;
  videoId?: string; // YouTube video ID
  title?: string;
}

const VideoModal: React.FC<VideoModalProps> = ({ 
  isOpen, 
  onClose, 
  videoId = 'dQw4w9WgXcQ', // Default demo video ID
  title = 'Decorebator Demo'
}) => {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity duration-300"
        onClick={onClose}
      />
      
      {/* Modal Content */}
      <div className="relative w-full max-w-4xl mx-4 bg-white rounded-2xl shadow-2xl overflow-hidden transform transition-all duration-300 scale-100">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-lg flex items-center justify-center">
              <i className="fas fa-play text-white text-sm"></i>
            </div>
            <h3 className="text-xl font-bold text-[#2D3436]">{title}</h3>
          </div>
          <button
            onClick={onClose}
            className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-gray-100 transition-colors duration-200"
            aria-label="Close modal"
          >
            <i className="fas fa-times text-[#636E72] text-lg"></i>
          </button>
        </div>

        {/* Video Container */}
        <div className="relative w-full bg-black" style={{ paddingBottom: '56.25%' /* 16:9 aspect ratio */ }}>
          <iframe
            className="absolute top-0 left-0 w-full h-full"
            src={`https://www.youtube.com/embed/${videoId}?autoplay=1&rel=0&modestbranding=1`}
            title={title}
            frameBorder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          />
        </div>

        {/* Demo Description */}
        <div className="p-6 bg-gradient-to-r from-orange-50 to-amber-50">
          <div className="text-center">
            <h4 className="text-lg font-bold text-[#2D3436] mb-2">
              See Decorebator in Action
            </h4>
            <p className="text-[#636E72] mb-4">
              Watch how AI-powered vocabulary learning works with spaced repetition, interactive quizzes, and comprehensive analytics.
            </p>
            
            {/* Feature Highlights */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="flex items-center justify-center space-x-2 bg-white/60 rounded-lg p-3">
                <i className="fas fa-brain text-[#FF7B54]"></i>
                <span>AI Content</span>
              </div>
              <div className="flex items-center justify-center space-x-2 bg-white/60 rounded-lg p-3">
                <i className="fas fa-clock text-[#4CAF50]"></i>
                <span>Spaced Repetition</span>
              </div>
              <div className="flex items-center justify-center space-x-2 bg-white/60 rounded-lg p-3">
                <i className="fas fa-gamepad text-[#9C27B0]"></i>
                <span>7 Quiz Modes</span>
              </div>
              <div className="flex items-center justify-center space-x-2 bg-white/60 rounded-lg p-3">
                <i className="fas fa-chart-line text-[#14B8A6]"></i>
                <span>Analytics</span>
              </div>
            </div>

            {/* CTA */}
            <div className="mt-6">
              <a
                href="/signup?plan=free"
                className="inline-flex items-center space-x-2 bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-3 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300"
                onClick={onClose}
              >
                <span>Start Learning Free</span>
                <i className="fas fa-arrow-right"></i>
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VideoModal;
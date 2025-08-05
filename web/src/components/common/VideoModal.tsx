'use client'

import React, { useEffect } from 'react'
import SmartDownloadButton from './SmartDownloadButton'

interface VideoModalProps {
  isOpen: boolean
  onClose: () => void
  videoId?: string // YouTube video ID
  title?: string
}

const VideoModal: React.FC<VideoModalProps> = ({
  isOpen,
  onClose,
  videoId = 'dQw4w9WgXcQ', // Default demo video ID
  title = 'Decorebator Demo',
}) => {
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose()
      }
    }

    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = 'unset'
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = 'unset'
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/70 backdrop-blur-sm transition-opacity duration-300"
        onClick={onClose}
      />

      {/* Modal Content */}
      <div className="relative mx-4 w-full max-w-4xl scale-100 transform overflow-hidden rounded-2xl bg-white shadow-2xl transition-all duration-300">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-200 p-6">
          <div className="flex items-center space-x-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-[#FF7B54] to-orange-600">
              <i className="fas fa-play text-sm text-white"></i>
            </div>
            <h3 className="text-xl font-bold text-[#2D3436]">{title}</h3>
          </div>
          <button
            onClick={onClose}
            className="flex h-10 w-10 items-center justify-center rounded-full transition-colors duration-200 hover:bg-gray-100"
            aria-label="Close modal"
          >
            <i className="fas fa-times text-lg text-[#636E72]"></i>
          </button>
        </div>

        {/* Video Container */}
        <div
          className="relative w-full bg-black"
          style={{ paddingBottom: '56.25%' /* 16:9 aspect ratio */ }}
        >
          <iframe
            className="absolute top-0 left-0 h-full w-full"
            src={`https://www.youtube.com/embed/${videoId}?autoplay=1&rel=0&modestbranding=1`}
            title={title}
            frameBorder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          />
        </div>

        {/* Demo Description */}
        <div className="bg-gradient-to-r from-orange-50 to-amber-50 p-6">
          <div className="text-center">
            <h4 className="mb-2 text-lg font-bold text-[#2D3436]">See Decorebator in Action</h4>
            <p className="mb-4 text-[#636E72]">
              Watch how AI-powered vocabulary learning works with spaced repetition, interactive
              quizzes, and comprehensive analytics.
            </p>

            {/* Feature Highlights */}
            <div className="grid grid-cols-2 gap-4 text-sm md:grid-cols-4">
              <div className="flex items-center justify-center space-x-2 rounded-lg bg-white/60 p-3">
                <i className="fas fa-brain text-[#FF7B54]"></i>
                <span>AI Content</span>
              </div>
              <div className="flex items-center justify-center space-x-2 rounded-lg bg-white/60 p-3">
                <i className="fas fa-clock text-[#4CAF50]"></i>
                <span>Spaced Repetition</span>
              </div>
              <div className="flex items-center justify-center space-x-2 rounded-lg bg-white/60 p-3">
                <i className="fas fa-gamepad text-[#9C27B0]"></i>
                <span>7 Quiz Modes</span>
              </div>
              <div className="flex items-center justify-center space-x-2 rounded-lg bg-white/60 p-3">
                <i className="fas fa-chart-line text-[#14B8A6]"></i>
                <span>Analytics</span>
              </div>
            </div>

            {/* CTA */}
            <div className="mt-6">
              <SmartDownloadButton onClick={onClose} size="medium">
                <span>Download App</span>
                <i className="fas fa-arrow-right"></i>
              </SmartDownloadButton>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default VideoModal

"use client"

import Image from 'next/image'

type Props = {
  shareTitle: string
  isCompleted: boolean
  currentIndex: number
  total: number
  score: number
  onClose: () => void
}

export default function QuizHeader({ shareTitle, isCompleted, currentIndex, total, score, onClose }: Props) {
  return (
    <div className="relative flex items-center justify-between border-b border-gray-100 p-4 sm:p-6">
      {/* Drag handle for mobile */}
      <div className="absolute left-1/2 top-1 -translate-x-1/2 sm:hidden">
        <div className="h-1.5 w-12 rounded-full bg-gray-300" />
      </div>

      <div className="flex items-center space-x-3 min-w-0">
        <div className="flex h-12 w-12 items-center justify-center">
          <Image src="/quiz-icon.png" alt="Quiz" width={36} height={36} />
        </div>
        <div className="min-w-0 sm:max-w-none">
          <h2 className="text-xl sm:text-2xl font-bold text-[#2D3436] leading-snug break-words" title={shareTitle}>
            {shareTitle}
          </h2>
          {!isCompleted && total > 0 && (
            <p className="text-sm text-[#636E72]">
              Question {currentIndex + 1} of {total} • Score: {score}/{currentIndex}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center space-x-2">
        <button
          onClick={onClose}
          className="flex h-10 w-10 items-center justify-center rounded-full transition-all duration-200 hover:bg-gray-100 active:scale-95"
        >
          <i className="fas fa-times text-lg text-[#636E72]"></i>
        </button>
      </div>
    </div>
  )
}



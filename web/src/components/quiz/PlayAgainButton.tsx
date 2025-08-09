"use client"

import React, { useState } from 'react'
import QuizPlayer from '@/components/quiz/QuizPlayer'
import type { QuizQuestion as Quiz } from '@/lib/api'

type Props = {
  quizzes: Quiz[]
  slug: string
  title: string
  className?: string
}

export default function PlayAgainButton({ quizzes, slug, title, className }: Props) {
  const [launchCount, setLaunchCount] = useState(0)

  return (
    <div className={className}>
      <button
        type="button"
        onClick={() => setLaunchCount((n) => n + 1)}
        className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-500 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:from-[#ff6b45] hover:to-orange-500 focus:outline-none focus:ring-2 focus:ring-orange-400 focus:ring-offset-2"
        aria-label="Play this quiz again"
      >
        <span className="text-base">▶</span>
        Play again
      </button>

      {launchCount > 0 ? (
        <QuizPlayer key={launchCount} quizzes={quizzes} initiallyOpen={true} slug={slug} title={title} />
      ) : null}
    </div>
  )
}



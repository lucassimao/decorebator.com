"use client"

import React, { useState } from 'react'
import Image from 'next/image'
import QuizPlayer from '@/components/quiz/QuizPlayer'
import AppStoreButton from '@/components/common/AppStoreButton'
import type { QuizQuestion as Quiz } from '@/lib/api'

type Props = {
  quizzes: Quiz[]
  slug: string
  title: string
}

export default function QuizWithFooter({ quizzes, slug, title }: Props) {
  const [open, setOpen] = useState(true)

  // Hide footer when ANY quiz modal on the page is open (including PlayAgainButton instances)
  React.useEffect(() => {
    const handler = (e: Event) => {
      const detail = (e as CustomEvent<boolean>).detail
      if (typeof detail === 'boolean') setOpen(detail)
    }
    if (typeof window !== 'undefined') {
      // Sync initial state with html class in case the event fired before mount
      try {
        const active = document.documentElement.classList.contains('quiz-open')
        setOpen(active)
      } catch {}
      window.addEventListener('quiz-open-change', handler as EventListener)
      return () => window.removeEventListener('quiz-open-change', handler as EventListener)
    }
    return () => {}
  }, [])

  return (
    <div className="relative">
      <QuizPlayer quizzes={quizzes} initiallyOpen={true} slug={slug} title={title} onOpenChange={setOpen} />

      <footer
        className={
          'pointer-events-none fixed inset-x-0 bottom-4 z-40 transition-opacity duration-300 ' +
          (open ? 'opacity-0' : 'opacity-100')
        }
      >
        <div className="pointer-events-auto mx-auto w-full max-w-3xl px-4">
          <div className="flex flex-col items-stretch gap-3 rounded-2xl border border-orange-200 bg-white/90 p-3 shadow-lg backdrop-blur sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-orange-100">
                <Image src="/quiz-icon.png" alt="Decorebator" width={20} height={20} />
              </div>
              <div className="text-sm text-[#2D3436]">
                <p className="font-semibold leading-tight">Turn your wordlists into viral quizzes</p>
                <p className="text-xs text-[#636E72]">Create, share, and grow your audience with Decorebator.</p>
              </div>
            </div>
            <div className="flex items-center justify-end gap-2">
              <div className="hidden sm:flex items-center gap-2">
                <AppStoreButton store="apple" size="small" />
                <AppStoreButton store="google" size="small" />
              </div>
              <div className="flex items-center gap-2 sm:hidden">
                <AppStoreButton store="apple" size="small" />
                <AppStoreButton store="google" size="small" />
              </div>
              <a
                href="https://decorebator.com"
                target="_blank"
                rel="noreferrer"
                className="rounded-xl border border-gray-200 px-4 py-2 text-sm font-semibold text-[#2D3436] hover:bg-gray-50 whitespace-nowrap"
              >
                Learn more
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}



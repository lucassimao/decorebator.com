'use client'

import React, { useEffect, useMemo, useState } from 'react'
import { createPortal } from 'react-dom'
import QuizDemoModal from './QuizDemoModal'
import type { QuizQuestion as Quiz } from '@/lib/api'

type Props = {
  quizzes: Quiz[]
  initiallyOpen?: boolean
  slug?: string
  title?: string
  onOpenChange?: (open: boolean) => void
}

export default function QuizPlayer({ quizzes, initiallyOpen = true, slug, title, onOpenChange }: Props) {
  const [isOpen, setIsOpen] = useState(initiallyOpen)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [completed, setCompleted] = useState(false)
  // Stats kept for future use, disable unused-var lint
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [tried, setTried] = useState(0)
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [correct, setCorrect] = useState(0)

  // Auto-open when quizzes arrive and initiallyOpen is true
  useEffect(() => {
    if (initiallyOpen && quizzes?.length > 0) {
      setIsOpen(true)
      onOpenChange?.(true)
    }
  }, [initiallyOpen, quizzes, onOpenChange])

  useEffect(() => {
    onOpenChange?.(isOpen)
    if (typeof window !== 'undefined') {
      try {
        const evt = new CustomEvent('quiz-open-change', { detail: isOpen })
        window.dispatchEvent(evt)
        document.documentElement.classList.toggle('quiz-open', isOpen)
      } catch {}
    }
  }, [isOpen, onOpenChange])

  // Shim: wrap QuizDemoModal to capture completion stats
  const Wrapped = useMemo(() => {
    return function WrappedQuizModal({ open }: { open: boolean }) {
      const shareUrl = (() => {
        if (!slug) return ''
        if (typeof window !== 'undefined' && window.location?.origin) {
          return `${window.location.origin}/q/${slug}`
        }
        return `https://decorebator.com/q/${slug}`
      })()
      return (
        <QuizDemoModal
          isOpen={open}
          onClose={() => setIsOpen(false)}
          demoQuizzes={quizzes}
          shareUrl={shareUrl}
          shareTitle={title || 'Decorebator Quiz'}
          onComplete={({ tried, correct }: { tried: number; correct: number }) => {
            setCompleted(true)
            setTried(tried)
            setCorrect(correct)
          }}
        />
      )
    }
  }, [quizzes, slug, title])

  const [mounted, setMounted] = useState(false)
  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted || typeof document === 'undefined') {
    return null
  }
  return createPortal(
    <div className="supports-[padding:env(safe-area-inset-bottom)]:pb-[env(safe-area-inset-bottom)] supports-[padding:env(safe-area-inset-top)]:pt-[env(safe-area-inset-top)]">
      <Wrapped open={isOpen} />
    </div>,
    document.body
  )
}



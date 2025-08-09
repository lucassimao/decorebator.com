'use client'

import React, { useState, useEffect } from 'react'
import Image from 'next/image'
import QuizHeader from '@/components/quiz/modal/QuizHeader'
import QuizProgress from '@/components/quiz/modal/QuizProgress'
import FastModeToggle from '@/components/quiz/modal/FastModeToggle'
import StickyActionBar from '@/components/quiz/modal/StickyActionBar'
import type { QuizQuestion as Quiz } from '@/lib/api'
import { submitPublicQuizAttempt } from '@/lib/api'
import AppStoreButton from '@/components/common/AppStoreButton'
import ShareQuizButton from '@/components/quiz/ShareQuizButton'

interface QuizDemoModalProps {
  isOpen: boolean
  onClose: () => void
  demoQuizzes: Quiz[]
  onComplete?: (stats: { tried: number; correct: number }) => void
  shareUrl?: string
  shareTitle?: string
}



// Removed ErrorReportModal for public demo

const QuizDemoModal: React.FC<QuizDemoModalProps> = ({ isOpen, onClose, demoQuizzes, onComplete, shareUrl = '', shareTitle = 'Decorebator Quiz' }) => {
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0)
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null)
  const [showFeedback, setShowFeedback] = useState(false)
  const [quizSet, setQuizSet] = useState<Quiz[]>([])
  const [score, setScore] = useState(0)
  const [isCompleted, setIsCompleted] = useState(false)
  const [isPlayingAudio, setIsPlayingAudio] = useState(false)
  const htmlAudioRef = React.useRef<HTMLAudioElement | null>(null)
  const [fastMode, setFastMode] = useState(false)
  // reporting removed
  const [userInput, setUserInput] = useState('')
  const [isWriteMode, setIsWriteMode] = useState(false)
  const [revealed, setRevealed] = useState<boolean[]>([])
  const reportedRef = React.useRef(false)
  const [imageLoading, setImageLoading] = useState(false)
  const [imageError, setImageError] = useState(false)
  const [playerName, setPlayerName] = useState('')
  const [playerEmail, setPlayerEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [startedAt, setStartedAt] = useState<number | null>(null)
  const [nameError, setNameError] = useState<string>('')
  const [emailError, setEmailError] = useState<string>('')
  const [submitOk, setSubmitOk] = useState<boolean | null>(null)
  const [submittedLocked, setSubmittedLocked] = useState(false)
  // const [quizSlug, setQuizSlug] = useState<string | null>(null)

  // no-op

  // Build a randomized quiz set up to MAX_QUESTIONS ensuring at least one of each available type
  const buildQuizSet = (provided: Quiz[], shareUrlLocal: string): Quiz[] => {
    const list = Array.isArray(provided) ? provided : []
    if (list.length === 0) return []

    // Partition by seen/unseen using slug key in localStorage
    let slug: string | undefined
    try {
      const slugMatch = (shareUrlLocal || '').match(/\/q\/([^\/?#]+)/)
      slug = slugMatch?.[1]
    } catch {}
    const key = slug ? `pq_seen:${slug}` : ''
    const seenIds: number[] = key ? JSON.parse(localStorage.getItem(key) || '[]') : []
    const seenSet = new Set<number>(seenIds)
    const unseen: Quiz[] = []
    const seen: Quiz[] = []
    for (const q of list) {
      if (seenSet.has(q.definitionId)) seen.push(q)
      else unseen.push(q)
    }

    // Deterministic daily seed using slug + date
    const seedBase = `${slug || 'quiz'}:${new Date().toISOString().slice(0, 10)}`
    let seed = 0
    for (let i = 0; i < seedBase.length; i++) seed = (seed * 31 + seedBase.charCodeAt(i)) >>> 0
    const seededShuffle = (arr: Quiz[]) => {
      const copy = [...arr]
      for (let i = copy.length - 1; i > 0; i--) {
        seed = (1664525 * seed + 1013904223) >>> 0
        const j = seed % (i + 1)
        const tmp = copy[i]
        copy[i] = copy[j]
        copy[j] = tmp
      }
      return copy
    }

    const reordered: Quiz[] = [...seededShuffle(unseen), ...seededShuffle(seen)]

    // Ensure at least one of each available type, then fill up to MAX_QUESTIONS
    const MAX_QUESTIONS = 20
    const maxTake = Math.min(MAX_QUESTIONS, reordered.length)
    const firstByType = new Map<Quiz['type'], Quiz>()
    for (const q of reordered) {
      if (!firstByType.has(q.type)) firstByType.set(q.type, q)
    }
    const picks: Quiz[] = Array.from(firstByType.values())
    const pickedIds = new Set<number>(picks.map((q) => q.definitionId))
    for (const q of reordered) {
      if (picks.length >= maxTake) break
      if (!pickedIds.has(q.definitionId)) {
        picks.push(q)
        pickedIds.add(q.definitionId)
      }
    }
    // If we still don't reach maxTake due to heavy duplicates, allow repeats
    let idx = 0
    while (picks.length < maxTake && reordered.length > 0) {
      picks.push(reordered[idx % reordered.length])
      idx++
    }

    return picks
  }

  const validateName = (name: string) => {
    if (!name || name.trim().length < 2) return 'Please enter your name'
    return ''
  }
  const validateEmail = (email: string) => {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!email || !re.test(email)) return 'Enter a valid email address'
    return ''
  }
  const isFormValid = () => validateName(playerName) === '' && validateEmail(playerEmail) === ''

  // Initialize quiz set when modal opens
  useEffect(() => {
    if (isOpen) {
      const provided = Array.isArray(demoQuizzes) ? demoQuizzes : []
      try {
        const nextSet = buildQuizSet(provided, shareUrl)
        setQuizSet(nextSet)
        // lock submission if already submitted
        const m = (shareUrl || '').match(/\/q\/([^\/?#]+)/)
        const slug = m?.[1]
        const stored = slug ? localStorage.getItem(`pq_submitted:${slug}`) : null
        setSubmittedLocked(stored === '1')
        // initialize from first question of the selected set
        setIsWriteMode(nextSet[0]?.type === 'WRITE_WORD_FROM_DEFINITION')
        setRevealed(new Array(nextSet[0]?.options?.length || 0).fill(false))
      } catch {
        setQuizSet(provided)
        setIsWriteMode(provided[0]?.type === 'WRITE_WORD_FROM_DEFINITION')
        setRevealed(new Array(provided[0]?.options?.length || 0).fill(false))
      }
      setCurrentQuestionIndex(0)
      setSelectedAnswer(null)
      setShowFeedback(false)
      setScore(0)
      setIsCompleted(false)
      setUserInput('')
      // reporting removed
      reportedRef.current = false
      setStartedAt(Date.now())
    }
  }, [isOpen, demoQuizzes, shareUrl])

  const currentQuiz = quizSet[currentQuestionIndex]

  // Update write mode when question changes
  useEffect(() => {
    if (currentQuiz) {
      setIsWriteMode(currentQuiz.type === 'WRITE_WORD_FROM_DEFINITION')
      if (currentQuiz.type === 'WORD_FROM_IMAGE') {
        setImageLoading(true)
        setImageError(false)
      } else {
        setImageLoading(false)
        setImageError(false)
      }
      // reset reveal state for new question
      const count = currentQuiz.options?.length || 0
      setRevealed(new Array(count).fill(false))
      // stagger reveal
      const timeouts: number[] = []
      for (let i = 0; i < count; i++) {
        const id = window.setTimeout(() => {
          setRevealed((prev) => {
            const next = [...prev]
            next[i] = true
            return next
          })
        }, 60 * i)
        timeouts.push(id)
      }
      return () => {
        timeouts.forEach((t) => window.clearTimeout(t))
      }
    }
  }, [currentQuiz, shareUrl])

  // Cleanup audio when modal closes
  useEffect(() => {
    if (!isOpen) {
      // Stop any playing audio when modal closes
      if (htmlAudioRef.current) {
        try {
          htmlAudioRef.current.pause()
          htmlAudioRef.current.currentTime = 0
        } catch {}
        htmlAudioRef.current = null
      }
      if ('speechSynthesis' in window) {
        speechSynthesis.cancel()
      }
      setIsPlayingAudio(false)
    }
  }, [isOpen])

  // Stop audio when navigating to different questions
  useEffect(() => {
    if (htmlAudioRef.current) {
      try {
        htmlAudioRef.current.pause()
        htmlAudioRef.current.currentTime = 0
      } catch {}
      htmlAudioRef.current = null
    }
    if ('speechSynthesis' in window) {
      speechSynthesis.cancel()
    }
    setIsPlayingAudio(false)
  }, [currentQuestionIndex])

  const handleAnswerSelect = (answerIndex: number) => {
    if (showFeedback) return // Prevent multiple selections

    setSelectedAnswer(answerIndex)
    setShowFeedback(true)

    // Check if correct
    const isCorrect = answerIndex === currentQuiz.answerIndex
    if (isCorrect) {
      setScore(score + 1)
    }

    // Auto-advance in fast mode
    if (fastMode) {
      setTimeout(() => {
        handleNext()
      }, 600)
    }
  }

  const handleWriteAnswer = () => {
    if (!userInput.trim()) return

    setShowFeedback(true)

    // Check if correct (case/diacritics insensitive)
    const normalize = (s: string) =>
      s
        .toLowerCase()
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .trim()
    const isCorrect =
      normalize(userInput) === normalize(currentQuiz.options[currentQuiz.answerIndex] || '')
    if (isCorrect) {
      setScore(score + 1)
      setSelectedAnswer(currentQuiz.answerIndex)
    } else {
      setSelectedAnswer(-1) // Mark as wrong
    }

    // Auto-advance in fast mode
    if (fastMode) {
      setTimeout(() => {
        handleNext()
      }, 600)
    }
  }

  const handleSkipQuestion = () => {
    // Treat skip as incorrect and immediately proceed without showing feedback
    setSelectedAnswer(null)
    setShowFeedback(false)
    handleNext()
  }

  // reporting removed

  const toggleFastMode = () => {
    setFastMode(!fastMode)

    // If enabling fast mode and feedback is shown, auto-advance
    if (!fastMode && showFeedback) {
      setTimeout(() => {
        handleNext()
      }, 600)
    }
  }

  const handleNext = () => {
    if (currentQuestionIndex < quizSet.length - 1) {
      setCurrentQuestionIndex(currentQuestionIndex + 1)
      setSelectedAnswer(null)
      setShowFeedback(false)
      setUserInput('')
      setIsWriteMode(currentQuiz?.type === 'WRITE_WORD_FROM_DEFINITION')
    } else {
      setIsCompleted(true)
      // Persist seen locally
      try {
        const slugMatch = (shareUrl || '').match(/\/q\/([^/?#]+)/)
        const slug = slugMatch?.[1]
        if (slug) {
          const key = `pq_seen:${slug}`
          const prev: number[] = JSON.parse(localStorage.getItem(key) || '[]')
          const set = new Set<number>(prev)
          for (const q of quizSet) set.add(q.definitionId)
          const merged = Array.from(set).slice(-120)
          localStorage.setItem(key, JSON.stringify(merged))
        }
      } catch {}
    }
  }

  // Notify completion once (still notify parent for UI), but attempt submission handled here
  useEffect(() => {
    if (isCompleted && !reportedRef.current) {
      reportedRef.current = true
      onComplete?.({ tried: quizSet.length, correct: score })
    }
  }, [isCompleted, quizSet.length, score, onComplete])

  const submitAttempt = async () => {
    try {
      if (!shareUrl) return
      const m = shareUrl.match(/\/q\/([^\/?#]+)/)
      const slug = m?.[1]
      if (!slug) return
      setSubmitting(true)
      const durationSeconds = startedAt ? Math.max(0, Math.round((Date.now() - startedAt) / 1000)) : undefined
      await submitPublicQuizAttempt({
        slug,
        tried: quizSet.length,
        correct: score,
        name: playerName || undefined,
        email: playerEmail || undefined,
        durationSeconds,
      })
      setSubmitOk(true)
      setSubmittedLocked(true)
      try { localStorage.setItem(`pq_submitted:${slug}`, '1') } catch {}
      if (typeof window !== 'undefined') {
        window.setTimeout(() => setSubmitOk(null), 3000)
      }
    } catch {
      // ignore errors for MVP
      setSubmitOk(false)
      if (typeof window !== 'undefined') {
        window.setTimeout(() => setSubmitOk(null), 3500)
      }
    } finally {
      setSubmitting(false)
    }
  }

  const handleRestart = () => {
    const provided = Array.isArray(demoQuizzes) ? demoQuizzes : []
    try {
      const nextSet = buildQuizSet(provided, shareUrl || '')
      setQuizSet(nextSet)
      setIsWriteMode(nextSet[0]?.type === 'WRITE_WORD_FROM_DEFINITION')
      setRevealed(new Array(nextSet[0]?.options?.length || 0).fill(false))
    } catch {
      setQuizSet(provided)
      setIsWriteMode(provided[0]?.type === 'WRITE_WORD_FROM_DEFINITION')
      setRevealed(new Array(provided[0]?.options?.length || 0).fill(false))
    }
    setCurrentQuestionIndex(0)
    setSelectedAnswer(null)
    setShowFeedback(false)
    setScore(0)
    setIsCompleted(false)
    setUserInput('')
  }

  // Audio playing functionality using Web Speech API
  const pickPreferredVoice = (): SpeechSynthesisVoice | null => {
    if (!('speechSynthesis' in window)) return null
    const voices = window.speechSynthesis.getVoices()
    if (!voices || voices.length === 0) return null
    const preferred = [
      'Google US English',
      'Google UK English Female',
      'Google UK English Male',
      'Microsoft Aria Online (Natural) - English (United States)',
      'Microsoft Guy Online (Natural) - English (United States)',
    ]
    for (const name of preferred) {
      const v = voices.find((vv) => vv.name === name)
      if (v) return v
    }
    const en = voices.find((v) => v.lang?.toLowerCase().startsWith('en'))
    return en || voices[0]
  }

  const playAudio = (text: string, audioUrl?: string) => {
    if (isPlayingAudio) return
    setIsPlayingAudio(true)

    // Prefer real audio URL if provided
    if (audioUrl) {
      try {
        // Stop any existing audio
        if (htmlAudioRef.current) {
          try {
            htmlAudioRef.current.pause()
            htmlAudioRef.current.currentTime = 0
          } catch {}
        }
        const audio = new Audio(audioUrl)
        htmlAudioRef.current = audio
        audio.onended = () => {
          setIsPlayingAudio(false)
        }
        audio.onerror = () => {
          // Fallback to speech synthesis if playback fails
          htmlAudioRef.current = null
          playWithSpeechSynthesis(text)
        }
        audio.play().catch(() => {
          // Autoplay blocked or error — fallback to speech
          htmlAudioRef.current = null
          playWithSpeechSynthesis(text)
        })
        return
      } catch {
        // In case constructing Audio fails, fallback
        htmlAudioRef.current = null
        playWithSpeechSynthesis(text)
        return
      }
    }

    // No URL — use speech synthesis
    playWithSpeechSynthesis(text)
  }

  const playWithSpeechSynthesis = (text: string) => {
    // Use Web Speech API for text-to-speech
    if ('speechSynthesis' in window) {
      const speak = () => {
        const utterance = new SpeechSynthesisUtterance(text)
        const v = pickPreferredVoice()
        if (v) utterance.voice = v
        utterance.rate = 0.95
        utterance.pitch = 1.0
        utterance.volume = 0.95

        utterance.onend = () => {
          setIsPlayingAudio(false)
        }

        utterance.onerror = () => {
          setIsPlayingAudio(false)
        }

        speechSynthesis.speak(utterance)
      }

      if (speechSynthesis.getVoices().length === 0) {
        window.speechSynthesis.onvoiceschanged = () => {
          speak()
          window.speechSynthesis.onvoiceschanged = null
        }
      } else {
        speak()
      }
    } else {
      // Final fallback if no speech synthesis
      setTimeout(() => {
        setIsPlayingAudio(false)
      }, 1200)
    }
  }

  const getQuizTitle = () => {
    if (!currentQuiz) return ''

    switch (currentQuiz.type) {
      case 'GUESS_MEANING':
        return 'What does this word mean?'
      case 'WORD_FROM_MEANING':
        return 'Which word matches this definition?'
      case 'WORD_FROM_IMAGE':
        return 'Which word does this image represent?'
      case 'COMPLETE_SENTENCE':
        return 'Complete the sentence:'
      case 'WRITE_WORD_FROM_DEFINITION':
        return 'What word matches this definition?'
      case 'WORD_FROM_AUDIO':
        return 'Which word did you hear?'
      case 'WORD_FROM_EXAMPLE_AUDIO':
        return 'Complete the sentence:'
      default:
        return 'Choose the correct answer:'
    }
  }

  const getOptionStyle = (index: number) => {
    let baseClasses =
      'flex items-center justify-between w-full p-4 rounded-xl border-2 transition-all duration-300 text-left transform '

    if (!showFeedback) {
      baseClasses += 'border-gray-200 bg-gray-50 hover:border-[#FF7B54] hover:bg-orange-50 '
      baseClasses += revealed[index] ? 'opacity-100 translate-y-0 ' : 'opacity-0 translate-y-1 '
    } else {
      if (index === currentQuiz.answerIndex) {
        baseClasses += 'border-green-500 bg-green-50 '
      } else if (index === selectedAnswer && selectedAnswer !== currentQuiz.answerIndex) {
        baseClasses += 'border-red-500 bg-red-50 '
      } else {
        baseClasses += 'border-gray-200 bg-gray-50 opacity-60 '
      }
    }

    return baseClasses
  }

  const getOptionTextStyle = (index: number) => {
    let baseClasses = 'flex-1 text-lg '

    if (!showFeedback) {
      baseClasses += 'text-[#2D3436] '
    } else {
      if (index === currentQuiz.answerIndex) {
        baseClasses += 'text-green-700 font-semibold '
      } else if (index === selectedAnswer && selectedAnswer !== currentQuiz.answerIndex) {
        baseClasses += 'text-red-700 '
      } else {
        baseClasses += 'text-[#636E72] '
      }
    }

    return baseClasses
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-[9999] flex items-stretch sm:items-center sm:justify-center">
      {/* Backdrop (dim background on all devices) */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity duration-300" onClick={onClose} />

        {/* Modal/Full-screen Content */}
        <div className="relative mx-0 min-h-[100dvh] w-screen overflow-hidden rounded-none bg-white/95 shadow-none animate-[fadeIn_200ms_ease-out] flex flex-col sm:mx-4 sm:h-auto sm:max-h-[90vh] sm:w-full sm:max-w-2xl sm:rounded-3xl sm:shadow-2xl supports-[padding:env(safe-area-inset-bottom)]:pb-[env(safe-area-inset-bottom)] supports-[padding:env(safe-area-inset-top)]:pt-[env(safe-area-inset-top)]">
        {/* Header */}
        <QuizHeader
          shareTitle={shareTitle}
          isCompleted={isCompleted}
          currentIndex={currentQuestionIndex}
          total={quizSet.length}
          score={currentQuestionIndex + (showFeedback ? 1 : 0) ? score : score}
          onClose={onClose}
        />

        {/* Progress Bar */}
        <QuizProgress current={currentQuestionIndex + 1} total={quizSet.length} hidden={isCompleted} />

        {!isCompleted && currentQuiz && (
          <FastModeToggle enabled={fastMode} onToggle={toggleFastMode} />
        )}

        {/* Quiz Content */}
        <div className="p-4 sm:p-8 flex-1 overflow-y-auto pb-28 sm:pb-4">
          {isCompleted ? (
            // Completion Screen
            <div className="space-y-6 text-center">
              <div className="mx-auto flex items-center justify-center">
                <Image src="/quiz-icon.png" alt="Quiz complete" width={80} height={80} />
              </div>

              <div>
                <h3 className="mb-2 text-2xl font-bold text-[#2D3436]">Quiz Complete!</h3>
                <p className="text-lg text-[#636E72]">
                  You scored {score} out of {quizSet.length} questions
                </p>
              </div>

              <div className="grid grid-cols-1 gap-3">
                <div className="grid grid-cols-1 gap-3">
                  <ShareQuizButton
                    url={shareUrl}
                    title={shareTitle}
                    text={`I scored ${quizSet.length > 0 ? Math.round((score / quizSet.length) * 100) : 0}% on “${shareTitle}” — can you beat me?`}
                    label="Share my score"
                    showLabelOnMobile={true}
                    className="flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-all duration-300 hover:shadow-lg w-full"
                  />

                {(submittedLocked || submitOk === true) ? null : (
                <div className="rounded-2xl border border-gray-200 p-4 text-left">
                  <p className="mb-3 text-sm font-semibold text-[#2D3436]">Want to appear on the leaderboard?</p>
                  {submitOk === false ? (
                    <div className="mb-3 flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                      <i className="fas fa-exclamation-circle" />
                      <span>Couldn’t submit right now. Please try again.</span>
                    </div>
                  ) : null}
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <input
                      type="text"
                      value={playerName}
                      onChange={(e) => {
                        setPlayerName(e.target.value)
                        setNameError(validateName(e.target.value))
                      }}
                      onBlur={() => setNameError(validateName(playerName))}
                      placeholder="Your name"
                      disabled={submittedLocked}
                      className={`w-full rounded-xl border-2 bg-white px-4 py-3 text-sm text-[#2D3436] placeholder:text-gray-400 focus:outline-none ${submittedLocked ? 'opacity-60 cursor-not-allowed' : nameError ? 'border-red-500 focus:border-red-500' : 'border-gray-200 focus:border-[#FF7B54]'}`}
                    />
                    {nameError ? <p className="text-xs text-red-600">{nameError}</p> : null}
                    <input
                      type="email"
                      value={playerEmail}
                      onChange={(e) => {
                        setPlayerEmail(e.target.value)
                        setEmailError(validateEmail(e.target.value))
                      }}
                      onBlur={() => setEmailError(validateEmail(playerEmail))}
                      placeholder="Your email"
                      disabled={submittedLocked}
                      className={`w-full rounded-xl border-2 bg-white px-4 py-3 text-sm text-[#2D3436] placeholder:text-gray-400 focus:outline-none ${submittedLocked ? 'opacity-60 cursor-not-allowed' : emailError ? 'border-red-500 focus:border-red-500' : 'border-gray-200 focus:border-[#FF7B54]'}`}
                    />
                    {emailError ? <p className="text-xs text-red-600">{emailError}</p> : null}
                  </div>
                  <div className="mt-3 flex items-center justify-end gap-2">
                    <button
                      disabled={submitting || !isFormValid() || Boolean(submitOk) || submittedLocked}
                      onClick={async () => {
                        const nErr = validateName(playerName)
                        const eErr = validateEmail(playerEmail)
                        setNameError(nErr)
                        setEmailError(eErr)
                        if (nErr || eErr) return
                        await submitAttempt()
                      }}
                      className="rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                      {submitting ? 'Submitting…' : (Boolean(submitOk) || submittedLocked) ? 'Submitted' : 'Submit to leaderboard'}
                    </button>
                  </div>
                </div>
                )}

                {/* Persuasive conversion block */}
                <div className="rounded-2xl border border-gray-200 p-5 text-left bg-white/70">
                  <div className="flex flex-col gap-2">
                    <h4 className="text-lg font-semibold text-[#2D3436]">Turn this win into a daily habit</h4>
                    <p className="text-sm text-[#636E72]">
                      Decorebator helps you remember new words with smart spaced repetition, bite‑size quizzes,
                      and rich audio & images. Build momentum, one quick session at a time.
                    </p>
                    <div className="flex items-center gap-2 pt-1 text-[#2D3436]">
                      <div className="text-[#F59E0B]">
                        <i className="fas fa-star"></i>
                        <i className="fas fa-star ml-0.5"></i>
                        <i className="fas fa-star ml-0.5"></i>
                        <i className="fas fa-star ml-0.5"></i>
                        <i className="fas fa-star ml-0.5"></i>
                      </div>
                      <span className="text-xs sm:text-sm text-[#636E72]">Loved by language learners worldwide</span>
                    </div>
                    <ul className="grid gap-1 sm:grid-cols-2 pl-1 text-sm text-[#2D3436]">
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        Daily reminders to keep you consistent
                      </li>
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        Smart repetition focuses on what you forget
                      </li>
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        8 quiz modes keep practice engaging
                      </li>
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        Works offline on the go
                      </li>
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        Add your own wordlists and share with friends
                      </li>
                      <li className="pl-5 relative">
                        <i className="fas fa-check absolute left-0 top-0.5 text-green-600"></i>
                        Track progress and celebrate milestones
                      </li>
                    </ul>
                    <div className="pt-2 text-xs text-[#8E979B]">No ads • Privacy‑first • Optional premium</div>
                  </div>
                </div>
                </div>
              </div>

              <div className="mt-2 flex flex-col items-center justify-center gap-3 sm:flex-row">
                <AppStoreButton store="apple" size="small" />
                <AppStoreButton store="google" size="small" />
              </div>

              <div className="pt-1">
                <button
                  onClick={() => {
                    try {
                      const m = (shareUrl || '').match(/\/q\/([^/?#]+)/)
                      const slug = m?.[1]
                      if (!slug) return
                      localStorage.removeItem(`pq_seen:${slug}`)
                    } catch {}
                  }}
                  className="text-xs text-[#999] underline-offset-2 hover:underline"
                >
                  Reset my progress for this quiz
                </button>
              </div>

              <div className="pt-2">
                <button
                  onClick={handleRestart}
                  className="text-sm font-semibold text-[#636E72] underline-offset-2 hover:underline"
                >
                  Play again
                </button>
              </div>
            </div>
          ) : currentQuiz ? (
            // Quiz Question
            <div className="space-y-6">


              {/* Question Title */}
              <h3 className="text-center text-xl font-semibold text-[#2D3436]">{getQuizTitle()}</h3>

              {/* Word/Question Display */}
              <div className="space-y-4 text-center">
                {currentQuiz.type === 'GUESS_MEANING' ? (
                  <div>
                    <div className="mb-2 text-4xl font-bold text-[#FF7B54]">
                      {currentQuiz.value}
                    </div>
                    {currentQuiz.pronunciation && (
                      <div className="font-mono text-lg text-[#636E72]">
                        {currentQuiz.pronunciation}
                      </div>
                    )}
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_AUDIO' ? (
                  // Only show play button for audio questions - no word displayed
                  <div className="rounded-xl bg-blue-50 p-8 text-center">
                    <p className="mb-4 text-lg text-[#2D3436]">
                      Listen to the audio and select the word you hear
                    </p>
                    <button
                      onClick={() => playAudio(currentQuiz.value, currentQuiz.audioURL)}
                      disabled={isPlayingAudio}
                    className={`mx-auto flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600 transition-transform duration-200 hover:scale-105 hover:shadow-lg ${
                        isPlayingAudio ? 'cursor-not-allowed opacity-70' : ''
                      }`}
                    >
                      {isPlayingAudio ? (
                        <i className="fas fa-spinner fa-spin text-2xl text-white"></i>
                      ) : (
                        <i className="fas fa-play text-2xl text-white"></i>
                      )}
                    </button>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_MEANING' ||
                  currentQuiz.type === 'WRITE_WORD_FROM_DEFINITION' ? (
                  <div className="rounded-xl bg-blue-50 p-6">
                    <p className="text-lg text-[#2D3436]">{currentQuiz.value}</p>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_EXAMPLE_AUDIO' ? (
                  // For example audio, show the sentence with blank and play button
                  <div className="rounded-xl bg-purple-50 p-6 text-center">
                    <p className="mb-4 text-lg text-[#2D3436]">{currentQuiz.value}</p>
                    <button
                      onClick={() => playAudio(currentQuiz.value, currentQuiz.audioURL)}
                      disabled={isPlayingAudio}
                      className={`mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600 transition-transform duration-200 hover:scale-105 hover:shadow-lg ${
                        isPlayingAudio ? 'cursor-not-allowed opacity-70' : ''
                      }`}
                    >
                      {isPlayingAudio ? (
                        <i className="fas fa-spinner fa-spin text-lg text-white"></i>
                      ) : (
                        <i className="fas fa-play text-lg text-white"></i>
                      )}
                    </button>
                  </div>
                ) : currentQuiz.type === 'WORD_FROM_IMAGE' ? (
                  <div className="rounded-xl overflow-hidden bg-gray-100">
                    {imageLoading && (
                      <div className="flex h-64 w-full items-center justify-center animate-pulse bg-gray-200">
                        <i className="fas fa-image text-3xl text-gray-400"></i>
                      </div>
                    )}
                    {!imageError && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={currentQuiz.value}
                        alt={currentQuiz.value || 'Quiz image'}
                        onLoad={() => setImageLoading(false)}
                        onError={() => {
                          setImageLoading(false)
                          setImageError(true)
                        }}
                        className={`${imageLoading ? 'hidden' : 'block'} max-h-64 sm:max-h-96 w-full object-cover`}
                      />
                    )}
                    {imageError && (
                      <div className="flex h-64 w-full flex-col items-center justify-center bg-gray-200 text-center">
                        <i className="fas fa-broken-image mb-2 text-3xl text-gray-500"></i>
                        <p className="text-sm text-gray-600">Image unavailable</p>
                      </div>
                    )}
                    {currentQuiz.imageDescription && (
                      <div className="border-t border-gray-200 p-4 text-center text-[#636E72]">
                        <p className="italic">{currentQuiz.imageDescription}</p>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="rounded-xl bg-purple-50 p-6">
                    <p className="text-lg text-[#2D3436]">{currentQuiz.value}</p>
                  </div>
                )}
              </div>

              {/* Answer Options or Write Input */}
              {isWriteMode ? (
                <div className="space-y-4">
                  <div className="space-y-3">
                    <input
                      type="text"
                      value={userInput}
                      onChange={(e) => setUserInput(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && !showFeedback && handleWriteAnswer()}
                      placeholder="Type your answer here..."
                      disabled={showFeedback}
                      className="w-full rounded-xl border-2 border-gray-200 bg-white p-4 text-lg text-[#2D3436] placeholder:text-gray-400 transition-colors duration-200 focus:border-[#FF7B54] focus:outline-none disabled:bg-gray-50"
                    />
                    {showFeedback && (
                      selectedAnswer === currentQuiz.answerIndex ? (
                        <div className="rounded-xl border-2 border-green-500 bg-green-50 p-4">
                          <div className="flex items-center justify-between">
                            <span className="font-semibold text-green-700">Correct!</span>
                            <i className="fas fa-check-circle text-xl text-green-500"></i>
                          </div>
                        </div>
                      ) : (
                        <div className="rounded-xl border-2 border-red-500 bg-red-50 p-4">
                          <div className="flex items-center justify-between">
                            <span className="text-red-700">
                              Not quite. Correct answer: {currentQuiz.options[currentQuiz.answerIndex]}
                            </span>
                            <i className="fas fa-times-circle text-xl text-red-500"></i>
                          </div>
                        </div>
                      )
                    )}
                  </div>

                  {/* Hide inline footer buttons across all sizes when sticky bar is present */}
                  {!showFeedback && (
                    <div className="hidden">
                      <button
                        onClick={handleWriteAnswer}
                        disabled={!userInput.trim()}
                        className="flex-1 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-transform duration-150 hover:shadow-lg active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        Submit Answer
                      </button>
                      <button
                        onClick={handleSkipQuestion}
                        className="rounded-xl border-2 border-gray-300 px-6 py-3 font-semibold text-gray-600 transition-transform duration-150 hover:border-gray-400 active:scale-[0.98]"
                      >
                        Skip
                      </button>
                    </div>
                  )}

                  {showFeedback && (
                    <div className="hidden">
                      <button
                        onClick={handleNext}
                        className="flex items-center space-x-2 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-8 py-3 font-semibold text-white transition-transform duration-150 hover:shadow-lg active:scale-[0.98]"
                      >
                        <span>
                          {currentQuestionIndex < quizSet.length - 1 ? 'Next Question' : 'Complete Quiz'}
                        </span>
                        <i className="fas fa-arrow-right text-sm"></i>
                      </button>
                    </div>
                  )}
                </div>
              ) : (
                <div className="space-y-3">
                  {currentQuiz.options.map((option, index) => (
                    <button
                      key={index}
                      onClick={() => handleAnswerSelect(index)}
                      className={getOptionStyle(index)}
                      disabled={showFeedback}
                    >
                      <span className={getOptionTextStyle(index)}>{option}</span>

                      {showFeedback && index === currentQuiz.answerIndex && (
                        <i className="fas fa-check-circle ml-3 text-xl text-green-500"></i>
                      )}

                      {showFeedback &&
                        index === selectedAnswer &&
                        selectedAnswer !== currentQuiz.answerIndex && (
                          <i className="fas fa-times-circle ml-3 text-xl text-red-500"></i>
                        )}
                    </button>
                  ))}
                </div>
              )}

              {/* Action Buttons moved to sticky bottom bar for better mobile UX */}
            </div>
          ) : (
            // Loading or Error State
            <div className="space-y-4 text-center">
              <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600">
                <i className="fas fa-exclamation-triangle text-xl text-white"></i>
              </div>
              <div>
                <h3 className="mb-2 text-xl font-bold text-[#2D3436]">Demo Unavailable</h3>
                <p className="mb-4 text-[#636E72]">Quiz demo is temporarily unavailable</p>
                <button
                  onClick={onClose}
                  className="rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-all duration-300 hover:shadow-lg"
                >
                  Close
                </button>
              </div>
            </div>
          )}
        </div>

        <StickyActionBar
          isCompleted={isCompleted}
          isWriteMode={isWriteMode}
          showFeedback={showFeedback}
          canSubmitWrite={Boolean(userInput.trim())}
          onSubmitWrite={handleWriteAnswer}
          onSkip={handleSkipQuestion}
          onNext={handleNext}
        />

        {/* Reporting removed */}
      </div>
    </div>
  )
}

export default QuizDemoModal

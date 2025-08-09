import type { Metadata } from 'next'
import Image from 'next/image'
import { notFound } from 'next/navigation'
import { getPublicQuizBySlug, getPublicQuizQuestions, getPublicQuizLeaderboard } from '@/lib/api'
import QuizWithFooter from '@/components/quiz/QuizWithFooter'
import ShareQuizButton from '@/components/quiz/ShareQuizButton'
import PlayAgainButton from '@/components/quiz/PlayAgainButton'

type Props = {
  params: Promise<{ slug: string; locale: string }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  const quiz = await getPublicQuizBySlug(slug).catch(() => null)

  if (!quiz) return {}

  const title = `${quiz.title} — Decorebator Quiz`
  const description = quiz.description || `Play this ${quiz.difficulty} quiz. Can you beat it?`
  const canonical = `https://decorebator.com/q/${quiz.slug}`
  const master = quiz.previewImageUrl || ''
  const ogTwitter = master.replace('-master.jpg', '-twitter_x.jpg') 
  const ogSquare = master.replace('-master.jpg', '-square.jpg') 

  return {
    title,
    description,
    alternates: { canonical },
    openGraph: {
      title,
      description,
      url: canonical,
      type: 'article',
      siteName: 'Decorebator',
      images: (
        [
          // 1200x630 primary for Facebook/X/Telegram/WhatsApp
          { url: ogTwitter, width: 1200, height: 630, alt: `${quiz.title} — Decorebator quiz preview` },
          // 1:1 square variant for platforms preferring square
          { url: ogSquare, width: 1024, height: 1024, alt: `${quiz.title} — Decorebator quiz preview (square)` },
          // Include master as an additional hint (dimensions may vary)
          ...(master ? [{ url: master, alt: `${quiz.title} — Decorebator quiz preview (master)` }] : [])
        ] as { url: string; width?: number; height?: number; alt?: string }[]
      ),
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      site: '@decorebator',
      images: [ogTwitter],
    },
  }
}

export default async function PublicQuizPage({ params }: Props) {
  const { slug } = await params
  const quiz = await getPublicQuizBySlug(slug).catch(() => null)
  if (!quiz) notFound()

  // Fetch questions for this public quiz; if unavailable, show a friendly message instead of fallback
  const [quizzes, leaderboard] = await Promise.all([
    getPublicQuizQuestions(slug, { next: { revalidate: 300 } }).catch(() => []),
    getPublicQuizLeaderboard(slug, { next: { revalidate: 30 } }).catch(() => []),
  ])


  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-b from-orange-50 via-white to-white">
      {/* Decorative background blobs */}
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute -top-24 -left-24 h-64 w-64 rounded-full bg-orange-200/40 blur-3xl" />
        <div className="absolute top-1/3 -right-24 h-72 w-72 rounded-full bg-amber-100/60 blur-3xl" />
      </div>

      {/* Simple header */}
      <header className="relative z-10 mx-auto w-full max-w-6xl px-4 py-3 sm:py-5">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Image src="/quiz-icon.png" alt="Decorebator" width={36} height={36} className="h-9 w-9" />
            <span className="text-2xl sm:text-3xl font-extrabold tracking-tight text-[#2D3436]">Decorebator</span>
          </div>
        </div>
      </header>

      <main className="relative z-10 mx-auto max-w-3xl px-4 pt-3 pb-8 sm:pt-8 sm:pb-12">
        <div className="mb-8 rounded-2xl border border-orange-100 bg-white/70 p-6 shadow-sm backdrop-blur">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="space-y-2">
              <h1 className="bg-gradient-to-r from-[#FF7B54] to-orange-600 bg-clip-text text-3xl font-extrabold text-transparent sm:text-4xl">
                {quiz.title}
              </h1>
              {quiz.description ? (
                <p className="text-[#636E72]">{quiz.description}</p>
              ) : null}
              <div className="flex flex-wrap gap-2 pt-1 text-sm">
                <span className="inline-flex items-center rounded-full bg-orange-100 px-3 py-1 font-medium text-[#FF7B54]">
                  {quiz.difficulty.toUpperCase()}
                </span>
                <span className="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 font-medium text-gray-600">
                  {quiz.timeLimitMinutes} min
                </span>
                {quiz.creatorName ? (
                  <span className="inline-flex items-center rounded-full bg-gray-50 px-3 py-1 font-medium text-gray-600">
                    by {quiz.creatorName}
                  </span>
                ) : null}
              </div>
            </div>
            <div className="flex items-center gap-2">
              {quizzes.length > 0 ? (
                <PlayAgainButton quizzes={quizzes} slug={quiz.slug} title={quiz.title} />
              ) : null}
              <ShareQuizButton
                url={`https://decorebator.com/q/${quiz.slug}`}
                title={quiz.title}
                label="Share this quiz"
                showLabelOnMobile={true}
              />
            </div>
          </div>
          {/* Leaderboard (Top 10) */}
          {leaderboard.length > 0 && (
            <div id="leaderboard" className="mt-6">
              <h3 className="mb-3 text-sm font-semibold uppercase tracking-wide text-[#636E72]">🏆 Top Players</h3>
              <ol className="rounded-xl border border-gray-100 bg-white">
                {leaderboard.slice(0, 10).map((row, idx) => {
                  const rank = row.rank ?? idx + 1
                  const initial = (row.name || 'A').trim().charAt(0).toUpperCase()
                  const score = Math.max(0, Math.min(100, Math.round(row.scorePercentage)))
                  const timeSec = Math.max(0, Math.round(row.completionTimeSeconds))
                  const medalStyles = [
                    'bg-gradient-to-br from-yellow-100 to-amber-200 text-amber-700',
                    'bg-gradient-to-br from-gray-100 to-gray-200 text-gray-700',
                    'bg-gradient-to-br from-orange-100 to-amber-100 text-orange-700',
                  ]
                  const rankBadgeClass = rank <= 3 ? medalStyles[rank - 1] : 'bg-orange-50 text-[#FF7B54]'
                  return (
                    <li key={idx} className="group px-4 py-3 transition-colors hover:bg-orange-50/40">
                      <div className="flex items-center justify-between gap-4">
                        <div className="flex items-center gap-3">
                          <span
                            className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${rankBadgeClass} shadow-[inset_0_1px_0_rgba(255,255,255,0.6)]`}
                            aria-label={`Rank ${rank}`}
                            title={`Rank ${rank}`}
                          >
                            {rank <= 3 ? (rank === 1 ? '🥇' : rank === 2 ? '🥈' : '🥉') : rank}
                          </span>
                          <div className="flex items-center gap-2">
                            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-700">
                              {initial}
                            </div>
                            <span className="font-medium text-[#2D3436]">{row.name}</span>
                          </div>
                        </div>
                        <div className="w-40 sm:w-56">
                          <div className="flex items-center justify-between text-xs text-[#636E72]">
                            <span className="font-medium text-[#2D3436]">{score}%</span>
                            <span className="tabular-nums">⏱ {timeSec}s</span>
                          </div>
                          <div className="mt-1 h-2 w-full overflow-hidden rounded-full bg-gray-100">
                            <div
                              className="h-full rounded-full bg-gradient-to-r from-[#FF7B54] via-orange-500 to-orange-400 transition-[width] duration-300 ease-out"
                              style={{ width: `${score}%` }}
                              aria-hidden
                            />
                          </div>
                        </div>
                      </div>
                    </li>
                  )
                })}
              </ol>
            </div>
          )}
        </div>

        {quizzes.length > 0 ? (
          <QuizWithFooter quizzes={quizzes} slug={quiz.slug} title={quiz.title} />
        ) : (
          <div className="rounded-2xl border border-gray-200 bg-white p-6 text-center shadow-sm">
            <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-orange-100">
              <i className="fas fa-exclamation-triangle text-orange-500"></i>
            </div>
            <h2 className="mb-1 text-xl font-semibold text-[#2D3436]">Quiz unavailable</h2>
            <p className="text-[#636E72]">We couldn’t load this quiz right now. Please refresh the page and try again.</p>
          </div>
        )}
      </main>
    </div>
  )
}



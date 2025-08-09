export type PublicQuiz = {
  id: number
  slug: string
  title: string
  description?: string | null
  difficulty: 'easy' | 'medium' | 'hard'
  timeLimitMinutes: number
  wordlistId: number
  creatorName?: string | null
  previewImageUrl?:string|null
}

export type QuizQuestion = {
  value: string
  options: string[]
  answerIndex: number
  id: number
  type: string
  pos: string
  isVerbType: boolean
  pronunciation?: string
  imageDescription?: string
  audioURL?: string
  definitionId: number
  wordId: number
}

export type LeaderboardEntry = {
  rank: number
  name: string
  scorePercentage: number
  completionTimeSeconds: number
}

function getApiBaseUrl(): string {
  const envUrl = process.env.NEXT_PUBLIC_API_URL?.trim()
  if (envUrl && envUrl.startsWith('http')) return envUrl
  // Default to Go API local port
  return 'http://localhost:3000'
}

export async function getPublicQuizBySlug(slug: string, init?: RequestInit): Promise<PublicQuiz | null> {
  const base = getApiBaseUrl()
  const res = await fetch(`${base}/public-quizzes/${encodeURIComponent(slug)}`, {
    method: 'GET',
    // Ensure dynamic fetch to always get fresh quiz metadata
    cache: 'no-store',
    ...init,
  })

  if (res.status === 404) return null
  if (!res.ok) throw new Error(`Failed to load public quiz: ${res.status}`)

  const data = await res.json()

  const quiz: PublicQuiz = {
    id: data.id,
    slug: data.slug,
    title: data.title,
    description: data.description ?? null,
    difficulty: data.difficulty,
    timeLimitMinutes: data.timeLimitMinutes ?? data.time_limit_minutes ?? 10,
    wordlistId: data.wordlistId ?? data.wordlist_id,
    creatorName: data.creatorName ?? data.creator_name ?? null,
    previewImageUrl: data.previewImageUrl
  }

  return quiz
}

export async function getPublicQuizQuestions(slug: string, init?: RequestInit): Promise<QuizQuestion[]> {
  const base = getApiBaseUrl()
  const res = await fetch(`${base}/public-quizzes/${encodeURIComponent(slug)}/questions`, {
    method: 'GET',
    ...(init || {}),
  })

  if (res.status === 404) return []
  if (!res.ok) throw new Error(`Failed to load public quiz questions: ${res.status}`)

  const data = await res.json()
  const list: QuizQuestion[] = Array.isArray(data)
    ? data
    : Array.isArray(data.questions)
    ? data.questions
    : []
  return list
}

export async function getPublicQuizLeaderboard(slug: string, init?: RequestInit): Promise<LeaderboardEntry[]> {
  const base = getApiBaseUrl()
  const res = await fetch(`${base}/public-quizzes/${encodeURIComponent(slug)}/leaderboard`, {
    method: 'GET',
    ...(init || {}),
  })
  if (res.status === 404) return []
  if (!res.ok) throw new Error(`Failed to load leaderboard: ${res.status}`)
  const data = await res.json()

  // Expect camelCase response: { top: [{ guestName, scorePercentage, completionTimeSeconds, rank }] }
  const list = Array.isArray(data?.top) ? data.top : []
  if (!Array.isArray(list)) return []

  type ApiEntry = {
    rank?: number
    guestName?: string
    scorePercentage?: number
    completionTimeSeconds?: number
  }
  return (list as ApiEntry[]).map((e, idx: number) => {
    const nameRaw = e.guestName ?? ''
    const name = (typeof nameRaw === 'string' ? nameRaw : String(nameRaw || '')).trim() || 'Anonymous'
    return {
      rank: Number(e.rank ?? idx + 1),
      name,
      scorePercentage: Number(e.scorePercentage ?? 0),
      completionTimeSeconds: Number(e.completionTimeSeconds ?? 0),
    } as LeaderboardEntry
  })
}

export async function submitPublicQuizAttempt(args: {
  slug: string
  tried: number
  correct: number
  name?: string
  email?: string
  durationSeconds?: number
}): Promise<void> {
  const base = getApiBaseUrl()
  const res = await fetch(`${base}/public-quizzes/${encodeURIComponent(args.slug)}/attempts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      tried: args.tried,
      correct: args.correct,
      name: args.name,
      email: args.email,
      durationSeconds: args.durationSeconds,
    }),
  })
  if (!res.ok) {
    // Silently ignore for MVP
  }
}



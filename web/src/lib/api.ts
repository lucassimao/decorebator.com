export type PublicQuiz = {
  id: number
  slug: string
  title: string
  description?: string | null
  difficulty: 'easy' | 'medium' | 'hard'
  timeLimitMinutes: number
  wordlistId: number
  creatorName?: string | null
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

  // Map API response to PublicQuiz type
  const quiz: PublicQuiz = {
    id: data.id,
    slug: data.slug,
    title: data.title,
    description: data.description ?? null,
    difficulty: data.difficulty,
    timeLimitMinutes: data.timeLimitMinutes ?? data.time_limit_minutes ?? 10,
    wordlistId: data.wordlistId ?? data.wordlist_id,
    creatorName: data.creatorName ?? data.creator_name ?? null,
  }

  return quiz
}



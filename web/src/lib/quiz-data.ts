// Quiz data fetching and types for demo

export interface Quiz {
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

export interface QuizResponse {
  quizzes: Quiz[]
  total: number
}

// Fallback quiz data for when API is unavailable
const FALLBACK_QUIZ_DATA: Quiz[] = [
  {
    value: 'Serendipity',
    options: [
      'A pleasant surprise or fortunate discovery',
      'A planned celebration',
      'A difficult challenge',
      'A daily routine',
    ],
    answerIndex: 0,
    id: 1,
    type: 'GUESS_MEANING',
    pos: 'noun',
    isVerbType: false,
    pronunciation: '/ˌser.ənˈdɪp.ɪ.ti/',
    definitionId: 1,
    wordId: 1,
  },
  {
    value: 'Innovation',
    options: ['Tradition', 'Innovation', 'Repetition', 'Imitation'],
    answerIndex: 1,
    id: 2,
    type: 'WORD_FROM_IMAGE',
    pos: 'noun',
    isVerbType: false,
    imageDescription: 'A lightbulb with gears and technology symbols inside',
    definitionId: 2,
    wordId: 2,
  },
  {
    value: "The scientist's research was _______ by her peers.",
    options: ['ignored', 'acclaimed', 'questioned', 'forgotten'],
    answerIndex: 1,
    id: 3,
    type: 'COMPLETE_SENTENCE',
    pos: 'verb',
    isVerbType: true,
    definitionId: 3,
    wordId: 3,
  },
  {
    value: 'To speak about in an unfavorably critical manner',
    options: ['Praise', 'Disparage', 'Encourage', 'Support'],
    answerIndex: 1,
    id: 4,
    type: 'WRITE_WORD_FROM_DEFINITION',
    pos: 'verb',
    isVerbType: true,
    definitionId: 4,
    wordId: 4,
  },
  {
    value: 'A feeling of great happiness and triumph',
    options: ['Melancholy', 'Euphoria', 'Anxiety', 'Confusion'],
    answerIndex: 1,
    id: 5,
    type: 'WORD_FROM_MEANING',
    pos: 'noun',
    isVerbType: false,
    pronunciation: '/juːˈfɔː.ri.ə/',
    definitionId: 5,
    wordId: 5,
  },
]

export async function fetchDemoQuizzes(): Promise<Quiz[]> {
  try {
    const API_BASE = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:4000'
    const STATIC_AUTH = process.env.STATIC_AUTHENTICATION || '12345'

    console.log('Fetching demo quizzes from:', `${API_BASE}/quiz/demo`)

    const response = await fetch(`${API_BASE}/quiz/demo`, {
      headers: {
        Authorization: STATIC_AUTH,
        'Content-Type': 'application/json',
      },
      // Add timeout and error handling
      signal: AbortSignal.timeout(10000), // 10 second timeout
    })

    if (!response.ok) {
      console.warn(`Demo quiz API returned ${response.status}, using fallback data`)
      return FALLBACK_QUIZ_DATA
    }

    const data: QuizResponse = await response.json()
    console.log(`Fetched ${data.total} demo quizzes from API`)
    return data.quizzes
  } catch (error) {
    console.warn('Failed to fetch demo quizzes, using fallback data:', error)
    return FALLBACK_QUIZ_DATA
  }
}

export function getRandomQuizSet(allQuizzes: Quiz[], count: number = 5): Quiz[] {
  if (!allQuizzes || allQuizzes.length === 0) {
    return []
  }

  // Create a copy and shuffle
  const shuffled = [...allQuizzes].sort(() => Math.random() - 0.5)

  // Return the requested number of questions
  return shuffled.slice(0, Math.min(count, shuffled.length))
}

export function getQuizTypeDisplayName(type: string): string {
  const typeMap: { [key: string]: string } = {
    GUESS_MEANING: 'Guess Meaning',
    WORD_FROM_MEANING: 'Word from Meaning',
    WORD_FROM_IMAGE: 'Word from Image',
    WORD_FROM_AUDIO: 'Word from Audio',
    COMPLETE_SENTENCE: 'Complete Sentence',
    WRITE_WORD_FROM_DEFINITION: 'Write Word',
    WORD_FROM_EXAMPLE_AUDIO: 'Example Audio',
    MEANING_FROM_AUDIO: 'Meaning from Audio',
  }

  return typeMap[type] || type
}

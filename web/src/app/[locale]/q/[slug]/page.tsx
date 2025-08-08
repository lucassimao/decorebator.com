import type { Metadata, ResolvingMetadata } from 'next'
import { notFound } from 'next/navigation'
import { getPublicQuizBySlug } from '@/lib/api'

type Props = {
  params: Promise<{ slug: string; locale: string }>
}

export async function generateMetadata(
  { params }: Props,
  _parent: ResolvingMetadata,
): Promise<Metadata> {
  const { slug } = await params
  const quiz = await getPublicQuizBySlug(slug).catch(() => null)

  if (!quiz) return {}

  const title = `${quiz.title} — Public Quiz`
  const description = quiz.description || `Play this ${quiz.difficulty} quiz. Can you beat it?`
  const canonical = `https://decorebator.com/q/${quiz.slug}`

  return {
    title,
    description,
    alternates: { canonical },
    openGraph: {
      title,
      description,
      url: canonical,
      type: 'article',
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
    },
  }
}

export default async function PublicQuizPage({ params }: Props) {
  const { slug } = await params
  const quiz = await getPublicQuizBySlug(slug).catch(() => null)
  if (!quiz) notFound()

  return (
    <main className="mx-auto max-w-3xl px-4 py-10 sm:py-14">
      <div className="mb-8 flex items-center justify-between">
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
      </div>

      <div className="rounded-3xl border border-gray-100 bg-white p-6 shadow-lg sm:p-8">
        <div className="space-y-6 text-center">
          <p className="text-lg text-[#2D3436]">
            Ready to play this quiz? You can try it in the app—no signup required to start.
          </p>

          <div className="flex flex-col justify-center gap-3 sm:flex-row">
            <a
              href="https://decorebator.com/app"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-[#FF7B54] to-orange-600 px-6 py-3 font-semibold text-white transition-all duration-300 hover:shadow-lg"
            >
              <i className="fas fa-play"></i>
              Start in App
            </a>
            <a
              href="/"
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-gray-100 px-6 py-3 font-semibold text-[#2D3436] transition-all duration-300 hover:bg-gray-200"
            >
              <i className="fas fa-download"></i>
              Get the App
            </a>
          </div>
        </div>
      </div>
    </main>
  )
}



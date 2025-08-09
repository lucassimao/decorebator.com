import { ImageResponse } from 'next/og'

export const runtime = 'edge'

// Locale-prefixed alias of /og/quiz-default to avoid redirects breaking OG fetchers
export async function GET(request: Request) {
  const width = 1200
  const height = 630

  const background = 'linear-gradient(135deg, #FFFFFF 0%, #F7F8FA 100%)'
  const primary = '#FF7B54'
  const textColor = '#2D3436'
  const subText = '#636E72'

  // Load app icon from /public to embed into the OG image
  let iconData: ArrayBuffer | null = null
  try {
    const iconUrl = new URL('../../../../../public/quiz-icon.png', import.meta.url)
    const res = await fetch(iconUrl)
    if (res.ok) iconData = await res.arrayBuffer()
  } catch {
    // ignore; will fall back to glyph
  }

  // Try to fetch quiz metadata when slug is provided (?slug=...)
  let quizTitle = 'Decorebator Quiz'
  let quizDifficulty = ''
  try {
    const { searchParams } = new URL(request.url)
    const slug = searchParams.get('slug') || ''
    if (slug) {
      const baseEnv = (process.env.NEXT_PUBLIC_API_URL || '').trim()
      const base = baseEnv && baseEnv.startsWith('http') ? baseEnv : 'http://localhost:3000'
      const res = await fetch(`${base}/public-quizzes/${encodeURIComponent(slug)}`, { cache: 'no-store' })
      if (res.ok) {
        const data = await res.json()
        if (data?.title) quizTitle = String(data.title)
        if (data?.difficulty) quizDifficulty = String(data.difficulty).toUpperCase()
      }
    }
  } catch {
    // ignore
  }

  return new ImageResponse(
    (
      <div style={{ width: '100%', height: '100%', display: 'flex' }}>
        <div style={{
          width: '100%', height: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', background
        }}>
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', padding: 24, borderRadius: 32, background: 'rgba(255,255,255,0.96)', border: '1px solid #E5E7EB', boxShadow: '0 16px 40px rgba(0,0,0,0.08)' }}>
            {iconData ? (
              <div
                style={{
                  display: 'flex',
                  padding: 14,
                  borderRadius: 28,
                  background: 'rgba(255,255,255,1)',
                  border: '1px solid #E5E7EB',
                  boxShadow: '0 10px 28px rgba(0,0,0,0.08)'
                }}
              >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={URL.createObjectURL(new Blob([iconData!]))}
                  width={96}
                  height={96}
                  style={{ borderRadius: 22, display: 'block' }}
                  alt="Decorebator"
                />
              </div>
            ) : (
              <div
                style={{
                  width: 90,
                  height: 90,
                  background: '#fff',
                  borderRadius: 22,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: primary,
                  fontSize: 48,
                  fontWeight: 800,
                  boxShadow: '0 10px 28px rgba(0,0,0,0.08)',
                  border: '1px solid #E5E7EB'
                }}
              >
                ?
              </div>
            )}

            <div style={{ display: 'flex', marginTop: 16, fontSize: 60, fontWeight: 900, letterSpacing: -1.0, color: textColor, textAlign: 'center' }}>{quizTitle}</div>
            <div style={{ display: 'flex', marginTop: 8, fontSize: 26, fontWeight: 600, color: subText }}>
              {quizDifficulty ? `${quizDifficulty} • Decorebator` : 'Decorebator'}
            </div>
            <div style={{ display: 'flex', marginTop: 6, fontSize: 24, fontWeight: 500, color: subText }}>
              Learn faster with beautiful, shareable quizzes
            </div>
          </div>
        </div>
      </div>
    ),
    { width, height },
  )
}



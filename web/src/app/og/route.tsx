import { ImageResponse } from 'next/og'
import { NextRequest } from 'next/server'

export const runtime = 'edge'

const taglines: Record<string, string> = {
  en: 'AI-Powered Vocabulary Learning',
  es: 'Aprendizaje de Vocabulario con IA',
  fr: "Apprentissage du Vocabulaire par l'IA",
  de: 'KI-gestütztes Vokabellernen',
  it: "Apprendimento del Vocabolario con l'IA",
  pt: 'Aprendizado de Vocabulário com IA',
  ja: 'AI搭載の語彙学習',
}

const subtitles: Record<string, string> = {
  en: 'Master any language with spaced repetition & 9 quiz modes',
  es: 'Domina cualquier idioma con repetición espaciada y 9 modos de quiz',
  fr: 'Maîtrisez toute langue avec la répétition espacée et 9 modes de quiz',
  de: 'Beherrsche jede Sprache mit Spaced Repetition & 9 Quiz-Modi',
  it: 'Padroneggia qualsiasi lingua con la ripetizione dilazionata e 9 modalità quiz',
  pt: 'Domine qualquer idioma com repetição espaçada e 9 modos de quiz',
  ja: '間隔反復と9つのクイズモードで言語をマスター',
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const locale = searchParams.get('locale') || 'en'
  const page = searchParams.get('page') || 'home'

  const tagline = taglines[locale] || taglines.en
  const subtitle = subtitles[locale] || subtitles.en

  // Page-specific titles
  const pageTitles: Record<string, Record<string, string>> = {
    help: {
      en: 'Help Center',
      es: 'Centro de Ayuda',
      fr: "Centre d'Aide",
      de: 'Hilfe-Center',
      it: 'Centro Assistenza',
      pt: 'Central de Ajuda',
      ja: 'ヘルプセンター',
    },
    blog: {
      en: 'Blog',
      es: 'Blog',
      fr: 'Blog',
      de: 'Blog',
      it: 'Blog',
      pt: 'Blog',
      ja: 'ブログ',
    },
    privacy: {
      en: 'Privacy Policy',
      es: 'Política de Privacidad',
      fr: 'Politique de Confidentialité',
      de: 'Datenschutzrichtlinie',
      it: 'Informativa sulla Privacy',
      pt: 'Política de Privacidade',
      ja: 'プライバシーポリシー',
    },
    terms: {
      en: 'Terms of Service',
      es: 'Términos de Servicio',
      fr: "Conditions d'Utilisation",
      de: 'Nutzungsbedingungen',
      it: 'Termini di Servizio',
      pt: 'Termos de Serviço',
      ja: '利用規約',
    },
  }

  const pageTitle = page !== 'home' ? pageTitles[page]?.[locale] || pageTitles[page]?.en : null

  return new ImageResponse(
    (
      <div
        style={{
          height: '100%',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #FDF6E3 0%, #FFF8E7 50%, #FFFBF0 100%)',
          fontFamily: 'system-ui, -apple-system, sans-serif',
        }}
      >
        {/* Decorative circles */}
        <div
          style={{
            position: 'absolute',
            top: -100,
            right: -100,
            width: 400,
            height: 400,
            borderRadius: '50%',
            background:
              'linear-gradient(135deg, rgba(255, 123, 84, 0.2) 0%, rgba(255, 215, 0, 0.1) 100%)',
          }}
        />
        <div
          style={{
            position: 'absolute',
            bottom: -150,
            left: -150,
            width: 500,
            height: 500,
            borderRadius: '50%',
            background:
              'linear-gradient(135deg, rgba(255, 215, 0, 0.15) 0%, rgba(255, 123, 84, 0.1) 100%)',
          }}
        />

        {/* Logo */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: 120,
            height: 120,
            borderRadius: 24,
            background: 'linear-gradient(135deg, #FF7B54 0%, #FF9F54 100%)',
            marginBottom: 32,
            boxShadow: '0 20px 40px rgba(255, 123, 84, 0.3)',
          }}
        >
          <span style={{ fontSize: 64, fontWeight: 800, color: 'white' }}>D</span>
        </div>

        {/* App name */}
        <h1
          style={{
            fontSize: 72,
            fontWeight: 800,
            background: 'linear-gradient(135deg, #FF7B54 0%, #E85A30 100%)',
            backgroundClip: 'text',
            color: 'transparent',
            margin: 0,
            marginBottom: 16,
          }}
        >
          Decorebator
        </h1>

        {/* Page title or tagline */}
        {pageTitle ? (
          <p
            style={{
              fontSize: 36,
              fontWeight: 600,
              color: '#2D3436',
              margin: 0,
              marginBottom: 16,
            }}
          >
            {pageTitle}
          </p>
        ) : (
          <p
            style={{
              fontSize: 36,
              fontWeight: 600,
              color: '#2D3436',
              margin: 0,
              marginBottom: 16,
            }}
          >
            {tagline}
          </p>
        )}

        {/* Subtitle (only for home page) */}
        {page === 'home' && (
          <p
            style={{
              fontSize: 24,
              fontWeight: 400,
              color: '#636E72',
              margin: 0,
              maxWidth: 800,
              textAlign: 'center',
            }}
          >
            {subtitle}
          </p>
        )}

        {/* Language flags */}
        <div
          style={{
            display: 'flex',
            gap: 16,
            marginTop: 48,
            fontSize: 32,
          }}
        >
          <span>🇺🇸</span>
          <span>🇪🇸</span>
          <span>🇫🇷</span>
          <span>🇩🇪</span>
          <span>🇮🇹</span>
          <span>🇧🇷</span>
          <span>🇯🇵</span>
        </div>
      </div>
    ),
    {
      width: 1200,
      height: 630,
    }
  )
}

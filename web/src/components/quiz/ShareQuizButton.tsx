"use client"

import React, { useCallback, useEffect, useMemo, useState } from 'react'

type Props = {
  url: string
  title: string
  text?: string
  className?: string
  label?: string
  showLabelOnMobile?: boolean
}

export default function ShareQuizButton({ url, title, text, className, label, showLabelOnMobile = false }: Props) {
  const [copied, setCopied] = useState(false)
  const [supportsShare, setSupportsShare] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    setSupportsShare(typeof navigator !== 'undefined' && typeof (navigator as Navigator & { share?: unknown }).share === 'function')
  }, [])

  const shareText = useMemo(
    () => text || `I’m playing “${title}” on Decorebator — can you beat my score?`,
    [text, title],
  )

  const track = useCallback(
    (name: string, payload?: Record<string, unknown>) => {
      try {
        if (typeof window !== 'undefined') {
          window.dispatchEvent(
            new CustomEvent('decorebator:analytics', {
              detail: { name, payload: { url, title, ...payload }, ts: Date.now() },
            }),
          )
          // @ts-expect-error gtag may not be defined in all environments
          if (window.gtag) {
            // @ts-expect-error gtag may not be defined in all environments
            window.gtag('event', name, { event_category: 'share', value: url, ...payload })
          }
        }
      } catch {
        // no-op
      }
    },
    [title, url],
  )

  const copyToClipboard = useCallback(async () => {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(url)
      } else {
        const ta = document.createElement('textarea')
        ta.value = url
        ta.style.position = 'fixed'
        ta.style.left = '-9999px'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      }
      setCopied(true)
      track('share_copy')
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // ignore
    }
  }, [track, url])

  const onShare = useCallback(async () => {
    try {
      const nav = navigator as Navigator & { share?: (data: { title: string; text: string; url: string }) => Promise<void> }
      if (supportsShare && typeof nav.share === 'function') {
        await nav.share({ title, text: shareText, url })
        track('share_native')
      } else {
        setMenuOpen((prev) => !prev)
      }
    } catch {
      // user cancelled or share not available
    }
  }, [shareText, supportsShare, title, track, url])

  const openShareUrl = useCallback(
    (shareUrl: string, network: 'twitter' | 'facebook' | 'whatsapp' | 'copy') => {
      try {
        window.open(shareUrl, '_blank', 'noopener,noreferrer')
        track(`share_${network}`)
      } catch {
        // ignore
      }
    },
    [track],
  )

  const onShareTwitter = useCallback(() => {
    const u = `https://twitter.com/intent/tweet?text=${encodeURIComponent(shareText)}&url=${encodeURIComponent(url)}`
    openShareUrl(u, 'twitter')
  }, [openShareUrl, shareText, url])

  const onShareFacebook = useCallback(() => {
    const u = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`
    openShareUrl(u, 'facebook')
  }, [openShareUrl, url])

  const onShareWhatsApp = useCallback(() => {
    const u = `https://api.whatsapp.com/send?text=${encodeURIComponent(`${shareText} ${url}`)}`
    openShareUrl(u, 'whatsapp')
  }, [openShareUrl, shareText, url])

  return (
    <div className="relative inline-flex flex-col items-center">
      <button
        onClick={onShare}
        className={
          className ||
          'inline-flex items-center gap-2 rounded-xl border border-gray-200 px-3 py-2 text-sm font-semibold text-[#2D3436] hover:bg-gray-50'
        }
        aria-label={label ? label : 'Share this quiz'}
      >
        <i className="fas fa-share-alt text-[#FF7B54]"></i>
        <span className={showLabelOnMobile ? 'inline' : 'hidden sm:inline'}>
          {copied ? 'Link copied!' : label || 'Share'}
        </span>
      </button>

      {!supportsShare && menuOpen && (
        <div className="z-50 mt-2 flex w-full flex-wrap items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white p-2 shadow-lg">
          <button onClick={onShareTwitter} className="inline-flex items-center gap-2 rounded-lg bg-[#1DA1F2]/10 px-3 py-2 text-xs font-semibold text-[#1DA1F2] hover:bg-[#1DA1F2]/20">
            <i className="fab fa-twitter"></i>
            <span>Twitter</span>
          </button>
          <button onClick={onShareFacebook} className="inline-flex items-center gap-2 rounded-lg bg-[#1877F2]/10 px-3 py-2 text-xs font-semibold text-[#1877F2] hover:bg-[#1877F2]/20">
            <i className="fab fa-facebook"></i>
            <span>Facebook</span>
          </button>
          <button onClick={onShareWhatsApp} className="inline-flex items-center gap-2 rounded-lg bg-[#25D366]/10 px-3 py-2 text-xs font-semibold text-[#25D366] hover:bg-[#25D366]/20">
            <i className="fab fa-whatsapp"></i>
            <span>WhatsApp</span>
          </button>
          <button onClick={copyToClipboard} className="inline-flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-2 text-xs font-semibold text-[#2D3436] hover:bg-gray-50">
            <i className="far fa-copy"></i>
            <span>{copied ? 'Link copied!' : 'Copy link'}</span>
          </button>
        </div>
      )}
    </div>
  )
}



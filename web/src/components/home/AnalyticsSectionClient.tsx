'use client'

import React, { useSyncExternalStore } from 'react'

// Direct import to bypass dynamic import caching
import AnalyticsSection from './AnalyticsSection'

const AnalyticsSectionClient: React.FC = () => {
  const mounted = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false
  )

  if (!mounted) {
    return (
      <section className="bg-gradient-to-b from-slate-50 via-white to-slate-50 py-16 sm:py-20 lg:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="h-[600px] animate-pulse rounded-3xl bg-slate-100" />
        </div>
      </section>
    )
  }

  return <AnalyticsSection />
}

export default AnalyticsSectionClient

'use client'

import React from 'react'
import dynamic from 'next/dynamic'

const AnalyticsSection = dynamic(() => import('./AnalyticsSection'), {
  ssr: false,
  loading: () => (
    <section className="bg-white py-16 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="h-96 animate-pulse rounded-3xl bg-slate-100" />
      </div>
    </section>
  ),
})

const AnalyticsSectionClient: React.FC = () => {
  return <AnalyticsSection />
}

export default AnalyticsSectionClient

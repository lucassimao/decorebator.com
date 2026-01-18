import React from 'react'
import { setRequestLocale } from 'next-intl/server'
import PageLayout from '../../components/layout/PageLayout'
import EnhancedHeroSection from '../../components/home/EnhancedHeroSection'
import SocialProofSection from '../../components/home/SocialProofSection'
import NewFeaturesSection from '../../components/home/NewFeaturesSection'
import AppShowcaseSection from '../../components/home/AppShowcaseSection'
import HowItWorksSection from '../../components/home/HowItWorksSection'
import AnalyticsSectionClient from '../../components/home/AnalyticsSectionClient'
import PricingSection from '../../components/home/PricingSection'
import FAQSection from '../../components/home/FAQSection'
import CTASection from '../../components/home/CTASection'

interface HomePageProps {
  params: Promise<{
    locale: string
  }>
}

const HomePage: React.FC<HomePageProps> = async ({ params }) => {
  const { locale } = await params

  // Set the locale for the request - this is crucial for server-side rendering
  setRequestLocale(locale)

  return (
    <PageLayout>
      <EnhancedHeroSection />
      <SocialProofSection />
      <NewFeaturesSection />
      <AppShowcaseSection />
      <HowItWorksSection />
      <AnalyticsSectionClient />
      <PricingSection />
      <FAQSection />
      <CTASection />
    </PageLayout>
  )
}

export default HomePage

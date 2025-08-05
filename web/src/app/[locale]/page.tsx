import React from 'react'
import { setRequestLocale } from 'next-intl/server'
import PageLayout from '../../components/layout/PageLayout'
import EnhancedHeroSection from '../../components/home/EnhancedHeroSection'
import NewFeaturesSection from '../../components/home/NewFeaturesSection'
import AppShowcaseSection from '../../components/home/AppShowcaseSection'
import HowItWorksSection from '../../components/home/HowItWorksSection'
import AnalyticsSection from '../../components/home/AnalyticsSection'
import PricingSection from '../../components/home/PricingSection'
import FAQSection from '../../components/home/FAQSection'
import CTASection from '../../components/home/CTASection'
import { fetchDemoQuizzes } from '../../lib/quiz-data'

interface HomePageProps {
  params: Promise<{
    locale: string
  }>
}

const HomePage: React.FC<HomePageProps> = async ({ params }) => {
  const { locale } = await params

  // Set the locale for the request - this is crucial for server-side rendering
  setRequestLocale(locale)

  // Fetch quiz data at build time
  const demoQuizzes = await fetchDemoQuizzes()

  return (
    <PageLayout>
      {/* Animated Background Elements */}
      <div className="pointer-events-none fixed inset-0 overflow-hidden">
        <div className="float-animation absolute top-20 left-10 h-32 w-32 rounded-full bg-orange-300 opacity-3 blur-3xl"></div>
        <div
          className="float-animation absolute right-10 bottom-20 h-40 w-40 rounded-full bg-yellow-300 opacity-3 blur-3xl"
          style={{ animationDelay: '3s' }}
        ></div>
      </div>

      <EnhancedHeroSection demoQuizzes={demoQuizzes} />
      <NewFeaturesSection />
      <AppShowcaseSection />
      <HowItWorksSection />
      <AnalyticsSection />
      <PricingSection />
      {/* <TestimonialsSection /> */}
      <FAQSection />
      <CTASection />
    </PageLayout>
  )
}

export default HomePage

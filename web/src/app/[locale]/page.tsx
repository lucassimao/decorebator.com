import React from 'react';
import PageLayout from '../../components/layout/PageLayout';
import EnhancedHeroSection from '../../components/home/EnhancedHeroSection';
import NewFeaturesSection from '../../components/home/NewFeaturesSection';
import AppShowcaseSection from '../../components/home/AppShowcaseSection';
import HowItWorksSection from '../../components/home/HowItWorksSection';
import AnalyticsSection from '../../components/home/AnalyticsSection';
import PricingSection from '../../components/home/PricingSection';
import FAQSection from '../../components/home/FAQSection';
import CTASection from '../../components/home/CTASection';
import { fetchDemoQuizzes } from '../../lib/quiz-data';

const HomePage: React.FC = async () => {
  // Fetch quiz data at build time
  const demoQuizzes = await fetchDemoQuizzes();

  return (
    <PageLayout>
      {/* Animated Background Elements */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-20 left-10 w-32 h-32 bg-orange-300 rounded-full opacity-3 blur-3xl float-animation"></div>
        <div className="absolute bottom-20 right-10 w-40 h-40 bg-yellow-300 rounded-full opacity-3 blur-3xl float-animation" style={{animationDelay: '3s'}}></div>
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
  );
};

export default HomePage;
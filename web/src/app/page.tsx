import React from 'react';
import { Feature } from '../types';
import HeroSection from '../components/home/HeroSection';
import FeaturesSection from '../components/home/FeaturesSection';
import HowItWorksSection from '../components/home/HowItWorksSection';
import PricingSection from '../components/home/PricingSection';
import SocialProofSection from '../components/home/SocialProofSection';
import SubscriptionCalloutSection from '../components/home/SubscriptionCalloutSection';
import EmailCaptureForm from '../components/home/EmailCaptureForm';
import FooterSection from '../components/home/FooterSection';

import { 
  ListPlusIcon, SparklesIcon, WordMeaningIcon, 
  ImageWordIcon, AudioWordIcon, ChatBubbleBottomCenterTextIcon 
} from '../components/home/icons';

const HomePage: React.FC = () => {
  const features: Feature[] = [
    {
      icon: <ListPlusIcon />,
      title: 'Build Your Vocabulary Lists',
      description: 'Easily create and manage multiple wordlists for any language. Add words or expressions you want to master.',
    },
    {
      icon: <SparklesIcon />,
      title: 'AI-Powered Enrichment',
      description: 'Our AI automatically finds meanings, relevant images, example sentences, and pronunciation audio for your words.',
    },
    {
      icon: <WordMeaningIcon />,
      title: 'Word ↔ Meaning Mastery',
      description: 'Test your knowledge by matching words to definitions, or definitions back to words, with AI-adapted difficulty.',
    },
    {
      icon: <ImageWordIcon />,
      title: 'Visual Word Association',
      description: 'Strengthen recall by identifying words from AI-selected images, making learning more engaging.',
    },
    {
      icon: <AudioWordIcon />,
      title: 'Pronunciation & Listening',
      description: 'Master pronunciation and listening comprehension by guessing words from authentic AI-generated audio clips.',
    },
     {
      icon: <ChatBubbleBottomCenterTextIcon />,
      title: 'Contextual Learning',
      description: 'Learn words in context by completing sentences, with AI ensuring varied and relevant examples for deeper understanding.',
    },
  ];

  return (
    <div className="min-h-screen bg-slate-50 text-slate-800 font-sans selection:bg-blue-200 selection:text-blue-700">
      {/* Main content wrapper */}
      <div className="max-w-7xl mx-auto">
        <HeroSection />
        <FeaturesSection features={features} />
        <HowItWorksSection />
        <PricingSection />
        <SocialProofSection />
        <SubscriptionCalloutSection />
        <EmailCaptureForm />
      </div>
      <FooterSection />
    </div>
  );
};

export default HomePage;
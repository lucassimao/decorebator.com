import React from 'react';
import Button from './Button';
import { Logo, AppStoreIcon, GooglePlayIcon } from './icons';

const HeroSection: React.FC = () => {
  return (
    <section className="py-20 md:py-32 text-center px-4 sm:px-6 lg:px-8" aria-labelledby="hero-heading">
      <div className="mb-6">
        <Logo className="inline-block" />
      </div>
      <h1 id="hero-heading" className="text-4xl sm:text-5xl md:text-6xl font-bold text-slate-900 mb-4">
        Create Custom Wordlists. <span className="text-blue-600">Learn Any Language Smarter.</span>
      </h1>
      <p className="text-lg sm:text-xl text-slate-600 max-w-3xl mx-auto mb-4">
        Decorebator uses AI to automatically enrich your vocabulary lists with meanings, images, audio, and generates personalized quizzes to accelerate your learning.
      </p>
      <p className="text-md text-slate-500 max-w-3xl mx-auto mb-8">
        <em>Currently focused on English vocabulary, with support for more languages coming soon!</em>
      </p>
      <div className="space-y-4 sm:space-y-0 sm:space-x-4">
        <Button variant="primary" size="lg" >
          Get Started Free
        </Button>
        <Button variant="outline" size="lg" href="#download-app" className="hidden sm:inline-flex">
          Download App
        </Button>
      </div>
      <div id="download-app" className="mt-8 flex justify-center space-x-4">
        <Button 
          href="#" 
          variant="secondary" 
          size="md" 
          className="shadow-md"
          icon={<AppStoreIcon className="h-5 w-5" />}
        >
          App Store
        </Button>
        <Button 
          href="#" 
          variant="secondary" 
          size="md" 
          className="shadow-md"
          icon={<GooglePlayIcon className="h-5 w-5" />}
        >
          Google Play
        </Button>
      </div>
    </section>
  );
};

export default HeroSection;
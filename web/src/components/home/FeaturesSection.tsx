import React from 'react';
import FeatureCard from './FeatureCard';
import { Feature } from '../../types';

interface FeaturesSectionProps {
  features: Feature[];
}

const FeaturesSection: React.FC<FeaturesSectionProps> = ({ features }) => {
  return (
    <section id="features" className="py-16 px-4 sm:px-6 lg:px-8" aria-labelledby="features-heading">
      <h2 id="features-heading" className="text-3xl font-bold text-slate-900 text-center mb-4">AI-Supercharged Vocabulary Learning</h2>
      <p className="text-lg text-slate-600 text-center max-w-2xl mx-auto mb-12">
        From creating lists to mastering words with diverse, AI-generated tests, Decorebator is your smart companion.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 md:gap-8">
        {features.map((feature) => (
          <FeatureCard key={feature.title} feature={feature} />
        ))}
      </div>
      {/* 
        Placeholder for swapping in real screenshots:
        Replace the FeatureCard component or add images within it.
        Example: <img src="https://picsum.photos/300/200?random=1" alt={feature.title} className="rounded-lg mb-4" />
      */}
    </section>
  );
};

export default FeaturesSection;
import React from 'react';
import Button from './Button';
import { ArrowRightIcon } from './icons';

const SubscriptionCalloutSection: React.FC = () => {
  return (
    <section id="subscription" className="py-16 my-12 bg-gradient-to-r from-blue-600 to-indigo-600 text-white rounded-2xl shadow-xl mx-4 sm:mx-6 lg:mx-8 px-4 sm:px-6 lg:px-12 text-center" aria-labelledby="subscription-heading">
      <h2 id="subscription-heading" className="text-3xl font-bold mb-4">Unlock Your Full Potential</h2>
      <p className="text-lg opacity-90 max-w-xl mx-auto mb-8">
        Start your 7-day free trial of Decorebator Premium. Access all AI enrichment tools, unlimited wordlist creation, all quiz modes, personalized AI feedback, and more.
      </p>
      <Button variant="secondary" size="lg"  className="bg-white text-blue-600 hover:bg-blue-50">
        Start Free Trial <ArrowRightIcon aria-hidden="true" className="ml-2 h-5 w-5" />
      </Button>
    </section>
  );
};

export default SubscriptionCalloutSection;
import React from 'react';
import Button from './Button';
import { CheckCircleIcon } from './icons';

const PricingSection: React.FC = () => {
  const plans = [
    {
      name: 'Monthly Plan',
      price: '$12',
      frequency: '/mo',
      features: ['AI Word Enrichment', 'All Quiz Modes', 'Unlimited Wordlists', 'Progress Tracking'],
      cta: 'Choose Monthly',
      bestValue: false,
    },
    {
      name: 'Annual Plan',
      price: '$120',
      frequency: '/yr',
      features: ['All Monthly Features', 'Priority Support', 'Early Access to New Features', 'Save 20% Annually'],
      cta: 'Choose Annual',
      bestValue: true,
    },
  ];

  return (
    <section id="pricing" className="py-16 px-4 sm:px-6 lg:px-8" aria-labelledby="pricing-heading">
      <h2 id="pricing-heading" className="text-3xl font-bold text-slate-900 text-center mb-4">
        Flexible Plans for Every Learner
      </h2>
      <p className="text-lg text-slate-600 text-center max-w-2xl mx-auto mb-12">
        Choose the plan that best fits your learning journey with Decorebator.
      </p>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
        {plans.map((plan) => (
          <div 
            key={plan.name} 
            className={`bg-white p-8 rounded-2xl shadow-xl border-2 ${plan.bestValue ? 'border-blue-600' : 'border-transparent'} relative flex flex-col`}
          >
            {plan.bestValue && (
              <span className="absolute top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-xs font-semibold px-3 py-1 rounded-full shadow-md">
                BEST VALUE
              </span>
            )}
            <h3 className="text-2xl font-semibold text-slate-800 mb-2 text-center">{plan.name}</h3>
            <p className="text-4xl font-bold text-slate-900 mb-1 text-center">
              {plan.price}
              <span className="text-lg font-normal text-slate-500">{plan.frequency}</span>
            </p>
            {plan.name === 'Annual Plan' && (
                 <p className="text-sm text-green-600 font-medium text-center mb-6">Save $24 compared to monthly!</p>
            )}
            {plan.name === 'Monthly Plan' && (
                 <p className="text-sm text-transparent font-medium text-center mb-6">placeholder</p> // to maintain alignment
            )}

            <ul className="space-y-3 mb-8 flex-grow">
              {plan.features.map((feature) => (
                <li key={feature} className="flex items-center">
                  <CheckCircleIcon className="h-5 w-5 text-green-500 mr-2 flex-shrink-0" />
                  <span className="text-slate-600 text-sm">{feature}</span>
                </li>
              ))}
            </ul>
            <Button variant={plan.bestValue ? 'primary' : 'outline'} size="lg" className="w-full mt-auto" >
              {plan.cta}
            </Button>
          </div>
        ))}
      </div>
      <p className="text-center text-sm text-slate-500 mt-8">
        All plans come with a 7-day free trial. Cancel anytime.
      </p>
    </section>
  );
};

export default PricingSection;
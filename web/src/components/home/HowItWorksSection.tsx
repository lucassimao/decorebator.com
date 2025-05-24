import React from 'react';

const HowItWorksSection: React.FC = () => {
  const steps = [
    { num: 1, title: 'Create Your Wordlists', desc: 'Add any words or expressions in your target language to personalized lists.' },
    { num: 2, title: 'AI Auto-Enriches & Prepares', desc: 'Decorebator fetches meanings, images, audio, and crafts tailored quizzes for you.' },
    { num: 3, title: 'Practice & Master', desc: 'Engage with diverse quiz types and track your progress as your vocabulary expands.' },
  ];

  return (
    <section id="how-it-works" className="py-16 my-12 bg-white rounded-2xl shadow-xl mx-4 sm:mx-6 lg:mx-8 px-4 sm:px-6 lg:px-12" aria-labelledby="how-it-works-heading">
      <h2 id="how-it-works-heading" className="text-3xl font-bold text-slate-900 text-center mb-12">Simple Steps to Fluency</h2>
      <ol className="space-y-10 md:space-y-0 md:flex md:justify-around md:space-x-8">
        {steps.map((step) => (
          <li key={step.num} className="flex items-start md:flex-col md:items-center md:text-center max-w-xs mx-auto">
            <span className="flex-shrink-0 h-12 w-12 bg-blue-600 text-white rounded-full flex items-center justify-center text-xl font-bold shadow-md mr-4 md:mr-0 md:mb-4" aria-hidden="true">
              {step.num}
            </span>
            <div>
              <h4 className="text-xl font-semibold text-slate-800 mb-1">{step.title}</h4>
              <p className="text-slate-600 text-sm">{step.desc}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
};

export default HowItWorksSection;
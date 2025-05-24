
import React from 'react';
import { Feature } from '../../types';

interface FeatureCardProps {
  feature: Feature;
}

const FeatureCard: React.FC<FeatureCardProps> = ({ feature }) => {
  return (
    <div className="bg-white p-6 rounded-2xl shadow-lg flex flex-col items-center text-center h-full min-w-[280px] sm:min-w-0">
      <div className="bg-blue-100 text-blue-600 p-4 rounded-full mb-4">
        {React.cloneElement(feature.icon, { className: 'h-8 w-8' })}
      </div>
      <h3 className="text-xl font-semibold text-slate-800 mb-2">{feature.title}</h3>
      <p className="text-slate-600 text-sm leading-relaxed">{feature.description}</p>
    </div>
  );
};

export default FeatureCard;

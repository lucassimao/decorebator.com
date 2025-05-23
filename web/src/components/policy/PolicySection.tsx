
import React from 'react';

interface PolicySectionProps {
  title: string;
  children: React.ReactNode;
}

const PolicySection: React.FC<PolicySectionProps> = ({ title, children }) => {
  return (
    <section className="mb-8 py-4 border-b border-slate-200 last-of-type:border-b-0">
      <h2 className="text-2xl font-semibold text-slate-800 mb-4">{title}</h2>
      <div className="prose prose-slate max-w-none text-slate-700">
        {children}
      </div>
    </section>
  );
};

export default PolicySection;

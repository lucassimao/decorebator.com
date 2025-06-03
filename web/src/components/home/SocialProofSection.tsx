import Image from 'next/image';
import React from 'react';

const SocialProofSection: React.FC = () => {
  return (
    <section id="social-proof" className="py-16 text-center px-4 sm:px-6 lg:px-8" aria-labelledby="social-proof-heading">
      <div className="max-w-md mx-auto">
        <p id="social-proof-heading" className="text-5xl font-bold text-blue-600 mb-2">10,000+</p>
        <p className="text-xl text-slate-700 font-medium mb-8">Words Mastered by Our Community</p>
        <div className="bg-white p-6 rounded-2xl shadow-lg">
          <Image 
            src="https://avatar.iran.liara.run/public/boy?username=LucasSimao" // ** REPLACE THIS with your actual image URL or path **
            alt="Lucas Simão, Creator of Decorebator" 
            className="w-24 h-24 rounded-full mx-auto mb-4 shadow-md" 
          />
          <blockquote className="text-slate-600 italic">
            <p>&quot;As a voracious reader, I often stumbled upon new English words. I built Decorebator to solve my own vocabulary struggles, and then realized it could help so many others learn more effectively!&quot;</p>
          </blockquote>
          <p className="mt-4 font-semibold text-slate-800">- Lucas Simão, Creator, CTO, Lead Engineer, Head of Support & Chief Coffee Maker</p>
          {/* Placeholder for user testimonial: Replace with real testimonials */}
        </div>
      </div>
    </section>
  );
};

export default SocialProofSection;
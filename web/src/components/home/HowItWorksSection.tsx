import React from 'react';

const HowItWorksSection: React.FC = () => {
  return (
    <section id="how-it-works" className="py-20 bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Your Learning Journey in 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> 3 Simple Steps</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            Our streamlined process makes vocabulary learning efficient, engaging, and effective.
          </p>
        </div>

        <div className="grid lg:grid-cols-3 gap-8 relative">
          {/* Step 1 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 w-12 h-12 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center text-white font-bold text-xl z-10">1</div>
            <div className="bg-gradient-to-br from-orange-50 to-amber-50 p-8 rounded-2xl border border-orange-100 h-full transform hover:scale-105 transition-transform duration-300">
              <div className="w-16 h-16 bg-white rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                <i className="fas fa-plus-circle text-[#FF7B54] text-3xl"></i>
              </div>
              <h3 className="text-2xl font-bold mb-4">Create Your Wordlist</h3>
              <p className="text-[#636E72] mb-6">
                Simply add words you want to learn. Our AI instantly generates comprehensive learning materials.
              </p>
              <ul className="space-y-3">
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">AI-generated definitions & examples</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">AI visual associations</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">High-quality audio pronunciations</span>
                </li>
              </ul>
            </div>
            {/* Connector */}
            <div className="hidden lg:block absolute top-1/2 -right-4 transform -translate-y-1/2 z-0">
              <i className="fas fa-arrow-right text-4xl text-orange-300"></i>
            </div>
          </div>

          {/* Step 2 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 w-12 h-12 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-full flex items-center justify-center text-white font-bold text-xl z-10">2</div>
            <div className="bg-gradient-to-br from-purple-50 to-pink-50 p-8 rounded-2xl border border-purple-100 h-full transform hover:scale-105 transition-transform duration-300">
              <div className="w-16 h-16 bg-white rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                <i className="fas fa-gamepad text-[#9C27B0] text-3xl"></i>
              </div>
              <h3 className="text-2xl font-bold mb-4">Take Engaging Quizzes</h3>
              <p className="text-[#636E72] mb-6">
                Practice with 8 different quiz modes that adapt to your learning progress.
              </p>
              <ul className="space-y-3">
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Meaning ↔ Word matching</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Image & audio recognition</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Sentence completion</span>
                </li>
              </ul>
            </div>
            {/* Connector */}
            <div className="hidden lg:block absolute top-1/2 -right-4 transform -translate-y-1/2 z-0">
              <i className="fas fa-arrow-right text-4xl text-purple-300"></i>
            </div>
          </div>

          {/* Step 3 */}
          <div className="relative">
            <div className="absolute -top-4 -left-4 w-12 h-12 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-full flex items-center justify-center text-white font-bold text-xl z-10">3</div>
            <div className="bg-gradient-to-br from-green-50 to-emerald-50 p-8 rounded-2xl border border-green-100 h-full transform hover:scale-105 transition-transform duration-300">
              <div className="w-16 h-16 bg-white rounded-2xl flex items-center justify-center mb-6 shadow-lg">
                <i className="fas fa-brain text-[#4CAF50] text-3xl"></i>
              </div>
              <h3 className="text-2xl font-bold mb-4">Master with Repetition</h3>
              <p className="text-[#636E72] mb-6">
                Our 7-box Leitner system schedules reviews at optimal intervals for maximum retention.
              </p>
              <ul className="space-y-3">
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Adaptive review scheduling</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Progress tracking & insights</span>
                </li>
                <li className="flex items-start">
                  <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                  <span className="text-sm">Long-term retention focus</span>
                </li>
              </ul>
            </div>
          </div>
        </div>

        {/* CTA */}
        <div className="text-center mt-16">
          <a href="#download" className="group bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-10 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 inline-block">
            <span>Download App</span>
            <i className="fas fa-download ml-2 group-hover:translate-x-2 transition-transform"></i>
          </a>
        </div>
      </div>
    </section>
  );
};

export default HowItWorksSection;
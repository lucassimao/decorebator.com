import React from 'react';

const AppShowcaseSection: React.FC = () => {
  return (
    <section className="py-20 bg-gradient-to-br from-orange-50 via-amber-50 to-yellow-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Experience the 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent">Future of Learning</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            Our intuitive interface and smart features make vocabulary learning engaging, efficient, and enjoyable.
          </p>
        </div>

        {/* Interactive Demo */}
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Feature Tabs */}
          <div className="space-y-6">
            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-orange-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-chart-line text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">Smart Progress Tracking</h3>
                  <p className="text-[#636E72] mb-3">Monitor your learning journey with detailed analytics, streak counters, and achievement badges.</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-orange-100 text-orange-700 rounded-full text-sm font-medium">Learning Velocity</span>
                    <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm font-medium">Retention Rate</span>
                    <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">Daily Streaks</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-green-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-robot text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">Multi-Language AI Content</h3>
                  <p className="text-[#636E72] mb-3">AI generates native-language content with proper grammar, cultural context, and language-optimized voices.</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm font-medium">Native Definitions</span>
                    <span className="px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm font-medium">Cultural Images</span>
                    <span className="px-3 py-1 bg-purple-100 text-purple-700 rounded-full text-sm font-medium">Optimized Audio</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white/80 backdrop-blur rounded-2xl p-6 shadow-lg border border-purple-100 transform hover:scale-105 transition-transform duration-300 cursor-pointer">
              <div className="flex items-start space-x-4">
                <div className="w-12 h-12 bg-gradient-to-br from-[#9C27B0] to-purple-600 rounded-xl flex items-center justify-center flex-shrink-0">
                  <i className="fas fa-box text-white"></i>
                </div>
                <div>
                  <h3 className="text-xl font-bold mb-2">Advanced Leitner System</h3>
                  <p className="text-[#636E72] mb-3">7-box spaced repetition system that adapts to your learning pace and optimizes review timing.</p>
                  <div className="flex flex-wrap gap-2">
                    <span className="px-3 py-1 bg-red-100 text-red-700 rounded-full text-sm font-medium">Box 1-2: New</span>
                    <span className="px-3 py-1 bg-yellow-100 text-yellow-700 rounded-full text-sm font-medium">Box 3-5: Learning</span>
                    <span className="px-3 py-1 bg-green-100 text-green-700 rounded-full text-sm font-medium">Box 6-7: Mastered</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Phone Mockup with Screens */}
          <div className="relative flex justify-center">
            <div className="relative">
              {/* Phone Frame */}
              <div className="w-80 h-[640px] bg-gray-900 rounded-[3rem] p-3 shadow-2xl">
                <div className="w-full h-full bg-white rounded-[2.5rem] overflow-hidden">
                  {/* Status Bar */}
                  <div className="bg-white px-6 py-3 flex justify-between items-center text-xs">
                    <span className="font-medium">9:41</span>
                    <div className="flex space-x-1">
                      <i className="fas fa-signal"></i>
                      <i className="fas fa-wifi"></i>
                      <i className="fas fa-battery-full"></i>
                    </div>
                  </div>
                  
                  {/* Dashboard Screen */}
                  <div className="p-6 h-full bg-gradient-to-b from-orange-50 to-amber-50">
                    <h2 className="text-2xl font-bold mb-6">Dashboard</h2>
                    
                    {/* Stats Grid */}
                    <div className="grid grid-cols-2 gap-4 mb-6">
                      <div className="bg-white rounded-2xl p-4 shadow-sm">
                        <div className="text-3xl font-bold text-[#FF7B54]">127</div>
                        <div className="text-sm text-[#636E72]">Words Learned</div>
                      </div>
                      <div className="bg-white rounded-2xl p-4 shadow-sm">
                        <div className="text-3xl font-bold text-[#4CAF50]">15</div>
                        <div className="text-sm text-[#636E72]">Day Streak 🔥</div>
                      </div>
                    </div>
                    
                    {/* Progress Card */}
                    <div className="bg-white rounded-2xl p-4 shadow-sm mb-4">
                      <div className="flex justify-between items-center mb-3">
                        <h3 className="font-semibold">Spanish Vocabulary</h3>
                        <span className="text-sm text-[#FF7B54] font-medium">78%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-3">
                        <div className="bg-gradient-to-r from-[#FF7B54] to-orange-600 rounded-full h-3 w-3/4"></div>
                      </div>
                    </div>
                    
                    {/* Review Button */}
                    <button className="w-full bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white py-4 rounded-2xl font-semibold flex items-center justify-center space-x-2">
                      <i className="fas fa-play"></i>
                      <span>Start Today&apos;s Review (12 words)</span>
                    </button>
                  </div>
                </div>
              </div>
              
              {/* Floating UI Elements */}
              <div className="absolute -top-8 -right-8 bg-white rounded-2xl p-4 shadow-xl transform rotate-12 hover:rotate-0 transition-transform duration-300">
                <div className="flex items-center space-x-2">
                  <div className="w-10 h-10 bg-[#FFD700] rounded-full flex items-center justify-center">
                    <i className="fas fa-trophy text-white"></i>
                  </div>
                  <div>
                    <div className="text-sm font-bold">Achievement Unlocked!</div>
                    <div className="text-xs text-[#636E72]">100 Words Mastered</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default AppShowcaseSection;
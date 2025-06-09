'use client';

import React, { useState, useEffect } from 'react';

const AnalyticsSection: React.FC = () => {
  const [animatedValues, setAnimatedValues] = useState({
    totalWords: 0,
    wordsMastered: 0,
    averageAccuracy: 0,
    streakDays: 0,
    responseTime: 0,
  });

  // Animate numbers on mount
  useEffect(() => {
    const timer = setTimeout(() => {
      setAnimatedValues({
        totalWords: 2847,
        wordsMastered: 1923,
        averageAccuracy: 89,
        streakDays: 42,
        responseTime: 1.8,
      });
    }, 500);

    return () => clearTimeout(timer);
  }, []);

  const metrics = [
    {
      icon: <i className="fas fa-brain text-3xl"></i>,
      title: 'Word Mastery Tracking',
      description: 'Individual word progress with mastery levels calculated from Leitner box progression and accuracy rates.',
      details: ['Mastery percentage (0-100%)', 'Streak counting', 'Box level progression (1-7)', 'Historical accuracy'],
    },
    {
      icon: <i className="fas fa-chart-line text-3xl"></i>,
      title: 'Learning Progress Analytics',
      description: 'Daily aggregated statistics showing your learning journey over time.',
      details: ['Words studied per day', 'New words added', 'Study time tracking', 'Response time analysis'],
    },
    {
      icon: <i className="fas fa-target text-3xl"></i>,
      title: 'Quiz Performance Insights',
      description: 'Detailed performance metrics across all 7 quiz types with success rates and timing.',
      details: ['Performance by quiz type', 'Average response times', 'Difficulty progression', 'Accuracy trends'],
    },
    {
      icon: <i className="fas fa-layer-group text-3xl"></i>,
      title: 'Leitner System Tracking',
      description: 'Advanced spaced repetition analytics showing word distribution across difficulty boxes.',
      details: ['Box distribution snapshots', 'Retention analysis', 'Optimal review timing', 'Memory consolidation'],
    },
  ];

  const dashboardStats = [
    {
      label: 'Total Words',
      value: animatedValues.totalWords,
      suffix: '',
      icon: <i className="fas fa-book text-[#FF7B54]"></i>,
      description: 'Across all wordlists',
    },
    {
      label: 'Words Mastered',
      value: animatedValues.wordsMastered,
      suffix: '',
      icon: <i className="fas fa-trophy text-[#FFD700]"></i>,
      description: '80%+ mastery level',
    },
    {
      label: 'Average Accuracy',
      value: animatedValues.averageAccuracy,
      suffix: '%',
      icon: <i className="fas fa-bullseye text-[#4CAF50]"></i>,
      description: 'Overall success rate',
    },
    {
      label: 'Current Streak',
      value: animatedValues.streakDays,
      suffix: ' days',
      icon: <i className="fas fa-fire text-[#FF7B54]"></i>,
      description: 'Daily study streak',
    },
  ];

  const quizTypes = [
    { name: 'Guess Meaning', accuracy: 94, attempts: 456 },
    { name: 'Word from Image', accuracy: 87, attempts: 234 },
    { name: 'Audio Recognition', accuracy: 82, attempts: 189 },
    { name: 'Complete Sentence', accuracy: 91, attempts: 312 },
    { name: 'Write from Definition', accuracy: 78, attempts: 145 },
    { name: 'Word from Meaning', accuracy: 88, attempts: 398 },
    { name: 'Audio Comprehension', accuracy: 85, attempts: 176 },
  ];

  return (
    <section id="analytics" className="py-20 bg-gradient-to-br from-[#FDF6E3] to-orange-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-16">
          <h2 className="text-4xl lg:text-5xl font-bold mb-4">
            Track Your Progress with 
            <span className="bg-gradient-to-r from-[#FF7B54] to-[#FFD700] bg-clip-text text-transparent"> Advanced Analytics</span>
          </h2>
          <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
            Comprehensive learning insights powered by sophisticated analytics engine. Monitor every aspect of your vocabulary journey with precision.
          </p>
        </div>

        {/* Dashboard Preview */}
        <div className="bg-white rounded-3xl shadow-2xl p-8 mb-16 transform hover:scale-105 transition-transform duration-300">
          <div className="text-center mb-8">
            <h3 className="text-2xl font-bold text-[#2D3436] mb-2">Your Learning Dashboard</h3>
            <p className="text-[#636E72]">Real-time statistics updated after every quiz</p>
          </div>
          
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-6">
            {dashboardStats.map((stat) => (
              <div key={stat.label} className="text-center group">
                <div className="bg-gradient-to-br from-gray-50 to-gray-100 rounded-2xl p-6 group-hover:shadow-lg transition-shadow duration-300">
                  <div className="text-4xl mb-3">{stat.icon}</div>
                  <div className="text-3xl font-bold text-[#2D3436] mb-1">
                    {stat.value.toLocaleString()}{stat.suffix}
                  </div>
                  <div className="text-sm font-semibold text-[#636E72] mb-1">{stat.label}</div>
                  <div className="text-xs text-[#636E72]">{stat.description}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Analytics Features Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-16">
          {metrics.map((metric) => (
            <div key={metric.title} className="bg-white/80 backdrop-blur rounded-3xl p-8 shadow-xl hover:shadow-2xl transform hover:-translate-y-2 transition-all duration-500">
              <div className="bg-gradient-to-br from-[#FF7B54] to-orange-600 text-white p-4 rounded-2xl mb-6 w-16 h-16 flex items-center justify-center">
                {metric.icon}
              </div>
              <h3 className="text-xl font-bold text-[#2D3436] mb-4">{metric.title}</h3>
              <p className="text-[#636E72] mb-6 leading-relaxed">{metric.description}</p>
              <ul className="space-y-2">
                {metric.details.map((detail, idx) => (
                  <li key={idx} className="flex items-center text-sm text-[#636E72]">
                    <i className="fas fa-check-circle text-[#4CAF50] mr-3"></i>
                    {detail}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Quiz Performance Chart Preview */}
        <div className="bg-white rounded-3xl shadow-2xl p-8 mb-16">
          <h3 className="text-2xl font-bold text-[#2D3436] text-center mb-8">Quiz Type Performance Analysis</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {quizTypes.map((quiz) => (
              <div key={quiz.name} className="bg-gradient-to-br from-gray-50 to-gray-100 rounded-xl p-4 hover:shadow-md transition-shadow duration-300">
                <div className="text-sm font-semibold text-[#2D3436] mb-2">{quiz.name}</div>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-[#636E72]">Accuracy</span>
                  <span className="text-lg font-bold text-[#FF7B54]">{quiz.accuracy}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
                  <div 
                    className="bg-gradient-to-r from-[#FF7B54] to-orange-600 h-2 rounded-full transition-all duration-1000"
                    style={{ width: `${quiz.accuracy}%` }}
                  ></div>
                </div>
                <div className="text-xs text-[#636E72]">{quiz.attempts} attempts</div>
              </div>
            ))}
          </div>
        </div>

        {/* Advanced Features Highlight */}
        <div className="bg-gradient-to-r from-[#FF7B54] to-orange-600 rounded-3xl p-8 text-white text-center">
          <h3 className="text-3xl font-bold mb-4">Advanced Learning Intelligence</h3>
          <p className="text-xl mb-8 opacity-90">
            Powered by materialized views and real-time analytics for instant insights
          </p>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="bg-white/10 backdrop-blur rounded-2xl p-6">
              <i className="fas fa-database text-4xl mb-4"></i>
              <h4 className="text-lg font-semibold mb-2">Optimized Performance</h4>
              <p className="text-sm opacity-90">Materialized views ensure lightning-fast analytics queries</p>
            </div>
            <div className="bg-white/10 backdrop-blur rounded-2xl p-6">
              <i className="fas fa-sync-alt text-4xl mb-4"></i>
              <h4 className="text-lg font-semibold mb-2">Real-time Updates</h4>
              <p className="text-sm opacity-90">Analytics refresh instantly after every quiz completion</p>
            </div>
            <div className="bg-white/10 backdrop-blur rounded-2xl p-6">
              <i className="fas fa-shield-alt text-4xl mb-4"></i>
              <h4 className="text-lg font-semibold mb-2">Data Integrity</h4>
              <p className="text-sm opacity-90">Transactional updates ensure consistent analytics data</p>
            </div>
          </div>

          <div className="mt-8">
            <a href="/signup?plan=free" className="bg-white text-[#FF7B54] px-8 py-4 rounded-full font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 inline-block">
              Start Tracking Your Progress
            </a>
          </div>
        </div>

        {/* Data Points Showcase */}
        <div className="mt-16 text-center">
          <h3 className="text-2xl font-bold text-[#2D3436] mb-8">Comprehensive Data Tracking</h3>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-6">
            {[
              { label: 'Mastery Levels', value: '0-100%' },
              { label: 'Quiz Types', value: '7 Types' },
              { label: 'Leitner Boxes', value: '7 Levels' },
              { label: 'Response Time', value: 'Milliseconds' },
              { label: 'Accuracy Rate', value: 'Real-time' },
              { label: 'Streak Tracking', value: 'Daily' },
            ].map((item) => (
              <div key={item.label} className="bg-white rounded-2xl p-4 shadow-lg hover:shadow-xl transition-shadow duration-300">
                <div className="text-2xl font-bold text-[#FF7B54] mb-1">{item.value}</div>
                <div className="text-xs text-[#636E72] font-medium">{item.label}</div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default AnalyticsSection;
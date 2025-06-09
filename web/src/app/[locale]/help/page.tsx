import React from 'react';
import Link from 'next/link';
import PageLayout from '../../../components/layout/PageLayout';

const HelpCenterPage: React.FC = () => {
  return (
    <PageLayout>
      <div className="pt-24 pb-20 min-h-screen">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="text-center mb-16">
            <h1 className="text-4xl lg:text-5xl font-bold text-[#2D3436] mb-4">
              Help Center
            </h1>
            <p className="text-xl text-[#636E72] max-w-3xl mx-auto">
              Everything you need to know about Decorebator - AI-powered vocabulary learning
            </p>
          </div>

          {/* Quick Navigation */}
          <div className="bg-gradient-to-br from-orange-50 to-amber-50 rounded-3xl p-8 mb-16">
            <h2 className="text-2xl font-bold text-[#2D3436] mb-6 text-center">Quick Navigation</h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <a href="#getting-started" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-rocket text-2xl text-[#FF7B54] mb-2"></i>
                <span className="text-sm font-medium">Getting Started</span>
              </a>
              <a href="#features" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-star text-2xl text-[#4CAF50] mb-2"></i>
                <span className="text-sm font-medium">Features</span>
              </a>
              <a href="#leitner-system" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-brain text-2xl text-[#9C27B0] mb-2"></i>
                <span className="text-sm font-medium">Leitner System</span>
              </a>
              <a href="#troubleshooting" className="flex flex-col items-center p-4 bg-white rounded-xl hover:shadow-lg transition-shadow duration-300">
                <i className="fas fa-tools text-2xl text-[#14B8A6] mb-2"></i>
                <span className="text-sm font-medium">Troubleshooting</span>
              </a>
            </div>
          </div>

          {/* Getting Started Section */}
          <section id="getting-started" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-rocket text-[#FF7B54] mr-3"></i>
                Getting Started
              </h2>
              
              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">1. Create Your Account</h3>
                  <ul className="space-y-2 text-[#636E72]">
                    <li>• Sign up with email and password</li>
                    <li>• Choose your learning plan (Free or Premium)</li>
                    <li>• Download the mobile app from your app store</li>
                    <li>• Sign in with the same credentials</li>
                  </ul>
                </div>
                
                <div>
                  <h3 className="text-xl font-bold mb-4">2. Start Learning</h3>
                  <ul className="space-y-2 text-[#636E72]">
                    <li>• Create your first wordlist</li>
                    <li>• Add words you want to learn</li>
                    <li>• AI automatically generates definitions and content</li>
                    <li>• Begin practicing with interactive quizzes</li>
                  </ul>
                </div>
              </div>

              <div className="mt-8 p-6 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl border border-blue-200">
                <h4 className="font-bold text-[#2D3436] mb-2 flex items-center">
                  <i className="fas fa-lightbulb text-blue-600 mr-2"></i>
                  Pro Tip
                </h4>
                <p className="text-[#636E72]">
                  Start with 5-10 words in your first wordlist to get familiar with the system before adding more!
                </p>
              </div>
            </div>
          </section>

          {/* Features Section */}
          <section id="features" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-star text-[#4CAF50] mr-3"></i>
                Core Features
              </h2>

              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#FF7B54] mb-3 flex items-center">
                    <i className="fas fa-brain mr-2"></i>
                    AI-Powered Content
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    Automatically generates definitions, images, audio pronunciations, and example sentences for every word you add.
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#4CAF50] mb-3 flex items-center">
                    <i className="fas fa-gamepad mr-2"></i>
                    8 Quiz Modes
                  </h4>
                  <ul className="text-sm text-[#636E72] space-y-1">
                    <li>• Guess Meaning</li>
                    <li>• Word from Meaning</li>
                    <li>• Image Association</li>
                    <li>• Audio Comprehension</li>
                    <li>• Sentence Completion</li>
                    <li>• Write from Definition</li>
                    <li>• Audio Recognition</li>
                    <li>• Example Audio Recognition</li>
                  </ul>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#9C27B0] mb-3 flex items-center">
                    <i className="fas fa-clock mr-2"></i>
                    Spaced Repetition
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    Uses the scientifically-proven Leitner system with 7 boxes to optimize when you review words for maximum retention.
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#14B8A6] mb-3 flex items-center">
                    <i className="fas fa-chart-line mr-2"></i>
                    Advanced Analytics
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    Track your progress with detailed statistics, word mastery levels, learning velocity, and retention rates.
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#6366F1] mb-3 flex items-center">
                    <i className="fas fa-globe mr-2"></i>
                    Multi-Language Support
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    Native AI support for 7 languages: English, Spanish, French, German, Italian, Portuguese, and Japanese.
                  </p>
                </div>

                <div className="border border-gray-200 rounded-xl p-6">
                  <h4 className="font-bold text-[#FF6B6B] mb-3 flex items-center">
                    <i className="fas fa-wifi-slash mr-2"></i>
                    Offline Support
                  </h4>
                  <p className="text-sm text-[#636E72]">
                    Premium users can download wordlists for offline access and continue learning without internet connection.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Leitner System Section */}
          <section id="leitner-system" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-brain text-[#9C27B0] mr-3"></i>
                How the Leitner System Works
              </h2>
              
              <p className="text-[#636E72] mb-8">
                Decorebator uses a sophisticated 7-box Leitner system for optimal vocabulary retention through spaced repetition.
              </p>

              <div className="overflow-x-auto mb-8">
                <table className="w-full border-collapse border border-gray-200 rounded-lg">
                  <thead className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white">
                    <tr>
                      <th className="border border-gray-200 p-4 text-left">Box</th>
                      <th className="border border-gray-200 p-4 text-left">Review Interval</th>
                      <th className="border border-gray-200 p-4 text-left">Purpose</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="bg-red-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 1</td>
                      <td className="border border-gray-200 p-4">Immediate</td>
                      <td className="border border-gray-200 p-4">New words, failed reviews</td>
                    </tr>
                    <tr className="bg-orange-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 2</td>
                      <td className="border border-gray-200 p-4">1 hour</td>
                      <td className="border border-gray-200 p-4">Recent learning</td>
                    </tr>
                    <tr className="bg-yellow-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 3</td>
                      <td className="border border-gray-200 p-4">1 day</td>
                      <td className="border border-gray-200 p-4">Short-term retention</td>
                    </tr>
                    <tr className="bg-blue-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 4</td>
                      <td className="border border-gray-200 p-4">3 days</td>
                      <td className="border border-gray-200 p-4">Medium-term retention</td>
                    </tr>
                    <tr className="bg-purple-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 5</td>
                      <td className="border border-gray-200 p-4">1 week</td>
                      <td className="border border-gray-200 p-4">Long-term retention</td>
                    </tr>
                    <tr className="bg-indigo-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 6</td>
                      <td className="border border-gray-200 p-4">2 weeks</td>
                      <td className="border border-gray-200 p-4">Extended retention</td>
                    </tr>
                    <tr className="bg-green-50">
                      <td className="border border-gray-200 p-4 font-bold">Box 7</td>
                      <td className="border border-gray-200 p-4">1 month</td>
                      <td className="border border-gray-200 p-4">Mastered words</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div className="grid md:grid-cols-2 gap-8">
                <div className="bg-green-50 border border-green-200 rounded-xl p-6">
                  <h4 className="font-bold text-green-800 mb-3 flex items-center">
                    <i className="fas fa-check-circle mr-2"></i>
                    Progression Rules
                  </h4>
                  <ul className="text-green-700 space-y-2">
                    <li>• <strong>Correct Answer:</strong> Word moves to the next box</li>
                    <li>• <strong>Incorrect Answer:</strong> Word resets to Box 1</li>
                    <li>• <strong>Box 7:</strong> Words stay in Box 7 when correct (mastered)</li>
                  </ul>
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-xl p-6">
                  <h4 className="font-bold text-blue-800 mb-3 flex items-center">
                    <i className="fas fa-graduation-cap mr-2"></i>
                    Quiz Type Progression
                  </h4>
                  <p className="text-blue-700 text-sm">
                    Different quiz types are introduced based on the word&apos;s box level to increase difficulty progressively as you master each word.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* Subscription Plans */}
          <section id="plans" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-crown text-[#FFD700] mr-3"></i>
                Subscription Plans
              </h2>

              <div className="grid md:grid-cols-3 gap-6">
                <div className="border border-gray-200 rounded-xl p-6 text-center">
                  <h3 className="text-xl font-bold text-[#4CAF50] mb-4">Free Plan</h3>
                  <div className="text-3xl font-bold mb-4">$0</div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    <li>• 1 wordlist maximum</li>
                    <li>• Up to 10 words per wordlist</li>
                    <li>• Basic quiz modes</li>
                    <li>• Online-only access</li>
                  </ul>
                  <Link href="/signup?plan=free" className="bg-[#4CAF50] text-white px-6 py-2 rounded-full font-semibold hover:bg-green-600 transition-colors duration-300 inline-block">
                    Get Started Free
                  </Link>
                </div>

                <div className="border-2 border-[#FF7B54] rounded-xl p-6 text-center relative">
                  <div className="absolute -top-3 left-1/2 transform -translate-x-1/2 bg-[#FF7B54] text-white px-4 py-1 rounded-full text-sm font-semibold">
                    Most Popular
                  </div>
                  <h3 className="text-xl font-bold text-[#FF7B54] mb-4">Monthly Premium</h3>
                  <div className="text-3xl font-bold mb-4">$6.99<span className="text-sm text-[#636E72]">/month</span></div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    <li>• Unlimited wordlists</li>
                    <li>• Unlimited words</li>
                    <li>• All 8 quiz modes</li>
                    <li>• Advanced analytics</li>
                    <li>• Offline support</li>
                    <li>• Priority support</li>
                  </ul>
                  <Link href="/signup?plan=monthly" className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white px-6 py-2 rounded-full font-semibold hover:shadow-lg transition-all duration-300 inline-block">
                    Start Monthly
                  </Link>
                </div>

                <div className="border border-gray-200 rounded-xl p-6 text-center">
                  <h3 className="text-xl font-bold text-[#9C27B0] mb-4">Annual Premium</h3>
                  <div className="text-3xl font-bold mb-2">$69.90<span className="text-sm text-[#636E72]">/year</span></div>
                  <div className="text-sm text-[#4CAF50] font-semibold mb-4">Save $13.98!</div>
                  <ul className="text-sm text-[#636E72] space-y-2 mb-6">
                    <li>• Everything in Monthly</li>
                    <li>• Best value for money</li>
                    <li>• Priority customer support</li>
                    <li>• Early access to features</li>
                  </ul>
                  <Link href="/signup?plan=annual" className="bg-gradient-to-r from-[#9C27B0] to-purple-600 text-white px-6 py-2 rounded-full font-semibold hover:shadow-lg transition-all duration-300 inline-block">
                    Start Annual
                  </Link>
                </div>
              </div>
            </div>
          </section>

          {/* Language Support */}
          <section id="languages" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-globe text-[#6366F1] mr-3"></i>
                Multi-Language Support
              </h2>

              <p className="text-[#636E72] mb-8">
                Decorebator provides comprehensive multi-language support with native AI processing for authentic learning experiences.
              </p>

              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">Supported Languages (7)</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-red-50 to-orange-50 rounded-lg">
                      <span className="text-2xl">🇺🇸</span>
                      <span className="font-semibold">English</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-yellow-50 to-red-50 rounded-lg">
                      <span className="text-2xl">🇪🇸</span>
                      <span className="font-semibold">Spanish</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-blue-50 to-white rounded-lg">
                      <span className="text-2xl">🇫🇷</span>
                      <span className="font-semibold">French</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-black to-red-50 rounded-lg">
                      <span className="text-2xl">🇩🇪</span>
                      <span className="font-semibold">German</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-green-50 to-red-50 rounded-lg">
                      <span className="text-2xl">🇮🇹</span>
                      <span className="font-semibold">Italian</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-green-50 to-yellow-50 rounded-lg">
                      <span className="text-2xl">🇵🇹</span>
                      <span className="font-semibold">Portuguese</span>
                    </div>
                    <div className="flex items-center space-x-3 p-3 bg-gradient-to-r from-red-50 to-white rounded-lg col-span-2">
                      <span className="text-2xl">🇯🇵</span>
                      <span className="font-semibold">Japanese</span>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-bold mb-4">AI Features per Language</h3>
                  <ul className="space-y-3 text-[#636E72]">
                    <li className="flex items-start">
                      <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                      <span><strong>Native Grammar:</strong> Language-specific rules and structures</span>
                    </li>
                    <li className="flex items-start">
                      <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                      <span><strong>Cultural Context:</strong> Culturally relevant images and examples</span>
                    </li>
                    <li className="flex items-start">
                      <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                      <span><strong>Optimized Audio:</strong> Native-sounding voice selection per language</span>
                    </li>
                    <li className="flex items-start">
                      <i className="fas fa-check-circle text-[#4CAF50] mt-1 mr-3"></i>
                      <span><strong>Verb Systems:</strong> Language-appropriate tenses and inflections</span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </section>

          {/* Troubleshooting Section */}
          <section id="troubleshooting" className="mb-16">
            <div className="bg-white rounded-3xl shadow-xl p-8">
              <h2 className="text-3xl font-bold text-[#2D3436] mb-6 flex items-center">
                <i className="fas fa-tools text-[#14B8A6] mr-3"></i>
                Troubleshooting & Support
              </h2>

              <div className="grid md:grid-cols-2 gap-8">
                <div>
                  <h3 className="text-xl font-bold mb-4">Common Issues</h3>
                  <div className="space-y-4">
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">Audio not playing</h4>
                      <p className="text-sm text-[#636E72]">Check your device volume and ensure you have a stable internet connection. Audio is generated by AI and requires online access.</p>
                    </div>
                    
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">Images not loading</h4>
                      <p className="text-sm text-[#636E72]">Images are generated by AI and may take a few moments to appear. Refresh the quiz or use the error reporting feature if issues persist.</p>
                    </div>
                    
                    <div className="border border-gray-200 rounded-lg p-4">
                      <h4 className="font-semibold text-[#2D3436] mb-2">Subscription not updating</h4>
                      <p className="text-sm text-[#636E72]">After completing payment, close and reopen the app to refresh your subscription status. Changes should appear within a few minutes.</p>
                    </div>
                  </div>
                </div>

                <div>
                  <h3 className="text-xl font-bold mb-4">Error Reporting</h3>
                  <p className="text-[#636E72] mb-4">
                    If you encounter issues with AI-generated content, use our built-in error reporting system:
                  </p>
                  <ul className="space-y-2 text-[#636E72] mb-6">
                    <li>• Report incorrect definitions</li>
                    <li>• Flag inappropriate images</li>
                    <li>• Report audio playback issues</li>
                    <li>• Submit irrelevant examples</li>
                    <li>• Flag loading problems</li>
                  </ul>
                  
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <h4 className="font-semibold text-blue-800 mb-2">Need More Help?</h4>
                    <p className="text-blue-700 text-sm mb-3">
                      Our support team is here to help you get the most out of Decorebator.
                    </p>
                    <a href="mailto:support@decorebator.com" className="text-blue-600 hover:text-blue-800 font-semibold text-sm">
                      Contact Support: support@decorebator.com
                    </a>
                  </div>
                </div>
              </div>
            </div>
          </section>

          {/* Contact Support */}
          <section className="text-center">
            <div className="bg-gradient-to-r from-[#FF7B54] to-orange-600 rounded-3xl p-8 text-white">
              <h2 className="text-3xl font-bold mb-4">Still Need Help?</h2>
              <p className="text-xl mb-6 opacity-90">
                Our support team is ready to assist you with any questions or issues.
              </p>
              <div className="flex flex-col sm:flex-row gap-4 justify-center">
                <a
                  href="mailto:support@decorebator.com"
                  className="bg-white text-[#FF7B54] px-8 py-3 rounded-full font-semibold hover:shadow-lg transition-all duration-300 inline-block"
                >
                  <i className="fas fa-envelope mr-2"></i>
                  Email Support
                </a>
                <Link
                  href="/#contact"
                  className="bg-white/20 backdrop-blur text-white px-8 py-3 rounded-full font-semibold border-2 border-white/50 hover:bg-white/30 transition-all duration-300 inline-block"
                >
                  <i className="fas fa-comments mr-2"></i>
                  Contact Form
                </Link>
              </div>
            </div>
          </section>
        </div>
      </div>
    </PageLayout>
  );
};

export default HelpCenterPage;
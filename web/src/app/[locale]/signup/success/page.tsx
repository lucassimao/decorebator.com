'use client';

import React, { useEffect, useState, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import PageLayout from '../../../../components/layout/PageLayout';

const SignUpSuccessContent: React.FC = () => {
  const searchParams = useSearchParams();
  const plan = searchParams.get('plan');
  const isPremium = searchParams.get('premium') === 'true';
  const fromStripe = searchParams.get('stripe') === 'true';
  
  const [showAnimation, setShowAnimation] = useState(false);

  useEffect(() => {
    // Trigger success animation
    const timer = setTimeout(() => {
      setShowAnimation(true);
    }, 300);

    return () => clearTimeout(timer);
  }, []);

  const handleStripeRedirect = () => {
    // In a real implementation, this would redirect to actual Stripe
    // For now, simulate the process
    
    // Simulate Stripe redirect and return
    setTimeout(() => {
      window.location.href = `/signup/success?plan=${plan}&premium=true&stripe=true`;
    }, 2000);
    
    // In reality, you would redirect to Stripe:
    // window.location.href = stripeUrl;
  };

  if (isPremium && !fromStripe) {
    // Premium user needs to go through Stripe first
    return (
      <PageLayout>
        <div className="pt-24 pb-20 min-h-screen">
          <div className="max-w-lg mx-auto px-4 sm:px-6 lg:px-8">
            <div className="bg-white rounded-3xl shadow-2xl p-8 text-center">
              <div className="w-20 h-20 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6">
                <i className="fas fa-credit-card text-white text-2xl"></i>
              </div>
              
              <h1 className="text-3xl font-bold text-[#2D3436] mb-4">
                Complete Your Payment
              </h1>
              
              <p className="text-[#636E72] mb-8">
                Your account has been created! Now let&apos;s set up your {plan === 'monthly' ? 'monthly' : 'annual'} premium subscription.
              </p>

              <div className="bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white rounded-2xl p-6 mb-8">
                <div className="text-lg font-bold mb-2">
                  {plan === 'monthly' ? 'Monthly Premium' : 'Annual Premium'}
                </div>
                <div className="text-2xl font-bold">
                  {plan === 'monthly' ? '$6.99/month' : '$69.90/year'}
                </div>
                {plan === 'annual' && (
                  <div className="text-sm bg-white/20 rounded-full px-3 py-1 mt-2 inline-block">
                    Save $13.98!
                  </div>
                )}
              </div>

              <button
                onClick={handleStripeRedirect}
                className="w-full bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white py-4 rounded-xl font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 mb-6"
              >
                <div className="flex items-center justify-center space-x-2">
                  <i className="fab fa-stripe text-xl"></i>
                  <span>Continue to Stripe Payment</span>
                </div>
              </button>

              <div className="text-xs text-[#636E72] flex items-center justify-center space-x-2">
                <i className="fas fa-shield-alt text-green-500"></i>
                <span>Secured by Stripe • SSL Encrypted</span>
              </div>
            </div>
          </div>
        </div>
      </PageLayout>
    );
  }

  // Success page for completed sign-ups
  return (
    <PageLayout>
      <div className="pt-24 pb-20 min-h-screen">
        <div className="max-w-lg mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            {/* Success Animation */}
            <div className={`transform transition-all duration-1000 ${showAnimation ? 'scale-100 opacity-100' : 'scale-50 opacity-0'}`}>
              <div className="w-24 h-24 bg-gradient-to-br from-[#4CAF50] to-green-600 rounded-full flex items-center justify-center mx-auto mb-8">
                <i className="fas fa-check text-white text-3xl"></i>
              </div>
            </div>

            <h1 className="text-4xl font-bold text-[#2D3436] mb-4">
              {isPremium ? 'Welcome to Premium!' : 'Welcome to Decorebator!'}
            </h1>
            
            <p className="text-xl text-[#636E72] mb-8">
              {isPremium 
                ? 'Your premium account is ready. Time to unlock unlimited vocabulary learning!'
                : 'Your free account is ready. Start your vocabulary learning journey today!'
              }
            </p>

            {/* Account Details */}
            <div className="bg-white rounded-3xl shadow-2xl p-8 mb-8">
              <div className="flex items-center justify-center space-x-3 mb-6">
                <div className="w-12 h-12 bg-gradient-to-br from-[#FF7B54] to-orange-600 rounded-full flex items-center justify-center">
                  <i className="fas fa-user text-white"></i>
                </div>
                <div className="text-left">
                  <div className="font-semibold text-[#2D3436]">Account Created</div>
                  <div className="text-sm text-[#636E72]">
                    {isPremium ? `${plan === 'monthly' ? 'Monthly' : 'Annual'} Premium Plan` : 'Free Plan'}
                  </div>
                </div>
              </div>

              {isPremium && (
                <div className="bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-200 rounded-xl p-4 mb-6">
                  <div className="flex items-center space-x-2 text-amber-800">
                    <i className="fas fa-crown"></i>
                    <span className="font-semibold">Premium Features Unlocked:</span>
                  </div>
                  <ul className="text-sm text-amber-700 mt-2 space-y-1">
                    <li>• Unlimited wordlists and words</li>
                    <li>• All 8 quiz modes and flashcards</li>
                    <li>• Advanced analytics and progress tracking</li>
                    <li>• Offline support</li>
                    <li>• Priority support</li>
                  </ul>
                </div>
              )}

              <div className="text-center">
                <h3 className="text-xl font-bold text-[#2D3436] mb-4">
                  Download the Decorebator App
                </h3>
                <p className="text-[#636E72] mb-6">
                  Get the mobile app to start learning on the go. Sign in with the same email you just used.
                </p>

                {/* App Store Buttons */}
                <div className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-6">
                  <a 
                    href="https://apps.apple.com/app/decorebator" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="group transform hover:scale-105 transition-transform duration-300"
                  >
                    <img 
                      src="https://upload.wikimedia.org/wikipedia/commons/3/3c/Download_on_the_App_Store_Badge.svg" 
                      alt="Download on App Store" 
                      className="h-14"
                    />
                  </a>
                  <a 
                    href="https://play.google.com/store/apps/details?id=com.decorebator" 
                    target="_blank" 
                    rel="noopener noreferrer"
                    className="group transform hover:scale-105 transition-transform duration-300"
                  >
                    <img 
                      src="https://upload.wikimedia.org/wikipedia/commons/7/78/Google_Play_Store_badge_EN.svg" 
                      alt="Get it on Google Play" 
                      className="h-14"
                    />
                  </a>
                </div>

                {/* Instructions */}
                <div className="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-xl p-6">
                  <h4 className="font-semibold text-[#2D3436] mb-3 flex items-center justify-center">
                    <i className="fas fa-mobile-alt text-blue-600 mr-2"></i>
                    Quick Setup Guide
                  </h4>
                  <ol className="text-sm text-[#636E72] space-y-2 text-left">
                    <li className="flex items-start">
                      <span className="bg-blue-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs mr-3 mt-0.5 flex-shrink-0">1</span>
                      Download the app from your preferred app store
                    </li>
                    <li className="flex items-start">
                      <span className="bg-blue-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs mr-3 mt-0.5 flex-shrink-0">2</span>
                      Open the app and tap &quot;Sign In&quot;
                    </li>
                    <li className="flex items-start">
                      <span className="bg-blue-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs mr-3 mt-0.5 flex-shrink-0">3</span>
                      Use the same email and password you just created
                    </li>
                    <li className="flex items-start">
                      <span className="bg-blue-600 text-white rounded-full w-5 h-5 flex items-center justify-center text-xs mr-3 mt-0.5 flex-shrink-0">4</span>
                      Start creating your first wordlist and begin learning!
                    </li>
                  </ol>
                </div>
              </div>
            </div>

            {/* Additional Actions */}
            <div className="space-y-4">
              <button
                onClick={() => window.location.href = '/'}
                className="w-full bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white py-4 rounded-xl font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300"
              >
                <div className="flex items-center justify-center space-x-2">
                  <i className="fas fa-home"></i>
                  <span>Return to Homepage</span>
                </div>
              </button>

              {!isPremium && (
                <div className="text-center">
                  <p className="text-[#636E72] mb-4">Want to upgrade to premium later?</p>
                  <p className="text-sm text-[#636E72]">
                    You can upgrade anytime from within the mobile app to unlock unlimited features.
                  </p>
                </div>
              )}
            </div>

            {/* Support Info */}
            <div className="mt-8 p-4 bg-gray-50 rounded-xl">
              <div className="flex items-center justify-center space-x-2 text-sm text-[#636E72]">
                <i className="fas fa-question-circle"></i>
                <span>Need help? Contact</span>
                <a href="mailto:support@decorebator.com" className="text-[#FF7B54] hover:text-orange-600 font-semibold">
                  support@decorebator.com
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  );
};

const SignUpSuccessPage: React.FC = () => {
  return (
    <Suspense fallback={
      <PageLayout>
        <div className="pt-24 pb-20 min-h-screen flex items-center justify-center">
          <div className="text-center">
            <i className="fas fa-spinner fa-spin text-4xl text-[#FF7B54] mb-4"></i>
            <p className="text-[#636E72]">Loading...</p>
          </div>
        </div>
      </PageLayout>
    }>
      <SignUpSuccessContent />
    </Suspense>
  );
};

export default SignUpSuccessPage;
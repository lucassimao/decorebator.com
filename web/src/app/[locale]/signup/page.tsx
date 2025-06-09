'use client';

import React, { useState, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import PageLayout from '../../../components/layout/PageLayout';

const SignUpContent: React.FC = () => {
  const searchParams = useSearchParams();
  const plan = searchParams.get('plan'); // 'free', 'monthly', 'annual'
  
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const isPremiumPlan = plan === 'monthly' || plan === 'annual';

  const validateForm = () => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    
    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }
    
    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }
    
    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) return;
    
    setIsLoading(true);
    
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Redirect based on plan type
      if (isPremiumPlan) {
        window.location.href = `/signup/success?plan=${plan}&premium=true`;
      } else {
        window.location.href = '/signup/success?plan=free';
      }
    } catch (error) {
      console.error('Sign up error:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const getPlanDetails = () => {
    switch (plan) {
      case 'monthly':
        return { name: 'Monthly Premium', price: '$6.99/month', savings: null };
      case 'annual':
        return { name: 'Annual Premium', price: '$69.90/year', savings: 'Save $13.98!' };
      default:
        return { name: 'Free Plan', price: 'Free forever', savings: null };
    }
  };

  const planDetails = getPlanDetails();

  return (
    <PageLayout>
      <div className="pt-24 pb-20 min-h-screen">
        <div className="max-w-md mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold text-[#2D3436] mb-2">
              {isPremiumPlan ? 'Start Your Premium Journey' : 'Start Learning Free'}
            </h1>
            <p className="text-[#636E72]">
              {isPremiumPlan 
                ? 'Create your account to unlock premium features' 
                : 'Create your free account and begin mastering vocabulary'
              }
            </p>
          </div>

          {/* Plan Badge */}
          <div className="mb-8">
            <div className={`bg-gradient-to-r ${isPremiumPlan ? 'from-[#FF7B54] to-orange-600' : 'from-[#4CAF50] to-green-600'} text-white rounded-2xl p-4 text-center`}>
              <div className="text-lg font-bold">{planDetails.name}</div>
              <div className="text-sm opacity-90">{planDetails.price}</div>
              {planDetails.savings && (
                <div className="text-xs bg-white/20 rounded-full px-3 py-1 mt-2 inline-block">
                  {planDetails.savings}
                </div>
              )}
            </div>
            
            {isPremiumPlan && (
              <div className="mt-4 p-4 bg-amber-50 border border-amber-200 rounded-xl">
                <div className="flex items-start space-x-3">
                  <i className="fas fa-info-circle text-amber-600 mt-0.5"></i>
                  <div className="text-sm text-amber-800">
                    <p className="font-semibold mb-1">Premium Sign-up Process:</p>
                    <p>After creating your account, you&apos;ll be redirected to Stripe for secure payment processing, then guided to download the app.</p>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Sign-up Form */}
          <div className="bg-white rounded-3xl shadow-2xl p-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Name Field */}
              <div>
                <label htmlFor="name" className="block text-sm font-semibold text-[#2D3436] mb-2">
                  Full Name
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  className={`w-full px-4 py-3 border-2 rounded-xl focus:outline-none transition-colors duration-300 ${
                    errors.name 
                      ? 'border-red-300 focus:border-red-500' 
                      : 'border-gray-200 focus:border-[#FF7B54]'
                  }`}
                  placeholder="Enter your full name"
                />
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name}</p>
                )}
              </div>

              {/* Email Field */}
              <div>
                <label htmlFor="email" className="block text-sm font-semibold text-[#2D3436] mb-2">
                  Email Address
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  className={`w-full px-4 py-3 border-2 rounded-xl focus:outline-none transition-colors duration-300 ${
                    errors.email 
                      ? 'border-red-300 focus:border-red-500' 
                      : 'border-gray-200 focus:border-[#FF7B54]'
                  }`}
                  placeholder="Enter your email address"
                />
                {errors.email && (
                  <p className="text-red-500 text-sm mt-1">{errors.email}</p>
                )}
              </div>

              {/* Password Field */}
              <div>
                <label htmlFor="password" className="block text-sm font-semibold text-[#2D3436] mb-2">
                  Password
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    id="password"
                    name="password"
                    value={formData.password}
                    onChange={handleInputChange}
                    className={`w-full px-4 py-3 pr-12 border-2 rounded-xl focus:outline-none transition-colors duration-300 ${
                      errors.password 
                        ? 'border-red-300 focus:border-red-500' 
                        : 'border-gray-200 focus:border-[#FF7B54]'
                    }`}
                    placeholder="Create a strong password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-4 top-1/2 transform -translate-y-1/2 text-[#636E72] hover:text-[#FF7B54] transition-colors duration-300"
                  >
                    <i className={`fas ${showPassword ? 'fa-eye-slash' : 'fa-eye'}`}></i>
                  </button>
                </div>
                {errors.password && (
                  <p className="text-red-500 text-sm mt-1">{errors.password}</p>
                )}
              </div>

              {/* Confirm Password Field */}
              <div>
                <label htmlFor="confirmPassword" className="block text-sm font-semibold text-[#2D3436] mb-2">
                  Confirm Password
                </label>
                <input
                  type={showPassword ? 'text' : 'password'}
                  id="confirmPassword"
                  name="confirmPassword"
                  value={formData.confirmPassword}
                  onChange={handleInputChange}
                  className={`w-full px-4 py-3 border-2 rounded-xl focus:outline-none transition-colors duration-300 ${
                    errors.confirmPassword 
                      ? 'border-red-300 focus:border-red-500' 
                      : 'border-gray-200 focus:border-[#FF7B54]'
                  }`}
                  placeholder="Confirm your password"
                />
                {errors.confirmPassword && (
                  <p className="text-red-500 text-sm mt-1">{errors.confirmPassword}</p>
                )}
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isLoading}
                className={`w-full bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white py-4 rounded-xl font-semibold text-lg hover:shadow-2xl transform hover:scale-105 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none ${
                  isLoading ? 'cursor-wait' : ''
                }`}
              >
                {isLoading ? (
                  <div className="flex items-center justify-center space-x-2">
                    <i className="fas fa-spinner fa-spin"></i>
                    <span>Creating Account...</span>
                  </div>
                ) : (
                  <div className="flex items-center justify-center space-x-2">
                    <span>
                      {isPremiumPlan ? 'Continue to Payment' : 'Create Free Account'}
                    </span>
                    <i className="fas fa-arrow-right"></i>
                  </div>
                )}
              </button>
            </form>

            {/* Footer Links */}
            <div className="mt-6 text-center">
              <p className="text-sm text-[#636E72]">
                Already have an account?{' '}
                <Link href="/signin" className="text-[#FF7B54] hover:text-orange-600 font-semibold transition-colors duration-300">
                  Sign In
                </Link>
              </p>
              
              <div className="mt-4 text-xs text-[#636E72]">
                By creating an account, you agree to our{' '}
                <Link href="/terms" className="text-[#FF7B54] hover:text-orange-600 transition-colors duration-300">
                  Terms of Service
                </Link>{' '}
                and{' '}
                <Link href="/privacy" className="text-[#FF7B54] hover:text-orange-600 transition-colors duration-300">
                  Privacy Policy
                </Link>
              </div>
            </div>
          </div>

          {/* Alternative Plans */}
          {!isPremiumPlan && (
            <div className="mt-8 text-center">
              <p className="text-[#636E72] mb-4">Want premium features?</p>
              <div className="flex flex-col sm:flex-row gap-3">
                <Link
                  href="/signup?plan=monthly"
                  className="flex-1 bg-white border-2 border-[#FF7B54] text-[#FF7B54] py-3 px-6 rounded-xl font-semibold hover:bg-[#FF7B54] hover:text-white transition-all duration-300"
                >
                  Monthly - $6.99
                </Link>
                <Link
                  href="/signup?plan=annual"
                  className="flex-1 bg-gradient-to-r from-[#FF7B54] to-orange-600 text-white py-3 px-6 rounded-xl font-semibold hover:shadow-lg transform hover:scale-105 transition-all duration-300"
                >
                  Annual - $69.90 (Save!)
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </PageLayout>
  );
};

const SignUpPage: React.FC = () => {
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
      <SignUpContent />
    </Suspense>
  );
};

export default SignUpPage;
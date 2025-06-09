
import React from 'react';
import PageLayout from '../../../components/layout/PageLayout';
import { PolicySection } from '../../../components/policy/PolicySection';

export default function TermsOfService() {
  return (
    <PageLayout>
      <div className="pt-24 pb-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h1 className="text-4xl md:text-5xl font-bold text-[#2D3436] mb-4">
              Terms of Service
            </h1>
            <p className="text-xl text-[#636E72]">
              Legal terms and conditions for using Decorebator
            </p>
            <p className="text-sm text-[#636E72] mt-4">
              Last updated: December 6, 2024
            </p>
          </div>

          <div className="bg-white rounded-2xl shadow-lg p-8 md:p-12">
            <PolicySection
              title="1. Acceptance of Terms"
              content={`
                <p class="mb-4">By accessing and using Decorebator ("the Service"), you accept and agree to be bound by the terms and provision of this agreement. If you do not agree to abide by the above, please do not use this service.</p>
                
                <p class="mb-4">These Terms of Service ("Terms") govern your use of our vocabulary learning application and related services operated by Decorebator ("us", "we", or "our").</p>
                
                <p class="mb-4">Your access to and use of the Service is conditioned on your acceptance of and compliance with these Terms. These Terms apply to all visitors, users, and others who access or use the Service.</p>
              `}
            />

            <PolicySection
              title="2. Description of Service"
              content={`
                <p class="mb-4">Decorebator is an AI-powered vocabulary learning platform that provides:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Vocabulary list creation and management in multiple languages</li>
                  <li>AI-generated definitions, example sentences, images, and audio pronunciations</li>
                  <li>Interactive quizzes and flashcards using spaced repetition (Leitner system)</li>
                  <li>Progress tracking and learning analytics</li>
                  <li>Premium subscription features including unlimited wordlists and offline access</li>
                  <li>Multi-language support for content generation</li>
                  <li>Error reporting system for AI-generated content quality control</li>
                </ul>
                
                <p class="mb-4">We reserve the right to modify, suspend, or discontinue the Service (or any part or content thereof) at any time without notice to you.</p>
              `}
            />

            <PolicySection
              title="3. User Accounts and Registration"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Account Creation</h4>
                <p class="mb-4">To access certain features of the Service, you must register for an account. When you create an account, you agree to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Provide accurate, current, and complete information</li>
                  <li>Maintain the security of your password and account</li>
                  <li>Accept responsibility for all activities that occur under your account</li>
                  <li>Notify us immediately of any unauthorized use of your account</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Age Requirements</h4>
                <p class="mb-4">You must be at least 13 years old to use the Service. Users between 13 and 18 should have parental consent before using the Service.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Account Termination</h4>
                <p class="mb-4">We may terminate or suspend your account and bar access to the Service immediately, without prior notice or liability, under our sole discretion, for any reason whatsoever including but not limited to a breach of the Terms.</p>
              `}
            />

            <PolicySection
              title="4. Subscription Plans and Billing"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Free Plan</h4>
                <p class="mb-4">The free plan includes limited access with the following restrictions:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>1 wordlist maximum</li>
                  <li>10 words per wordlist</li>
                  <li>Basic quiz modes only</li>
                  <li>Online-only access</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Premium Plans</h4>
                <p class="mb-4">Premium subscriptions are available with the following pricing:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Monthly Plan:</strong> $6.99 per month, billed monthly</li>
                  <li><strong>Annual Plan:</strong> $69.90 per year, billed annually</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Billing and Payment</h4>
                <p class="mb-4">Payment processing is handled securely through Stripe. By subscribing, you agree to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Pay all applicable fees and charges incurred through your account</li>
                  <li>Automatic renewal unless you cancel your subscription</li>
                  <li>Price changes with 30 days advance notice</li>
                  <li>No refunds for partial billing periods</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Cancellation</h4>
                <p class="mb-4">You may cancel your subscription at any time. Cancellations take effect at the end of the current billing period. No refunds will be provided for unused portions of your subscription.</p>
              `}
            />

            <PolicySection
              title="5. Acceptable Use Policy"
              content={`
                <p class="mb-4">You agree not to use the Service to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Violate any applicable laws or regulations</li>
                  <li>Infringe upon the rights of others</li>
                  <li>Transmit any harmful, offensive, or objectionable content</li>
                  <li>Attempt to gain unauthorized access to the Service or other user accounts</li>
                  <li>Use automated scripts or bots to access the Service</li>
                  <li>Reverse engineer, decompile, or disassemble any portion of the Service</li>
                  <li>Upload or share content that contains viruses or malicious code</li>
                  <li>Abuse the error reporting system with false or malicious reports</li>
                  <li>Share your account credentials with other users</li>
                </ul>
                
                <p class="mb-4">We reserve the right to investigate and take appropriate action against users who violate this policy, including account suspension or termination.</p>
              `}
            />

            <PolicySection
              title="6. Intellectual Property Rights"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Service Content</h4>
                <p class="mb-4">The Service and its original content, features, and functionality are owned by Decorebator and are protected by international copyright, trademark, patent, trade secret, and other intellectual property laws.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">User Content</h4>
                <p class="mb-4">You retain ownership of the content you create, including vocabulary lists and custom words. By using the Service, you grant us a limited license to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Store and process your content to provide the Service</li>
                  <li>Generate AI-powered enhancements (definitions, images, audio)</li>
                  <li>Improve our services and AI models (in aggregated, non-identifiable form)</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">AI-Generated Content</h4>
                <p class="mb-4">AI-generated content (definitions, images, audio) created for your wordlists becomes part of your user content. However, we retain the right to use aggregated patterns from this content to improve our services.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Third-Party Content</h4>
                <p class="mb-4">The Service may contain content from third-party sources, including AI providers. Such content is subject to the respective third-party terms and licenses.</p>
              `}
            />

            <PolicySection
              title="7. Privacy and Data Protection"
              content={`
                <p class="mb-4">Your privacy is important to us. Our Privacy Policy explains how we collect, use, and protect your information when you use the Service. By using the Service, you agree to the collection and use of information in accordance with our Privacy Policy.</p>
                
                <p class="mb-4">Key privacy points:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>We collect necessary information to provide and improve the Service</li>
                  <li>Your learning data is used to implement spaced repetition algorithms</li>
                  <li>We share data with trusted service providers (AI services, payment processors, email services)</li>
                  <li>You have rights to access, correct, and delete your personal data</li>
                  <li>EU users have additional rights under GDPR</li>
                </ul>
                
                <p class="mb-4">For complete details, please review our <a href="/privacy" class="text-orange-600 hover:text-orange-800 underline">Privacy Policy</a> and <a href="/gdpr" class="text-orange-600 hover:text-orange-800 underline">GDPR Compliance page</a>.</p>
              `}
            />

            <PolicySection
              title="8. AI-Generated Content and Accuracy"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Content Generation</h4>
                <p class="mb-4">Our Service uses artificial intelligence to generate educational content including definitions, example sentences, images, and audio pronunciations. While we strive for accuracy, AI-generated content may contain errors or inaccuracies.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Quality Control</h4>
                <p class="mb-4">We provide error reporting tools that allow users to report issues with AI-generated content. We use these reports to improve content quality, but we cannot guarantee the accuracy of all AI-generated material.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">User Responsibility</h4>
                <p class="mb-4">Users should verify important information and use the Service as a learning aid rather than a definitive source. We recommend consulting authoritative dictionaries and language resources for critical information.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Content Ownership</h4>
                <p class="mb-4">AI-generated content is created specifically for your learning needs and becomes part of your user content, subject to the intellectual property terms outlined above.</p>
              `}
            />

            <PolicySection
              title="9. Disclaimers and Limitation of Liability"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Service Disclaimer</h4>
                <p class="mb-4">The Service is provided on an "AS IS" and "AS AVAILABLE" basis without warranties of any kind, either express or implied, including but not limited to implied warranties of merchantability, fitness for a particular purpose, or non-infringement.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Educational Purpose</h4>
                <p class="mb-4">The Service is designed for educational purposes and language learning. We do not guarantee specific learning outcomes or proficiency levels.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Limitation of Liability</h4>
                <p class="mb-4">In no event shall Decorebator, its directors, employees, partners, agents, suppliers, or affiliates be liable for any indirect, incidental, special, consequential, or punitive damages, including without limitation, loss of profits, data, use, goodwill, or other intangible losses, resulting from your use of the Service.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Maximum Liability</h4>
                <p class="mb-4">Our total liability to you for all claims arising from or relating to the Service shall not exceed the amount you paid us in the twelve (12) months preceding the claim.</p>
              `}
            />

            <PolicySection
              title="10. Governing Law and Dispute Resolution"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Governing Law</h4>
                <p class="mb-4">These Terms shall be interpreted and governed by the laws of [Your Jurisdiction], without regard to its conflict of law provisions.</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Dispute Resolution</h4>
                <p class="mb-4">Any disputes arising from these Terms or your use of the Service shall be resolved through:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Good faith negotiation between the parties</li>
                  <li>Binding arbitration if negotiation fails</li>
                  <li>Jurisdiction in the courts of [Your Jurisdiction] for any disputes not subject to arbitration</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Class Action Waiver</h4>
                <p class="mb-4">You agree that any arbitration or proceeding shall be limited to the dispute between us and you individually. You waive any right to participate in a class action lawsuit or class-wide arbitration.</p>
              `}
            />

            <PolicySection
              title="11. Changes to Terms"
              content={`
                <p class="mb-4">We reserve the right to modify or replace these Terms at any time. If a revision is material, we will provide at least 30 days notice prior to any new terms taking effect.</p>
                
                <p class="mb-4">Material changes will be communicated through:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Email notification to registered users</li>
                  <li>Prominent notice in the mobile application</li>
                  <li>Updated "Last Modified" date on this page</li>
                </ul>
                
                <p class="mb-4">Your continued use of the Service after any changes indicates your acceptance of the new Terms.</p>
              `}
            />

            <PolicySection
              title="12. Contact Information"
              content={`
                <p class="mb-4">If you have any questions about these Terms of Service, please contact us:</p>
                
                <div class="bg-gray-50 p-6 rounded-lg mt-4">
                  <p class="mb-2"><strong>Email:</strong> legal@decorebator.com</p>
                  <p class="mb-2"><strong>Support:</strong> support@decorebator.com</p>
                  <p class="mb-2"><strong>Address:</strong> [Your Business Address]</p>
                  <p class="mb-2"><strong>Phone:</strong> [Your Phone Number]</p>
                </div>
                
                <p class="mb-4 mt-4">We will respond to your inquiry within 5 business days.</p>
              `}
            />

            <div className="mt-12 p-6 bg-orange-50 rounded-lg border border-orange-200">
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0">
                  <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-orange-900 mb-2">
                    Fair and Transparent Terms
                  </h3>
                  <p className="text-orange-800">
                    These terms are designed to protect both users and our service while enabling the best possible vocabulary learning experience. 
                    We believe in clear, fair policies that respect your rights as a learner.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </PageLayout>
  );
}
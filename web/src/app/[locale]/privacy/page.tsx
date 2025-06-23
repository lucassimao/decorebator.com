'use client';

import React from 'react';
import PageLayout from '../../../components/layout/PageLayout';
import { PolicySection } from '../../../components/policy/PolicySection';

export default function PrivacyPolicy() {
  return (
    <PageLayout>
      <div className="pt-24 pb-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h1 className="text-4xl md:text-5xl font-bold text-[#2D3436] mb-4">
              Privacy Policy
            </h1>
            <p className="text-xl text-[#636E72]">
              Your privacy is important to us
            </p>
            <p className="text-sm text-[#636E72] mt-4">
              Last updated: January 2025
            </p>
          </div>

          <div className="bg-white rounded-2xl shadow-lg p-8 md:p-12">
            <PolicySection
              title="Introduction"
              content={`
                <p class="mb-4">Welcome to Decorebator. This Privacy Policy explains how we collect, use, share, and protect your personal information when you use our vocabulary learning app. We are committed to protecting your privacy and ensuring transparency about our data practices.</p>
              `}
            />

            <PolicySection
              title="Information We Collect"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Personal Information</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Email address for account creation</li>
                  <li>Password (encrypted and never stored in plain text)</li>
                  <li>Display name (optional)</li>
                  <li>Subscription status and payment information (processed by Stripe)</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Learning Data</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Words and vocabulary lists you create</li>
                  <li>Learning progress and quiz performance</li>
                  <li>Spaced repetition schedule data</li>
                  <li>Practice time and learning streaks</li>
                  <li>Error reports for AI-generated content</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Technical Data</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Device type and operating system</li>
                  <li>App version and language settings</li>
                  <li>Anonymous usage analytics (PostHog)</li>
                  <li>Error logs and crash reports (Sentry)</li>
                  <li>IP address (for security purposes only)</li>
                </ul>
              `}
            />

            <PolicySection
              title="How We Use Your Information"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Core App Functionality</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Account creation and authentication</li>
                  <li>Synchronizing learning progress across devices</li>
                  <li>Delivering personalized spaced repetition schedules</li>
                  <li>Generating AI-powered vocabulary content</li>
                  <li>Ensuring content appropriateness through automated moderation</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">App Improvement</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Analyzing usage patterns to improve features</li>
                  <li>Identifying and fixing technical issues</li>
                  <li>Developing new educational features</li>
                  <li>Optimizing app performance</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Communication</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Sending subscription renewal notifications</li>
                  <li>Providing customer support</li>
                  <li>Sharing important app updates (with your consent)</li>
                </ul>
              `}
            />

            <PolicySection
              title="Data Sharing and Third Parties"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Service Providers</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>OpenAI</strong>: For generating vocabulary definitions, images, and audio</li>
                  <li><strong>Stripe</strong>: For secure payment processing (subscription management)</li>
                  <li><strong>SendGrid</strong>: For transactional email delivery</li>
                  <li><strong>Sentry</strong>: For error monitoring and crash reporting</li>
                  <li><strong>PostHog</strong>: For privacy-focused analytics</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Data Processing</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Vocabulary terms are sent to OpenAI for content generation</li>
                  <li>Payment information is processed securely by Stripe</li>
                  <li>Analytics data is anonymized and aggregated</li>
                  <li>No personal data is sold to third parties</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Legal Requirements</h4>
                <p class="mb-4">We may disclose information if required by law, court order, or to protect our rights and safety.</p>
              `}
            />

            <PolicySection
              title="Data Security"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Technical Safeguards</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>End-to-end encryption for sensitive data</li>
                  <li>Secure API communications using HTTPS</li>
                  <li>JWT tokens for authenticated sessions</li>
                  <li>Secure local storage on your device</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Access Controls</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Limited employee access to personal data</li>
                  <li>Regular security audits and updates</li>
                  <li>Data minimization practices</li>
                  <li>Secure development practices</li>
                </ul>
              `}
            />

            <PolicySection
              title="Your Rights and Choices"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Account Management</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Access</strong>: View your personal information in app settings</li>
                  <li><strong>Update</strong>: Modify your profile and preferences anytime</li>
                  <li><strong>Delete</strong>: Request complete account deletion</li>
                  <li><strong>Export</strong>: Download your learning data</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Privacy Controls</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Analytics</strong>: Opt out of usage analytics</li>
                  <li><strong>Notifications</strong>: Control email and push notifications</li>
                  <li><strong>Data Processing</strong>: Limit AI content generation features</li>
                  <li><strong>Location</strong>: No location data is collected</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Subscription Management</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Cancel</strong>: End subscriptions through app stores</li>
                  <li><strong>Refunds</strong>: Request refunds according to store policies</li>
                  <li><strong>Data Retention</strong>: Learning data preserved during subscription</li>
                </ul>
              `}
            />

            <PolicySection
              title="Data Retention"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Active Accounts</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Personal information: Retained while account is active</li>
                  <li>Learning progress: Preserved for educational continuity</li>
                  <li>Analytics data: Aggregated and anonymized after 90 days</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Deleted Accounts</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Personal information: Deleted within 30 days</li>
                  <li>Learning data: Permanently removed</li>
                  <li>Legal requirements: Some data may be retained as required by law</li>
                </ul>
              `}
            />

            <PolicySection
              title="Children's Privacy"
              content={`
                <p class="mb-4">Decorebator is suitable for all ages, including children under 13. For users under 13:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Parental consent is required for account creation</li>
                  <li>Limited data collection focused on educational features</li>
                  <li>No behavioral advertising or tracking</li>
                  <li>Enhanced privacy protections</li>
                </ul>
              `}
            />

            <PolicySection
              title="International Data Transfers"
              content={`
                <p class="mb-4">Your information may be transferred to and processed in countries other than your own, including:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>United States (OpenAI, Stripe, SendGrid)</li>
                  <li>European Union (data protection compliance)</li>
                  <li>Other regions where our service providers operate</li>
                </ul>
                
                <p class="mb-4">We ensure appropriate safeguards are in place for international transfers.</p>
              `}
            />

            <PolicySection
              title="Cookies and Tracking"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Local Storage</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>App preferences and settings</li>
                  <li>Authentication tokens</li>
                  <li>Offline learning data</li>
                  <li>No third-party tracking cookies</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Analytics</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Anonymized usage statistics</li>
                  <li>Performance monitoring</li>
                  <li>Error tracking</li>
                  <li>No personal identification</li>
                </ul>
              `}
            />

            <PolicySection
              title="Updates to Privacy Policy"
              content={`
                <p class="mb-4">We may update this Privacy Policy to reflect:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Changes in app features</li>
                  <li>Legal requirements</li>
                  <li>Improved privacy practices</li>
                  <li>User feedback and requests</li>
                </ul>
                
                <p class="mb-4">Significant changes will be communicated through:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>In-app notifications</li>
                  <li>Email updates (if opted in)</li>
                  <li>App Store update notes</li>
                </ul>
              `}
            />

            <PolicySection
              title="Regional Privacy Rights"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">European Union (GDPR)</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Right to access your personal data</li>
                  <li>Right to rectify inaccurate information</li>
                  <li>Right to erase your data</li>
                  <li>Right to restrict processing</li>
                  <li>Right to data portability</li>
                  <li>Right to object to processing</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">California (CCPA)</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Right to know what personal information is collected</li>
                  <li>Right to delete personal information</li>
                  <li>Right to opt-out of sale (we don't sell data)</li>
                  <li>Right to non-discrimination</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Other Regions</h4>
                <p class="mb-4">We comply with applicable privacy laws in all regions where the app is available.</p>
              `}
            />

            <PolicySection
              title="Contact Information"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Privacy Questions</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Email</strong>: privacy@decorebator.com</li>
                  <li><strong>Response Time</strong>: Within 72 hours</li>
                  <li><strong>Data Requests</strong>: Processed within 30 days</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">General Support</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>In-App</strong>: Settings > Help & Support</li>
                  <li><strong>Email</strong>: support@decorebator.com</li>
                  <li><strong>Website</strong>: https://decorebator.com/privacy</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Data Protection Officer</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Email</strong>: dpo@decorebator.com</li>
                  <li><strong>Role</strong>: Overseeing privacy compliance and user rights</li>
                </ul>
              `}
            />

            <PolicySection
              title="Transparency Report"
              content={`
                <p class="mb-4">We are committed to transparency about our privacy practices:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>No data sales to third parties</li>
                  <li>Minimal data collection for core functionality</li>
                  <li>User control over privacy settings</li>
                  <li>Regular privacy policy updates</li>
                  <li>Open communication about data practices</li>
                </ul>
              `}
            />

            <PolicySection
              title="Conclusion"
              content={`
                <p class="mb-4">Your privacy is fundamental to our mission of providing effective vocabulary learning. We are committed to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Protecting your personal information</li>
                  <li>Being transparent about our practices</li>
                  <li>Giving you control over your data</li>
                  <li>Continuously improving our privacy protections</li>
                </ul>
                
                <p class="mb-4">By using Decorebator, you agree to this Privacy Policy. Please review it regularly for updates and contact us with any questions or concerns.</p>
              `}
            />

            <div className="mt-12 p-6 bg-orange-50 rounded-lg border border-orange-200">
              <div className="flex items-start space-x-3">
                <div className="flex-shrink-0">
                  <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-orange-900 mb-2">
                    Your Privacy Matters
                  </h3>
                  <p className="text-orange-800">
                    We take your privacy seriously and are committed to protecting your personal information. If you have any questions or concerns about our privacy practices, please don&apos;t hesitate to contact us.
                  </p>
                  <p className="text-orange-800 mt-3">
                    Effective Date: June 1, 2025 | Version: 1.0 | Next Review: July 2026
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
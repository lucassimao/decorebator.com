
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
              How we collect, use, and protect your personal information
            </p>
            <p className="text-sm text-[#636E72] mt-4">
              Last updated: December 6, 2024
            </p>
          </div>

          <div className="bg-white rounded-2xl shadow-lg p-8 md:p-12">
            <PolicySection
              title="1. Information We Collect"
              content={`
                <p class="mb-4">We collect information you provide directly to us, such as when you:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Create an account and provide your email address, name, and password</li>
                  <li>Complete your user profile with optional information like country, date of birth, and profile picture</li>
                  <li>Create vocabulary lists and add words to study</li>
                  <li>Use our quiz and flashcard features</li>
                  <li>Subscribe to our premium services</li>
                  <li>Contact us for support or feedback</li>
                  <li>Report errors in AI-generated content</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Automatically Collected Information</h4>
                <p class="mb-4">When you use our services, we automatically collect certain information, including:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Learning progress and quiz performance data</li>
                  <li>Usage analytics and app interaction patterns</li>
                  <li>Device information and operating system details</li>
                  <li>IP address and general location information</li>
                  <li>Log data including access times and pages viewed</li>
                </ul>
              `}
            />

            <PolicySection
              title="2. How We Use Your Information"
              content={`
                <p class="mb-4">We use the information we collect to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Provide, maintain, and improve our vocabulary learning services</li>
                  <li>Generate personalized AI-powered content (definitions, images, audio)</li>
                  <li>Track your learning progress and implement spaced repetition algorithms</li>
                  <li>Process subscription payments and manage your account</li>
                  <li>Send important service announcements and subscription notifications</li>
                  <li>Provide customer support and respond to your inquiries</li>
                  <li>Improve our AI models and content quality based on error reports</li>
                  <li>Analyze usage patterns to enhance user experience</li>
                  <li>Comply with legal obligations and protect our rights</li>
                </ul>
              `}
            />

            <PolicySection
              title="3. Information Sharing and Disclosure"
              content={`
                <p class="mb-4">We do not sell, trade, or otherwise transfer your personal information to third parties, except in the following circumstances:</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Service Providers</h4>
                <p class="mb-4">We share information with trusted third-party service providers who assist us in operating our services:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>AI Providers:</strong> For generating definitions, images, and audio pronunciations</li>
                  <li><strong>Stripe:</strong> For processing subscription payments securely</li>
                  <li><strong>SendGrid:</strong> For sending transactional emails</li>
                  <li><strong>Cloud Storage:</strong> For storing user-generated content and AI-generated media</li>
                  <li><strong>Analytics Providers:</strong> For understanding app usage and improving our services</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Legal Requirements</h4>
                <p class="mb-4">We may disclose your information when required by law or to:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Comply with legal process or government requests</li>
                  <li>Protect our rights, property, or safety</li>
                  <li>Investigate fraud or security issues</li>
                  <li>Enforce our Terms of Service</li>
                </ul>
              `}
            />

            <PolicySection
              title="4. Data Security and Storage"
              content={`
                <p class="mb-4">We implement robust security measures to protect your personal information:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Encryption:</strong> All data is encrypted in transit using TLS/SSL protocols</li>
                  <li><strong>Password Security:</strong> Passwords are hashed using industry-standard bcrypt</li>
                  <li><strong>JWT Authentication:</strong> Secure token-based authentication for API access</li>
                  <li><strong>Database Security:</strong> PostgreSQL with proper access controls and monitoring</li>
                  <li><strong>Secure Storage:</strong> Mobile tokens stored in device Keychain/Keystore</li>
                  <li><strong>Regular Security Audits:</strong> Ongoing security assessments and updates</li>
                </ul>
                
                <p class="mb-4 mt-6">Your data is stored on secure servers in data centers with physical and electronic security measures. We retain your information for as long as your account is active or as needed to provide services.</p>
              `}
            />

            <PolicySection
              title="5. Your Rights and Choices"
              content={`
                <p class="mb-4">You have the following rights regarding your personal information:</p>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Account Management</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Access:</strong> View and update your profile information at any time</li>
                  <li><strong>Correction:</strong> Modify incorrect or outdated personal data</li>
                  <li><strong>Deletion:</strong> Delete your account and associated data</li>
                  <li><strong>Data Export:</strong> Request a copy of your learning data</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Communication Preferences</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Opt out of marketing communications (subscription emails will continue)</li>
                  <li>Manage notification preferences in your account settings</li>
                  <li>Contact us to update your communication preferences</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">EU Residents (GDPR)</h4>
                <p class="mb-4">If you are located in the European Union, you have additional rights under GDPR. Please see our <a href="/gdpr" class="text-orange-600 hover:text-orange-800 underline">GDPR Compliance page</a> for detailed information.</p>
              `}
            />

            <PolicySection
              title="6. Children's Privacy"
              content={`
                <p class="mb-4">Our services are not intended for children under 13 years of age. We do not knowingly collect personal information from children under 13.</p>
                
                <p class="mb-4">If you are a parent or guardian and believe your child has provided us with personal information, please contact us immediately. We will take steps to remove such information and terminate the child's account.</p>
                
                <p class="mb-4">For users between 13 and 18 years old, we recommend parental guidance when using our services.</p>
              `}
            />

            <PolicySection
              title="7. International Data Transfers"
              content={`
                <p class="mb-4">Your information may be transferred to and processed in countries other than your own. These countries may have different data protection laws than your country.</p>
                
                <p class="mb-4">When we transfer personal information internationally, we ensure appropriate safeguards are in place, including:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Standard contractual clauses approved by relevant authorities</li>
                  <li>Adequacy decisions by the European Commission</li>
                  <li>Other appropriate safeguards to protect your data</li>
                </ul>
              `}
            />

            <PolicySection
              title="8. Cookies and Tracking Technologies"
              content={`
                <p class="mb-4">We use cookies and similar tracking technologies to enhance your experience. For detailed information about our cookie practices, please see our <a href="/cookies" class="text-orange-600 hover:text-orange-800 underline">Cookie Policy</a>.</p>
                
                <p class="mb-4">You can control cookie settings through your browser preferences, though some features may not function properly if cookies are disabled.</p>
              `}
            />

            <PolicySection
              title="9. Changes to This Privacy Policy"
              content={`
                <p class="mb-4">We may update this Privacy Policy from time to time to reflect changes in our practices or applicable law. We will notify you of significant changes by:</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Posting the updated policy on our website with a new "Last Updated" date</li>
                  <li>Sending an email notification for material changes</li>
                  <li>Displaying a prominent notice in our mobile application</li>
                </ul>
                
                <p class="mb-4">Your continued use of our services after any changes indicates your acceptance of the updated Privacy Policy.</p>
              `}
            />

            <PolicySection
              title="10. Contact Us"
              content={`
                <p class="mb-4">If you have any questions about this Privacy Policy or our privacy practices, please contact us:</p>
                
                <div class="bg-gray-50 p-6 rounded-lg mt-4">
                  <p class="mb-2"><strong>Email:</strong> privacy@decorebator.com</p>
                  <p class="mb-2"><strong>Data Protection Officer:</strong> dpo@decorebator.com</p>
                  <p class="mb-2"><strong>Address:</strong> [Your Business Address]</p>
                  <p class="mb-2"><strong>Phone:</strong> [Your Phone Number]</p>
                </div>
                
                <p class="mb-4 mt-4">We will respond to your inquiry within 30 days (or sooner as required by applicable law).</p>
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
                    We are committed to protecting your privacy and being transparent about our data practices. 
                    This policy explains how we handle your information to provide you with the best vocabulary learning experience.
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
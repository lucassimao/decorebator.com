
'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import PageLayout from '../../../components/layout/PageLayout';
import { PolicySection } from '../../../components/policy/PolicySection';

export default function TermsOfService() {
  const t = useTranslations('terms');
  
  return (
    <PageLayout>
      <div className="pt-24 pb-20">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h1 className="text-4xl md:text-5xl font-bold text-[#2D3436] mb-4">
              {t('title')}
            </h1>
            <p className="text-xl text-[#636E72]">
              {t('subtitle')}
            </p>
            <p className="text-sm text-[#636E72] mt-4">
              {t('lastUpdated')}
            </p>
          </div>

          <div className="bg-white rounded-2xl shadow-lg p-8 md:p-12">
            <PolicySection
              title={t('sections.1.title')}
              content={`
                <p class="mb-4">${t('sections.1.content')}</p>
              `}
            />

            <PolicySection
              title={t('sections.2.title')}
              content={`
                <p class="mb-4">${t('sections.2.intro')}</p>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  ${t('sections.2.features').map((feature: string) => `<li>${feature}</li>`).join('')}
                </ul>
              `}
            />

            <PolicySection
              title={t('sections.3.title')}
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">${t('sections.3.accountCreation.title')}</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  ${t('sections.3.accountCreation.items').map((item: string) => `<li>${item}</li>`).join('')}
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">${t('sections.3.accountResponsibilities.title')}</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  ${t('sections.3.accountResponsibilities.items').map((item: string) => `<li>${item}</li>`).join('')}
                </ul>
              `}
            />

            <PolicySection
              title="4. Subscription Plans"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Free Plan</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>1 wordlist maximum</li>
                  <li>Up to 10 words per wordlist</li>
                  <li>Basic quiz modes</li>
                  <li>Limited features</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Premium Plans</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li><strong>Monthly</strong>: $6.99/month</li>
                  <li><strong>Annual</strong>: $69.90/year</li>
                  <li>Unlimited wordlists and words</li>
                  <li>Advanced analytics</li>
                  <li>Priority support</li>
                  <li>All features included</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Billing Terms</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Subscriptions auto-renew unless cancelled</li>
                  <li>Payment processed through App Store/Google Play</li>
                  <li>Refunds subject to platform policies</li>
                  <li>Price changes with 30-day notice</li>
                </ul>
              `}
            />

            <PolicySection
              title="5. Acceptable Use"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Permitted Uses</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Personal vocabulary learning</li>
                  <li>Educational purposes</li>
                  <li>Creating appropriate vocabulary content</li>
                  <li>Sharing wordlists with proper content</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Prohibited Uses</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Uploading offensive, illegal, or inappropriate content</li>
                  <li>Attempting to reverse engineer the app</li>
                  <li>Using automated tools or bots</li>
                  <li>Violating intellectual property rights</li>
                  <li>Sharing account credentials</li>
                  <li>Commercial use without permission</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Content Guidelines</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Keep vocabulary appropriate for all ages</li>
                  <li>No spam, malware, or harmful content</li>
                  <li>Respect copyright and intellectual property</li>
                  <li>Report inappropriate content</li>
                  <li>All user-generated content is subject to automated moderation</li>
                  <li>Content that violates guidelines will be automatically rejected</li>
                  <li>Repeated violations may result in account restrictions</li>
                </ul>
              `}
            />

            <PolicySection
              title="6. User-Generated Content"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Ownership</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>You retain ownership of wordlists you create</li>
                  <li>You grant us license to process and display your content</li>
                  <li>We may use anonymized data for app improvement</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Content Responsibility</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>You are responsible for content you upload</li>
                  <li>We may remove inappropriate content</li>
                  <li>Repeated violations may result in account suspension</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Automated Content Moderation</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>All wordlists and vocabulary words are automatically screened for appropriateness</li>
                  <li>Content is evaluated for educational value, profanity, and safety</li>
                  <li>Multi-language filtering ensures global community standards</li>
                  <li>Content status (approved, pending, rejected) is tracked and logged</li>
                  <li>Appeals process available for incorrectly flagged content</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">AI-Generated Content</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>AI-generated definitions, images, and audio are provided as-is</li>
                  <li>Report errors using in-app reporting system</li>
                  <li>We continuously improve AI content quality</li>
                </ul>
              `}
            />

            <PolicySection
              title="7. Privacy and Data Protection"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Data Collection</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>We collect minimal data necessary for app functionality</li>
                  <li>Detailed practices outlined in Privacy Policy</li>
                  <li>You can request data deletion at any time</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Data Security</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>We implement industry-standard security measures</li>
                  <li>Regular security audits and updates</li>
                  <li>Prompt notification of any security incidents</li>
                </ul>
              `}
            />

            <PolicySection
              title="8. Intellectual Property"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Our Rights</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Decorebator app and brand are our property</li>
                  <li>AI-generated content created through our systems</li>
                  <li>All improvements and enhancements to the app</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Your Rights</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Vocabulary wordlists you create</li>
                  <li>Personal learning progress data</li>
                  <li>Feedback and suggestions you provide</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Third-Party Content</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Some content may be generated using third-party AI services</li>
                  <li>Appropriate licenses and attributions are maintained</li>
                  <li>Report any copyright concerns promptly</li>
                </ul>
              `}
            />

            <PolicySection
              title="9. Disclaimers and Limitations"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Service Availability</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>App provided "as is" without warranties</li>
                  <li>We may temporarily suspend service for maintenance</li>
                  <li>No guarantee of uninterrupted access</li>
                  <li>Features may change with updates</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Educational Disclaimer</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>App is a learning tool, not a comprehensive language course</li>
                  <li>Results may vary based on individual usage</li>
                  <li>Supplement with other learning methods as needed</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">AI Content Disclaimer</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>AI-generated content may contain errors</li>
                  <li>Use error reporting system for corrections</li>
                  <li>We continuously improve content accuracy</li>
                </ul>
              `}
            />

            <PolicySection
              title="10. Limitation of Liability"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Scope of Liability</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Our liability limited to subscription fees paid</li>
                  <li>No liability for indirect or consequential damages</li>
                  <li>Force majeure events beyond our control excluded</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">User Responsibility</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>You use the app at your own risk</li>
                  <li>Responsible for your own learning outcomes</li>
                  <li>Maintain appropriate device security</li>
                </ul>
              `}
            />

            <PolicySection
              title="11. Termination"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Termination by You</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Cancel subscription through app stores</li>
                  <li>Delete account through app settings</li>
                  <li>Data deletion within 30 days of account closure</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Termination by Us</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>We may suspend accounts for Terms violations</li>
                  <li>30-day notice for non-violation terminations</li>
                  <li>Immediate termination for serious violations</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Effect of Termination</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Access to app and data ends</li>
                  <li>Subscriptions cancelled according to platform policies</li>
                  <li>Certain provisions survive termination</li>
                </ul>
              `}
            />

            <PolicySection
              title="12. Geographic Restrictions"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Availability</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>App available in countries where legally permitted</li>
                  <li>Some features may vary by region</li>
                  <li>Compliance with local laws required</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Export Controls</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Software subject to export control laws</li>
                  <li>Users responsible for compliance with local regulations</li>
                </ul>
              `}
            />

            <PolicySection
              title="13. Updates and Changes"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">App Updates</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Regular updates to improve functionality</li>
                  <li>Security updates may be mandatory</li>
                  <li>New features may be added or removed</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Terms Changes</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Terms may be updated periodically</li>
                  <li>Significant changes communicated through app</li>
                  <li>Continued use constitutes acceptance</li>
                </ul>
              `}
            />

            <PolicySection
              title="14. Third-Party Services"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Payment Processing</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Stripe processes subscription payments</li>
                  <li>Subject to Stripe's terms and policies</li>
                  <li>We are not responsible for payment processor issues</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">AI Services</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>OpenAI powers content generation features</li>
                  <li>Subject to OpenAI's usage policies</li>
                  <li>We monitor for appropriate use</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">App Stores</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Distribution through Apple App Store and Google Play</li>
                  <li>Subject to platform terms and policies</li>
                  <li>Platform-specific features and limitations apply</li>
                </ul>
              `}
            />

            <PolicySection
              title="15. Support and Contact"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Customer Support</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>In-app support available in Settings</li>
                  <li>Email support: support@decorebator.com</li>
                  <li>Response within 72 hours for Premium users</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Technical Issues</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Report bugs through in-app system</li>
                  <li>Regular bug fixes in app updates</li>
                  <li>Community support resources available</li>
                </ul>
              `}
            />

            <PolicySection
              title="16. Governing Law"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Jurisdiction</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Terms governed by laws of [Jurisdiction]</li>
                  <li>Disputes resolved through binding arbitration</li>
                  <li>Class action waiver applies</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">International Users</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Local laws may provide additional rights</li>
                  <li>Terms adapted to comply with regional requirements</li>
                  <li>Contact us for region-specific questions</li>
                </ul>
              `}
            />

            <PolicySection
              title="17. Accessibility"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Commitment</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>App designed with accessibility in mind</li>
                  <li>Screen reader support and keyboard navigation</li>
                  <li>Continuous improvements based on user feedback</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Assistance</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Contact support for accessibility questions</li>
                  <li>Alternative formats available upon request</li>
                  <li>Reasonable accommodations provided</li>
                </ul>
              `}
            />

            <PolicySection
              title="18. Educational Use"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Schools and Institutions</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Special educational licenses available</li>
                  <li>Bulk pricing for institutional use</li>
                  <li>Student privacy protections enhanced</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Parental Controls</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Parents can manage children's accounts</li>
                  <li>Enhanced privacy for users under 13</li>
                  <li>Educational progress sharing with parents</li>
                </ul>
              `}
            />

            <PolicySection
              title="19. Security"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Data Protection</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Regular security assessments</li>
                  <li>Prompt security incident response</li>
                  <li>User notification of significant threats</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">User Security</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Enable device security features</li>
                  <li>Use strong, unique passwords</li>
                  <li>Report suspicious activity immediately</li>
                </ul>
              `}
            />

            <PolicySection
              title="20. Miscellaneous"
              content={`
                <h4 class="text-lg font-semibold mb-3 mt-6">Entire Agreement</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>These Terms constitute the complete agreement</li>
                  <li>Supersede all previous agreements</li>
                  <li>May only be modified in writing</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Severability</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>Invalid provisions do not affect remaining Terms</li>
                  <li>Terms interpreted to maximum extent permitted by law</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Assignment</h4>
                <ul class="list-disc pl-6 mb-4 space-y-2">
                  <li>We may assign rights and obligations</li>
                  <li>You may not assign without our consent</li>
                  <li>Terms bind successors and assigns</li>
                </ul>
                
                <h4 class="text-lg font-semibold mb-3 mt-6">Contact Information</h4>
                <div class="bg-gray-50 p-6 rounded-lg mt-4">
                  <p class="mb-2"><strong>Legal:</strong> legal@decorebator.com</p>
                  <p class="mb-2"><strong>General:</strong> support@decorebator.com</p>
                  <p class="mb-2"><strong>Privacy:</strong> privacy@decorebator.com</p>
                </div>
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
                    Acknowledgment
                  </h3>
                  <p className="text-orange-800">
                    By using Decorebator, you acknowledge that you have read, understood, and agree to be bound by these Terms of Service and our Privacy Policy.
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
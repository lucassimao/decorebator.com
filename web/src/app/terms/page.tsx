
import React from 'react';
import TosSection from '../../components/tos/TosSection';
import { Logo } from '../../components/home/icons';
import Button from '../../components/home/Button';
import Link from 'next/link';

const TermsOfServicePage: React.FC = () => {
  return (
    <div className="min-h-screen bg-slate-50 text-slate-800 font-sans selection:bg-blue-200 selection:text-blue-700">
      <header className="py-6 bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center">
          <Link href="/" aria-label="Back to Home page">
            <Logo className="h-8" />
          </Link>
        </div>
      </header>
      
      <main className="max-w-3xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <h1 className="text-4xl font-bold text-slate-900 mb-10 text-center">Terms of Service</h1>

        <TosSection title="1. Acceptance of Terms">
          <p className="mb-4">
            By accessing or using the Decorebator application (the &quot;App&quot;) and any related services (collectively, the &quot;Services&quot;), 
            you agree to be bound by these Terms of Service (&quot;Terms&quot;). If you do not agree to all of these Terms, do not use the Services.
          </p>
          <p>
            We may modify the Terms at any time, and such modifications will be effective immediately upon posting. 
            You agree to review the Terms periodically to be aware of such modifications.
          </p>
        </TosSection>

        <TosSection title="2. Use of the App">
          <p className="mb-2"><strong>User Accounts:</strong> You may be required to create an account to access certain features of the App. You are responsible for maintaining the confidentiality of your account information and for all activities that occur under your account.</p>
          <p className="mb-2"><strong>Acceptable Use:</strong> You agree not to use the Services for any unlawful purpose or in any way that could harm, disable, overburden, or impair the Services. You agree not to attempt to gain unauthorized access to any part of the Services or to any other user&pos;s account.</p>
        </TosSection>

        <TosSection title="3. User Content">
          <p className="mb-4">
            If you create, upload, or share content (e.g., custom wordlists) through the App (&quot;User Content&quot;), you retain ownership of your User Content. 
            However, by submitting User Content, you grant Decorebator a worldwide, non-exclusive, royalty-free license to use, reproduce, modify, and distribute your User Content solely for the purpose of operating and providing the Services to you.
          </p>
        </TosSection>

        <TosSection title="4. AI-Generated Content">
          <p className="mb-4">
            The App utilizes Artificial Intelligence (AI) to generate content such as word meanings, example sentences, images, and audio pronunciations. 
            While we strive for accuracy, AI-generated content may occasionally be incorrect, incomplete, or reflect biases from the data it was trained on. 
            Decorebator does not guarantee the accuracy or reliability of AI-generated content and is not liable for any errors or omissions.
          </p>
          <p>
            You understand that AI-generated content is provided for educational and informational purposes only and should not be solely relied upon.
          </p>
        </TosSection>

        <TosSection title="5. Subscription and Payment">
          <p className="mb-4">
            Certain features of the App may be available through a subscription plan (&quot;Subscription&quot;). Details of current subscription plans and pricing are available within the App or on our Website.
            By purchasing a Subscription, you agree to pay the applicable fees. All fees are non-refundable except as required by law or as explicitly stated in our refund policy.
          </p>
          <p>
            Subscriptions may renew automatically unless cancelled. You can manage or cancel your Subscription through your account settings or the respective app store.
          </p>
        </TosSection>

        <TosSection title="6. Intellectual Property">
          <p className="mb-4">
            The Services and all materials therein, including software, text, graphics, logos, and AI models (excluding User Content you provide), are the property of Decorebator or its licensors and are protected by copyright, trademark, and other intellectual property laws.
          </p>
        </TosSection>

        <TosSection title="7. Disclaimers and Limitation of Liability">
          <p className="mb-2">THE SERVICES ARE PROVIDED &quot;AS IS&quot; AND &quot;AS AVAILABLE&quot; WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING, BUT NOT LIMITED TO, IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.</p>
          <p className="mb-2">DECOREBATOR DOES NOT WARRANT THAT THE SERVICES WILL BE UNINTERRUPTED, ERROR-FREE, OR FREE OF VIRUSES OR OTHER HARMFUL COMPONENTS.</p>
          <p>TO THE FULLEST EXTENT PERMITTED BY LAW, IN NO EVENT SHALL DECOREBATOR BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES ARISING OUT OF OR RELATED TO YOUR USE OF THE SERVICES.</p>
        </TosSection>

        <TosSection title="8. Termination">
          <p className="mb-4">
            We may terminate or suspend your access to the Services, without prior notice or liability, for any reason, including if you breach these Terms. 
            Upon termination, your right to use the Services will immediately cease.
          </p>
        </TosSection>

        <TosSection title="9. Governing Law">
          <p className="mb-4">
            These Terms shall be governed and construed in accordance with the laws of [Your Jurisdiction, e.g., State of California, USA], without regard to its conflict of law provisions.
          </p>
        </TosSection>

        <TosSection title="10. Changes to Terms">
          <p className="mb-4">
            We reserve the right, at our sole discretion, to modify or replace these Terms at any time. If a revision is material, we will provide at least 30 days&apos; notice prior to any new terms taking effect. 
            What constitutes a material change will be determined at our sole discretion.
          </p>
          <p className="mt-2"><em>Last updated: {new Date().toLocaleDateString()}</em></p>
        </TosSection>

        <TosSection title="11. Contact Information">
          <p>
            If you have any questions about these Terms, please contact us at terms@decorebator.com or by post to:
          </p>
          <p className="mt-2">
            Decorebator Project<br />
            [Your Address, if applicable, or &quot;Online Service&quot;]
          </p>
        </TosSection>

        <div className="mt-12 text-center">
          <Button href="/" variant="outline" size="md">
            Back to Home
          </Button>
        </div>
      </main>

       <footer className="bg-slate-100 border-t border-slate-200 text-slate-600 py-8 mt-16">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 text-center text-sm">
          <p>&copy; {new Date().getFullYear()} Decorebator. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
};

export default TermsOfServicePage;
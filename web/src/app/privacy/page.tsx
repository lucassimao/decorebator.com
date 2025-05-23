
import React from 'react';
import PolicySection from '../../components/policy/PolicySection'; // Updated import path
import { Logo } from '../../components/home/icons'; // Updated import path
import Button from '../../components/home/Button'; // Updated import path
import Link from 'next/link';

const PolicyPage: React.FC = () => {
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
        <h1 className="text-4xl font-bold text-slate-900 mb-10 text-center">Privacy Policy</h1>

        <PolicySection title="1. Introduction">
          <p className="mb-4">
            Welcome to Decorebator! We are committed to protecting your personal information and your right to privacy. 
            If you have any questions or concerns about this privacy notice, or our practices with regards to your personal information, 
            please contact us at privacy@decorebator.com.
          </p>
          <p>
            This privacy notice describes how we might use your information if you:
          </p>
          <ul className="list-disc list-inside my-2 ml-4">
            <li>Visit our website at [Your Website URL]</li>
            <li>Download and use our mobile application — Decorebator</li>
            <li>Engage with us in other related ways ― including any sales, marketing, or events</li>
          </ul>
          <p>
            In this privacy notice, if we refer to:
            "App," we are referring to any application of ours that references or links to this policy, including any listed above.
            "Website," we are referring to our website.
            "Services," we are referring to our App, Website, and other related services, including any sales, marketing, or events.
          </p>
        </PolicySection>

        <PolicySection title="2. Information We Collect">
          <p className="mb-4">
            <strong>Personal information you disclose to us:</strong> We collect personal information that you voluntarily provide to us when you 
            register on the Services, express an interest in obtaining information about us or our products and Services, when you participate 
            in activities on the Services (such as posting messages in our online forums or entering competitions, contests or giveaways) 
            or otherwise when you contact us.
          </p>
          <p className="mb-4">
            The personal information that we collect depends on the context of your interactions with us and the Services, the choices you make 
            and the products and features you use. The personal information we collect may include the following: names; email addresses; 
            usernames; passwords; contact preferences; contact or authentication data; billing addresses; debit/credit card numbers.
          </p>
           <p>
            <strong>Information automatically collected:</strong> We automatically collect certain information when you visit, use or navigate the 
            Services. This information does not reveal your specific identity (like your name or contact information) but may include device 
            and usage information, such as your IP address, browser and device characteristics, operating system, language preferences, 
            referring URLs, device name, country, location, information about how and when you use our Services and other technical information.
          </p>
        </PolicySection>

        <PolicySection title="3. How We Use Your Information">
          <p className="mb-4">
            We use personal information collected via our Services for a variety of business purposes described below. We process your personal 
            information for these purposes in reliance on our legitimate business interests, in order to enter into or perform a contract with you, 
            with your consent, and/or for compliance with our legal obligations.
          </p>
          <ul className="list-disc list-inside my-2 ml-4">
            <li>To facilitate account creation and logon process.</li>
            <li>To post testimonials.</li>
            <li>Request feedback.</li>
            <li>To enable user-to-user communications.</li>
            <li>To manage user accounts.</li>
            <li>To send administrative information to you.</li>
            {/* Add more uses as needed */}
          </ul>
        </PolicySection>
        
        <PolicySection title="4. Will Your Information Be Shared With Anyone?">
          <p className="mb-4">
            We only share information with your consent, to comply with laws, to provide you with services, to protect your rights, or to fulfill business obligations.
          </p>
          {/* Add more details on sharing if applicable */}
        </PolicySection>

        <PolicySection title="5. Data Security">
          <p>
            We have implemented appropriate technical and organizational security measures designed to protect the security of any personal 
            information we process. However, despite our safeguards and efforts to secure your information, no electronic transmission over 
            the Internet or information storage technology can be guaranteed to be 100% secure, so we cannot promise or guarantee that hackers, 
            cybercriminals, or other unauthorized third parties will not be able to defeat our security, and improperly collect, access, steal, 
            or modify your information.
          </p>
        </PolicySection>

        <PolicySection title="6. Your Privacy Rights">
          <p>
            In some regions (like the European Economic Area and UK), you have certain rights under applicable data protection laws. These may 
            include the right (i) to request access and obtain a copy of your personal information, (ii) to request rectification or erasure; 
            (iii) to restrict the processing of your personal information; and (iv) if applicable, to data portability. In certain circumstances, 
            you may also have the right to object to the processing of your personal information.
          </p>
        </PolicySection>

        <PolicySection title="7. Changes to This Privacy Policy">
          <p>
            We may update this privacy notice from time to time. The updated version will be indicated by an updated "Revised" date and the 
            updated version will be effective as soon as it is accessible. We encourage you to review this privacy notice frequently to be 
            informed of how we are protecting your information.
          </p>
          <p className="mt-2"><em>Last updated: {new Date().toLocaleDateString()}</em></p>
        </PolicySection>

        <PolicySection title="8. Contact Us">
          <p>
            If you have questions or comments about this notice, you may email us at privacy@decorebator.com or by post to:
          </p>
          <p className="mt-2">
            [Your Company Name, if applicable]<br />
            [Your Company Address, if applicable]
          </p>
        </PolicySection>

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

export default PolicyPage;
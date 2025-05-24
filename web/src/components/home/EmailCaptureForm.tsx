'use client'
import React, { useState } from 'react';
import Button from './Button';
import { MailIcon } from './icons';

const EmailCaptureForm: React.FC = () => {
  const [email, setEmail] = useState('');

  const handleEmailSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // In a real app, you would handle the email submission here (e.g., API call)
    console.log('Email submitted for newsletter:', email);
    alert(`Thank you for joining the newsletter! We'll keep you updated at ${email}`);
    setEmail('');
  };

  return (
    <section id="newsletter" className="py-16 px-4 sm:px-6 lg:px-8" aria-labelledby="newsletter-heading">
      <div className="max-w-xl mx-auto text-center">
        <MailIcon aria-hidden="true" className="h-12 w-12 text-blue-600 mx-auto mb-4" />
        <h2 id="newsletter-heading" className="text-3xl font-bold text-slate-900 mb-3">Stay Ahead of the Curve</h2>
        <p className="text-slate-600 mb-8">
          Join our newsletter for learning tips, app updates, and early access to new features.
        </p>
        <form onSubmit={handleEmailSubmit} className="flex flex-col sm:flex-row gap-3 max-w-md mx-auto">
          <label htmlFor="email-newsletter-input" className="sr-only">Email for newsletter</label>
          <input
            id="email-newsletter-input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Enter your email"
            required
            aria-required="true"
            className="flex-grow p-3 border border-slate-300 rounded-2xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none shadow-sm"
          />
          <Button type="submit" variant="primary" size="md" className="sm:w-auto w-full">
            Subscribe
          </Button>
        </form>
      </div>
    </section>
  );
};

export default EmailCaptureForm;
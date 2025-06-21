'use client';

import React from 'react';
import { useTranslations } from 'next-intl';
import { FacebookIcon, TwitterIcon, InstagramIcon, LinkedInIcon } from './icons';
import Link from 'next/link';

const FooterSection: React.FC = () => {
  const t = useTranslations('footer');
  return (
    <footer className="bg-slate-800 text-slate-400 py-12 mt-16" aria-labelledby="footer-heading">
      <h2 id="footer-heading" className="sr-only">Footer</h2>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">Decorebator</h5>
            <ul className="space-y-2">
              <li><a href="#features" className="hover:text-slate-100">Features</a></li>
              <li><a href="#how-it-works" className="hover:text-slate-100">How It Works</a></li>
              <li><a href="#pricing" className="hover:text-slate-100">Pricing</a></li>
              <li><a href="#faq" className="hover:text-slate-100">FAQ</a></li>
            </ul>
          </div>
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">Legal</h5>
            <ul className="space-y-2">
              <li><Link href="/privacy" className="hover:text-slate-100">Privacy Policy</Link></li>
              <li><Link href="/terms" className="hover:text-slate-100">Terms of Service</Link></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-slate-700 pt-8 flex flex-col sm:flex-row justify-between items-center">
          <p className="text-sm">&copy; {new Date().getFullYear()} Decorebator. All rights reserved.</p>
          <div className="flex space-x-4 mt-4 sm:mt-0">
            <a href="#" className="hover:text-slate-100" aria-label="Decorebator on Facebook"><FacebookIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label="Decorebator on Twitter"><TwitterIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label="Decorebator on Instagram"><InstagramIcon className="h-6 w-6" /></a>
            <a href="#" className="hover:text-slate-100" aria-label="Decorebator on LinkedIn"><LinkedInIcon className="h-6 w-6" /></a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default FooterSection;
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
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-8">
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">Decorebator</h5>
            <ul className="space-y-2">
              <li><a href="#features" className="hover:text-slate-100">Features</a></li>
              <li><a href="#how-it-works" className="hover:text-slate-100">How It Works</a></li>
              <li><a href="#pricing" className="hover:text-slate-100">Pricing</a></li> {/* Added pricing link */}
            </ul>
          </div>
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">Resources</h5>
            <ul className="space-y-2">
              <li><a href="#" className="hover:text-slate-100">Blog</a></li>
              <li><Link href="/help" className="hover:text-slate-100">Help Center</Link></li>
              <li><Link href="/#faq" className="hover:text-slate-100">FAQ</Link></li>
            </ul>
          </div>
          <div>
            <h5 className="text-slate-200 font-semibold mb-3">{t('about.title', { fallback: 'About' })}</h5>
            <div className="space-y-3">
              <div className="flex items-center space-x-3">
                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-[#FF7B54] to-orange-600 flex items-center justify-center overflow-hidden">
                  {/* Placeholder for profile picture - replace with actual image */}
                  <i className="fas fa-user text-white text-lg"></i>
                </div>
                <div>
                  <p className="text-slate-200 text-sm font-medium">{t('about.soloEntrepreneur', { fallback: 'Solo Entrepreneur' })}</p>
                  <p className="text-slate-400 text-xs">{t('about.location', { fallback: 'Brazil 🇧🇷' })}</p>
                </div>
              </div>
              <p className="text-xs text-slate-400 leading-relaxed">
                {t('about.description', { fallback: 'Responsible for the architecture, full implementation, and design of this project. Solving a personal problem and thought it would be interesting to share with others.' })}
              </p>
              <a 
                href="https://twitter.com/yourusername" 
                target="_blank" 
                rel="noopener noreferrer"
                className="inline-flex items-center space-x-2 text-xs text-[#FF7B54] hover:text-orange-400 transition-colors duration-200"
              >
                <i className="fab fa-twitter"></i>
                <span>{t('about.followTwitter', { fallback: 'Follow on Twitter' })}</span>
              </a>
            </div>
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
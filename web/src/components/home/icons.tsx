import React from 'react';
import { IconProps } from '../../types';

// Placeholder for a more sophisticated logo
export const Logo: React.FC<IconProps> = ({ className }) => (
  <span className={`font-bold text-3xl text-blue-600 ${className}`}>
    Decorebator
  </span>
);

export const AppStoreIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
    <path d="M17.482 10.772a4.372 4.372 0 00-1.668-.346c-.925 0-1.942.463-2.818 1.132a4.433 4.433 0 00-1.578 2.807c.047 2.023 1.763 3.194 2.723 3.194.88 0 1.66-.51 2.536-1.226a.703.703 0 00.21-.51c0-1.03-.99-1.532-2.227-1.532-.552 0-.944.116-1.336.346.924-1.485 1.53-2.023 2.56-2.865zm-5.07-6.31S10.64 2 9.127 2C6.37 2 4 4.398 4 7.19c0 1.273.688 3.03 1.83 3.98.881.716 1.486.832 2.4.832.552 0 1.11-.093 1.668-.326a4.775 4.775 0 002.09-1.897c.419-.716.882-2.188.606-3.31-.276-1.124-1.29-1.92-2.157-1.92-.65 0-1.064.348-1.383.695.186-.602.233-1.227.233-1.712a2.792 2.792 0 00-.093-.81zM12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0z"/>
  </svg>
);

export const GooglePlayIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="24" height="24">
    <path d="M4.26 2.92l10.488 6.084-5.05 5.007L4.26 2.92zM16.143 9.6L3.075 21.08a.998.998 0 001.525.866L17.67 14.73l-3.846-3.816.01-.01.31-1.303zM18.168 12.69l-3.19-3.162L20.64 7.33c.91-.53 1.14-1.75.5-2.64-.63-.89-1.83-1.11-2.74-.58L3.66 10.71c-.16.1-.2.34-.1.5l2.37 2.348 12.24-.004z"/>
  </svg>
);

export const WordMeaningIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 6.042A8.967 8.967 0 006 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 016 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 016-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0018 18a8.967 8.967 0 00-6 2.292m0-14.25v14.25" />
  </svg>
);

export const ImageWordIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
  </svg>
);

export const AudioWordIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M19.114 5.636a9 9 0 010 12.728M16.463 8.288a5.25 5.25 0 010 7.424M6.75 8.25l4.72-4.72a.75.75 0 011.28.53v15.88a.75.75 0 01-1.28.53l-4.72-4.72H4.51c-.88 0-1.704-.507-1.938-1.354A9.01 9.01 0 012.25 12c0-.83.112-1.633.322-2.396C2.806 8.756 3.63 8.25 4.51 8.25H6.75z" />
  </svg>
);

export const CheckCircleIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
);

export const ArrowRightIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
  </svg>
);

export const MailIcon: React.FC<IconProps> = ({ className }) => (
 <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
  </svg>
);

export const FacebookIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} fill="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <path d="M22.675 0H1.325C.593 0 0 .593 0 1.325v21.351C0 23.407.593 24 1.325 24H12.82v-9.294H9.692v-3.622h3.128V8.413c0-3.1 1.893-4.788 4.659-4.788 1.325 0 2.463.099 2.795.143v3.24l-1.918.001c-1.504 0-1.795.715-1.795 1.763v2.313h3.587l-.467 3.622h-3.12V24h6.116c.732 0 1.325-.593 1.325-1.325V1.325C24 .593 23.407 0 22.675 0z" />
  </svg>
);

export const TwitterIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} fill="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <path d="M23.953 4.57a10 10 0 01-2.825.775 4.958 4.958 0 002.163-2.723c-.951.555-2.005.959-3.127 1.184a4.92 4.92 0 00-8.384 4.482C7.69 8.095 4.067 6.13 1.64 3.162a4.822 4.822 0 00-.666 2.475c0 1.71.87 3.213 2.188 4.096a4.904 4.904 0 01-2.228-.616v.06a4.923 4.923 0 003.946 4.827 4.996 4.996 0 01-2.212.085 4.936 4.936 0 004.604 3.417 9.867 9.867 0 01-6.102 2.105c-.39 0-.779-.023-1.17-.067a13.995 13.995 0 007.557 2.209c9.053 0 13.998-7.496 13.998-13.985 0-.213 0-.425-.015-.637A9.955 9.955 0 0024 4.59z" />
  </svg>
);

export const InstagramIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} fill="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919C8.416 2.175 8.796 2.163 12 2.163zm0 1.802C9.042 3.965 8.75 3.977 7.478 4.031c-2.7.123-3.996 1.417-4.12 4.12C3.305 9.275 3.293 9.523 3.293 12c0 2.477.012 2.724.067 3.999.124 2.703 1.418 3.996 4.12 4.12 1.27.054 1.517.066 4.462.066s3.192-.012 4.462-.066c2.703-.124 3.996-1.418 4.12-4.12.054-1.27.066-1.517.066-4.462s-.012-3.192-.066-4.462c-.124-2.703-1.418-3.996-4.12-4.12C15.276 3.977 15.028 3.965 12 3.965zm0 2.989c-2.714 0-4.913 2.199-4.913 4.913s2.199 4.913 4.913 4.913 4.913-2.199 4.913-4.913S14.714 6.954 12 6.954zm0 7.831c-1.6 0-2.913-1.312-2.913-2.918S10.4 9.046 12 9.046s2.913 1.312 2.913 2.918-1.313 2.918-2.913 2.918zm4.465-7.772a1.25 1.25 0 100-2.499 1.25 1.25 0 000 2.499z" />
  </svg>
);

export const LinkedInIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} fill="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <path d="M19 0H5a5 5 0 00-5 5v14a5 5 0 005 5h14a5 5 0 005-5V5a5 5 0 00-5-5zM8 19H5V8h3v11zM6.5 6.732a1.732 1.732 0 110-3.464 1.732 1.732 0 010 3.464zM20 19h-3v-5.5c0-1.42-.49-2.38-1.77-2.38-.97 0-1.53.65-1.78 1.28-.09.23-.11.54-.11.85V19H10V8h3v1.337c.41-.78 1.39-1.877 3.05-1.877 2.23 0 3.95 1.45 3.95 4.56V19z" />
  </svg>
);

export const ListPlusIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 6.75h7.5M8.25 12h7.5m-7.5 5.25h7.5M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 17.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM12 3.75v16.5M3.75 12h16.5m-16.5 5.25H12m6.75-5.25V3.75M12 3.75H3.75m16.5 0H12m6.75 5.25h-5.25M12 20.25h8.25A2.25 2.25 0 0022.5 18V6a2.25 2.25 0 00-2.25-2.25H3.75A2.25 2.25 0 001.5 6v12a2.25 2.25 0 002.25 2.25H12zM16.5 9.75v3m0 0v3m0-3h3m-3 0h-3" />
     <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 12h16.5m-16.5 5.25H12m6.75-5.25V3.75M12 3.75H3.75m16.5 0H12m6.75 5.25h-5.25M12 20.25h8.25A2.25 2.25 0 0022.5 18V6a2.25 2.25 0 00-2.25-2.25H3.75A2.25 2.25 0 001.5 6v12a2.25 2.25 0 002.25 2.25H12zM16.5 9.75v3m0 0v3m0-3h3m-3 0h-3" />
    <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 12h-15" /> {/* Simplified version */}
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" /> {/* Another simplified version for plus */}
    <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25H12" /> {/* Lines for list */}
    <path strokeLinecap="round" strokeLinejoin="round" d="M15 18.75v-3.75m0 0V11.25m0 3.75h3.75m-3.75 0H11.25" /> {/* Plus sign */}
     <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 6.75h7.5M8.25 12h7.5m-7.5 5.25h7.5M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 17.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM15 12h3.75m-3.75 0V8.25m0 3.75v3.75" />
  </svg>
);


export const SparklesIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L1.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.25 7.5l.813 2.846a4.5 4.5 0 012.187 2.187L24.096 12l-2.846.813a4.5 4.5 0 01-2.187 2.187L18.25 19.5l-.813-2.846a4.5 4.5 0 01-2.187-2.187L12.404 12l2.846-.813a4.5 4.5 0 012.187-2.187L18.25 7.5z" />
  </svg>
);


export const ChatBubbleBottomCenterTextIcon: React.FC<IconProps> = ({ className }) => (
  <svg className={className} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 9.75a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375m-13.5 3.01c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.184-4.183a.375.375 0 01.265-.112c.878-.082 1.796-.217 2.664-.397a23.915 23.915 0 004.563-1.024c1.406-.534 2.252-1.902 2.252-3.428 0-1.933-2.318-3.514-5.379-3.514H4.875c-3.06 0-5.379 1.58-5.379 3.514zM4.125 12h15.75" />
  </svg>
);

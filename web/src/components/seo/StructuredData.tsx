import React from 'react';

interface StructuredDataProps {
  type?: 'website' | 'softwareApplication' | 'mobileApplication';
}

const StructuredData: React.FC<StructuredDataProps> = ({ type = 'website' }) => {
  const generateStructuredData = () => {
    const baseData = {
      "@context": "https://schema.org",
      "@type": type === 'website' ? 'WebSite' : 'SoftwareApplication',
      "name": "Decorebator",
      "description": "AI-powered vocabulary learning platform with spaced repetition, interactive quizzes, and comprehensive analytics",
      "url": "https://decorebator.com",
      "creator": {
        "@type": "Organization",
        "name": "Decorebator Team"
      },
      "publisher": {
        "@type": "Organization", 
        "name": "Decorebator"
      }
    };

    if (type === 'softwareApplication' || type === 'mobileApplication') {
      return {
        ...baseData,
        "@type": "SoftwareApplication",
        "applicationCategory": "EducationalApplication",
        "applicationSubCategory": "Language Learning",
        "operatingSystem": "iOS, Android",
        "offers": [
          {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD",
            "name": "Free Plan",
            "description": "1 wordlist, up to 10 words, basic quiz modes"
          },
          {
            "@type": "Offer", 
            "price": "6.99",
            "priceCurrency": "USD",
            "name": "Monthly Premium",
            "description": "Unlimited wordlists and words, all 8 quiz modes, advanced analytics, offline support"
          },
          {
            "@type": "Offer",
            "price": "69.90", 
            "priceCurrency": "USD",
            "name": "Annual Premium",
            "description": "Everything in Monthly Premium with annual savings"
          }
        ],
        "aggregateRating": {
          "@type": "AggregateRating",
          "ratingValue": "4.9",
          "ratingCount": "10000",
          "bestRating": "5",
          "worstRating": "1"
        },
        "featureList": [
          "AI-powered content generation",
          "7-box Leitner spaced repetition system", 
          "8 interactive quiz modes",
          "Multi-language support (7 languages)",
          "Visual learning with AI-generated images",
          "Audio pronunciation with TTS",
          "Comprehensive analytics and progress tracking",
          "Offline support for premium users",
          "Error reporting and content quality control",
          "Interactive flashcards with flip animations"
        ],
        "screenshot": "https://decorebator.com/app-screenshot.jpeg",
        "softwareVersion": "2.0",
        "datePublished": "2024-01-01",
        "dateModified": "2025-01-01",
        "inLanguage": ["en", "es", "fr", "de", "it", "pt", "ja"],
        "installUrl": {
          "iOS": "https://apps.apple.com/app/decorebator",
          "Android": "https://play.google.com/store/apps/details?id=com.decorebator"
        }
      };
    }

    if (type === 'website') {
      return {
        ...baseData,
        "potentialAction": {
          "@type": "SearchAction",
          "target": "https://decorebator.com/search?q={search_term_string}",
          "query-input": "required name=search_term_string"
        },
        "mainEntity": {
          "@type": "SoftwareApplication",
          "name": "Decorebator",
          "applicationCategory": "EducationalApplication"
        }
      };
    }

    return baseData;
  };

  const faqStructuredData = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    "mainEntity": [
      {
        "@type": "Question",
        "name": "How does the AI content generation work?",
        "acceptedAnswer": {
          "@type": "Answer",
          "text": "Our AI automatically generates comprehensive learning materials for each word you add. The system creates contextually relevant definitions, example sentences, culturally-aware images, and high-quality audio pronunciations in your target language, with language-specific grammar rules and voice optimization."
        }
      },
      {
        "@type": "Question", 
        "name": "What is the Advanced Leitner spaced repetition system?",
        "acceptedAnswer": {
          "@type": "Answer",
          "text": "Decorebator uses a scientifically optimized 7-box spaced repetition system with intervals: Box 1 (immediate), Box 2 (1 hour), Box 3 (6 hours), Box 4 (1 day), Box 5 (3 days), Box 6 (7 days), and Box 7 (1 month). Words progress through boxes based on your performance, with intelligent quiz type progression for optimal learning."
        }
      },
      {
        "@type": "Question",
        "name": "What platforms and devices are supported?",
        "acceptedAnswer": {
          "@type": "Answer", 
          "text": "Decorebator is available as a mobile app for iOS and Android devices with full feature support and multi-language capabilities. Your progress synchronizes across all your devices with automatic session management."
        }
      },
      {
        "@type": "Question",
        "name": "What languages are supported for AI content generation?",
        "acceptedAnswer": {
          "@type": "Answer",
          "text": "Decorebator supports 7 languages with native AI processing: English, Spanish, French, German, Italian, Portuguese, and Japanese. Each language receives culturally-aware content generation, proper grammar rules, and optimized voice selection."
        }
      }
    ]
  };

  const organizationStructuredData = {
    "@context": "https://schema.org",
    "@type": "Organization",
    "name": "Decorebator",
    "url": "https://decorebator.com",
    "logo": "https://decorebator.com/icon-512x512.png",
    "description": "AI-powered vocabulary learning platform helping users master languages through spaced repetition and interactive learning",
    "foundingDate": "2024",
    "contactPoint": {
      "@type": "ContactPoint",
      "contactType": "customer service",
      "email": "support@decorebator.com"
    },
    "sameAs": [
      "https://twitter.com/decorebator",
      "https://facebook.com/decorebator", 
      "https://linkedin.com/company/decorebator"
    ]
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(generateStructuredData())
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(faqStructuredData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(organizationStructuredData)
        }}
      />
    </>
  );
};

export default StructuredData;
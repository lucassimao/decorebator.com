import React from 'react';
import { statsConfig } from '@/config/statsConfig';

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
        ...(statsConfig.showStats.starRating && {
          "aggregateRating": {
            "@type": "AggregateRating",
            "ratingValue": statsConfig.values.starRating.split('/')[0],
            "ratingCount": statsConfig.showStats.userCount ? statsConfig.values.userCount.replace('+', '').replace(',', '') : "1000",
            "bestRating": "5",
            "worstRating": "1"
          }
        }),
        "featureList": [
          "AI-powered content generation",
          "7-box Leitner spaced repetition system", 
          "8 interactive quiz modes",
          ...(statsConfig.showStats.languageCount ? [`Multi-language support (${statsConfig.values.languageCount} languages)`] : []),
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
          "text": `Decorebator supports ${statsConfig.showStats.languageCount ? statsConfig.values.languageCount : 'multiple'} languages with native AI processing: English, Spanish, French, German, Italian, Portuguese, and Japanese. Each language receives culturally-aware content generation, proper grammar rules, and optimized voice selection.`
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

  const breadcrumbStructuredData = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    "itemListElement": [
      {
        "@type": "ListItem",
        "position": 1,
        "name": "Home",
        "item": "https://decorebator.com"
      },
      {
        "@type": "ListItem", 
        "position": 2,
        "name": "Language Learning App",
        "item": "https://decorebator.com"
      }
    ]
  };

  const educationalOrganizationData = {
    "@context": "https://schema.org",
    "@type": "EducationalOrganization",
    "name": "Decorebator",
    "url": "https://decorebator.com",
    "logo": "https://decorebator.com/icon-512x512.png",
    "description": "Advanced AI-powered vocabulary learning platform for language education",
    "educationalCredentialAwarded": "Vocabulary mastery certification",
    "hasCredential": {
      "@type": "EducationalOccupationalCredential",
      "name": "Language Learning Progress Tracking",
      "description": "Comprehensive vocabulary learning progress with spaced repetition analytics"
    },
    "teaches": [
      "Vocabulary acquisition",
      "Language comprehension", 
      "Spaced repetition methodology",
      "Interactive language learning",
      "Multi-language proficiency"
    ]
  };

  const courseStructuredData = {
    "@context": "https://schema.org",
    "@type": "Course",
    "name": "AI-Powered Vocabulary Learning",
    "description": "Master vocabulary in any language using AI-generated content, spaced repetition, and interactive quizzes",
    "provider": {
      "@type": "Organization",
      "name": "Decorebator"
    },
    "educationalLevel": "All levels",
    "courseMode": "online",
    "numberOfCredits": 0,
    "timeRequired": "PT30M",
    "inLanguage": ["en", "es", "fr", "de", "it", "pt", "ja"],
    "coursePrerequisites": "None",
    "syllabusSections": [
      {
        "@type": "Syllabus",
        "name": "AI Content Generation",
        "description": "Learn with AI-generated definitions, images, and audio"
      },
      {
        "@type": "Syllabus", 
        "name": "Spaced Repetition System",
        "description": "Master the 7-box Leitner system for optimal retention"
      },
      {
        "@type": "Syllabus",
        "name": "Interactive Quiz Modes",
        "description": "Practice with 8 different quiz types and learning approaches"
      }
    ]
  };

  const productStructuredData = {
    "@context": "https://schema.org",
    "@type": "Product",
    "name": "Decorebator - AI Vocabulary Learning App",
    "description": "Advanced vocabulary learning platform using AI content generation and spaced repetition",
    "brand": {
      "@type": "Brand",
      "name": "Decorebator"
    },
    "category": "Educational Software",
    "offers": [
      {
        "@type": "Offer",
        "name": "Free Plan",
        "price": "0",
        "priceCurrency": "USD",
        "availability": "https://schema.org/InStock",
        "description": "1 wordlist, up to 10 words, basic quiz modes"
      },
      {
        "@type": "Offer",
        "name": "Monthly Premium",
        "price": "6.99",
        "priceCurrency": "USD", 
        "priceSpecification": {
          "@type": "RecurringPaymentFrequency",
          "frequency": "Monthly"
        },
        "availability": "https://schema.org/InStock",
        "description": "Unlimited wordlists and words, all quiz modes, offline support"
      },
      {
        "@type": "Offer",
        "name": "Annual Premium",
        "price": "69.90",
        "priceCurrency": "USD",
        "priceSpecification": {
          "@type": "RecurringPaymentFrequency", 
          "frequency": "Yearly"
        },
        "availability": "https://schema.org/InStock",
        "description": "Best value - includes everything in Monthly plus early access features"
      }
    ],
    "hasFeatureList": [
      "AI-powered content generation",
      "7-box Leitner spaced repetition",
      "8 interactive quiz modes",
      "Multi-language support (7 languages)",
      "Offline learning capability",
      "Advanced progress analytics",
      "Visual learning with AI images",
      "Audio pronunciation training"
    ],
    "operatingSystem": ["iOS", "Android"],
    "applicationCategory": "EducationalApplication",
    "isAccessibleForFree": true
  };

  const howToStructuredData = {
    "@context": "https://schema.org",
    "@type": "HowTo",
    "name": "How to Learn Vocabulary with Decorebator",
    "description": "Step-by-step guide to mastering vocabulary using AI-powered spaced repetition",
    "image": "https://decorebator.com/app-screenshot.jpeg",
    "estimatedCost": {
      "@type": "MonetaryAmount",
      "currency": "USD",
      "value": "0"
    },
    "step": [
      {
        "@type": "HowToStep",
        "name": "Create Your Account",
        "text": "Sign up for free and download the mobile app",
        "image": "https://decorebator.com/step1-signup.jpg"
      },
      {
        "@type": "HowToStep", 
        "name": "Add Words to Learn",
        "text": "Create wordlists and add vocabulary words you want to master",
        "image": "https://decorebator.com/step2-addwords.jpg"
      },
      {
        "@type": "HowToStep",
        "name": "AI Generates Content",
        "text": "Our AI automatically creates definitions, images, and audio for each word",
        "image": "https://decorebator.com/step3-ai-content.jpg"
      },
      {
        "@type": "HowToStep",
        "name": "Practice with Quizzes", 
        "text": "Use 8 different quiz modes to test your knowledge and memory",
        "image": "https://decorebator.com/step4-quiz.jpg"
      },
      {
        "@type": "HowToStep",
        "name": "Track Your Progress",
        "text": "Monitor your learning with detailed analytics and mastery levels",
        "image": "https://decorebator.com/step5-analytics.jpg"
      }
    ],
    "totalTime": "PT30M"
  };

  const videoObjectData = {
    "@context": "https://schema.org",
    "@type": "VideoObject",
    "name": "Decorebator App Demo - AI Vocabulary Learning",
    "description": "See how Decorebator uses AI to generate comprehensive learning materials and spaced repetition for vocabulary mastery",
    "thumbnailUrl": "https://decorebator.com/app-screenshot.jpeg",
    "uploadDate": "2024-12-01",
    "duration": "PT2M30S",
    "contentUrl": "https://decorebator.com/app-demo.mp4",
    "embedUrl": "https://decorebator.com/embed/demo",
    "interactionStatistic": {
      "@type": "InteractionCounter",
      "interactionType": "https://schema.org/WatchAction",
      "userInteractionCount": 15000
    },
    "publisher": {
      "@type": "Organization",
      "name": "Decorebator"
    }
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
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(breadcrumbStructuredData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(educationalOrganizationData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(courseStructuredData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(productStructuredData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(howToStructuredData)
        }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(videoObjectData)
        }}
      />
    </>
  );
};

export default StructuredData;
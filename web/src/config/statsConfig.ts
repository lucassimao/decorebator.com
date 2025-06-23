/**
 * Statistics Configuration
 * 
 * This configuration controls the display of various statistics and social proof
 * elements throughout the application. Used for marketing stats, user counts, etc.
 */

export const statsConfig = {
  // Global visibility toggles
  showStats: {
    userCount: false,      // "10,000+ active learners"
    starRating: false,     // "4.9/5" rating display
    languageCount: false,  // "7 native AI languages"
  },
  
  // Configurable values (when stats are enabled)
  values: {
    userCount: "10,000+",
    starRating: "4.9/5", 
    languageCount: "7",
    totalLanguages: "7 languages with native AI",
  },
  
  // Locations where stats appear
  locations: {
    heroSection: {
      showSocialProof: false, // Hide entire social proof section
    },
    ctaSection: {
      showUserCount: false,   // Hide user count in CTA text
    },
    testimonials: {
      showStats: false,       // Hide stats in testimonials section
    }
  }
};
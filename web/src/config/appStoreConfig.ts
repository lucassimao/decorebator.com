/**
 * App Store Configuration
 * 
 * This configuration controls the behavior of download buttons throughout the app.
 * When the apps are approved and live in the stores, update the URLs here.
 */

export const appStoreConfig = {
  // Toggle this to false to disable the "coming soon" modal globally
  showPendingModal: true,
  
  // Set these to actual store URLs when apps are approved
  // null values will trigger the pending modal (if showPendingModal is true)
  appStoreUrl: 'https://apps.apple.com/app/decorebator/id6749329064',
  playStoreUrl: 'https://play.google.com/store/apps/details?id=com.lsimaocosta.decorebator',
};
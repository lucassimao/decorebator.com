import offlineManager from "./offlineManager";

// Test utility to simulate offline mode
export const simulateOfflineMode = {
  // Go offline
  goOffline: () => {
    // This is a test utility - in real usage, network status is detected automatically
    console.log("🔴 Simulating offline mode");
    // Note: This won't actually change the network status in offlineManager
    // as it's controlled by NetInfo. This is just for testing UI states.
  },

  // Go online
  goOnline: () => {
    console.log("🟢 Simulating online mode");
  },

  // Set premium status for testing
  setPremium: (isPremium: boolean) => {
    console.log(`💎 Setting premium status to: ${isPremium}`);
    offlineManager.setUserPremiumStatus(isPremium);
  },

  // Check current status
  checkStatus: () => {
    const isOnline = offlineManager.getNetworkStatus();
    const isOfflineAvailable = offlineManager.isOfflineAvailable();
    console.log("📊 Current status:");
    console.log(`   Network: ${isOnline ? "Online" : "Offline"}`);
    console.log(`   Offline Available: ${isOfflineAvailable}`);
  },

  // Clear cache
  clearCache: async () => {
    console.log("🗑️  Clearing offline cache...");
    await offlineManager.clearCache();
    console.log("✅ Cache cleared");
  },

  // Clean up expired quizzes
  cleanupExpired: async () => {
    console.log("🧹 Cleaning up expired quizzes...");
    await offlineManager.cleanupExpiredQuizzes();
    console.log("✅ Cleanup complete");
  },

  // Get cached quizzes for a wordlist
  getCachedQuizzes: async (wordlistId: number) => {
    console.log(`📋 Getting cached quizzes for wordlist ${wordlistId}...`);
    const quizzes = await offlineManager.getAllCachedQuizzes(wordlistId);
    console.log(`Found ${quizzes.length} cached quiz(zes)`);
    return quizzes;
  },
};

// Export for development console
if (__DEV__) {
  (global as any).offlineTest = simulateOfflineMode;
}

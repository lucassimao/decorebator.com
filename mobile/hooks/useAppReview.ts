import { useState, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import * as StoreReview from "expo-store-review";
import Constants from "expo-constants";

const REVIEW_COMPLETED_KEY = "appReviewCompleted";
const REVIEW_DISMISSED_PREFIX = "appReviewDismissedForVersion:";

// Thresholds for showing review prompt
const MIN_QUIZ_COUNT = 10;
const MIN_ACCURACY = 0.7;

type ReviewEligibility = {
  isEligible: boolean;
  reason?: string;
};

/**
 * Hook for managing app store review prompts.
 *
 * Logic:
 * - Shows after a positive quiz session (10+ questions, 70%+ accuracy)
 * - Never shows again if user tapped "Yes, I love it!"
 * - Shows once per app version if user tapped "Not really"
 * - "Maybe later" allows prompting again in the same version
 */
export const useAppReview = () => {
  const [showReviewModal, setShowReviewModal] = useState(false);

  const appVersion = Constants.expoConfig?.version ?? "unknown";

  /**
   * Check if user is eligible for a review prompt
   */
  const checkEligibility = useCallback(async (): Promise<ReviewEligibility> => {
    try {
      const [completed, dismissed] = await AsyncStorage.multiGet([
        REVIEW_COMPLETED_KEY,
        `${REVIEW_DISMISSED_PREFIX}${appVersion}`,
      ]);

      const hasCompleted = completed[1] === "true";
      const hasDismissed = dismissed[1] === "true";

      if (hasCompleted) {
        return { isEligible: false, reason: "already_completed" };
      }

      if (hasDismissed) {
        return { isEligible: false, reason: "dismissed_this_version" };
      }

      return { isEligible: true };
    } catch (error) {
      if (__DEV__) {
        console.warn("Error checking review eligibility:", error);
      }
      return { isEligible: false, reason: "storage_error" };
    }
  }, [appVersion]);

  /**
   * Check quiz performance and prompt for review if eligible
   */
  const checkAndPromptReview = useCallback(
    async (quizCount: number, correctCount: number): Promise<boolean> => {
      // Check performance thresholds
      if (quizCount < MIN_QUIZ_COUNT) {
        if (__DEV__) {
          console.log(
            `Review prompt skipped: quiz count ${quizCount} < ${MIN_QUIZ_COUNT}`,
          );
        }
        return false;
      }

      const accuracy = correctCount / quizCount;
      if (accuracy < MIN_ACCURACY) {
        if (__DEV__) {
          console.log(
            `Review prompt skipped: accuracy ${(accuracy * 100).toFixed(1)}% < ${MIN_ACCURACY * 100}%`,
          );
        }
        return false;
      }

      // Check eligibility
      const { isEligible, reason } = await checkEligibility();
      if (!isEligible) {
        if (__DEV__) {
          console.log(`Review prompt skipped: ${reason}`);
        }
        return false;
      }

      if (__DEV__) {
        console.log(
          `Review prompt showing: ${quizCount} quizzes, ${(accuracy * 100).toFixed(1)}% accuracy`,
        );
      }

      setShowReviewModal(true);
      return true;
    },
    [checkEligibility],
  );

  /**
   * User tapped "Yes, I love it!" - trigger native review and mark as completed
   */
  const handleLoveIt = useCallback(async () => {
    setShowReviewModal(false);

    // Mark as completed first (optimistic - we can't know if they actually reviewed)
    try {
      await AsyncStorage.setItem(REVIEW_COMPLETED_KEY, "true");
    } catch (error) {
      if (__DEV__) {
        console.warn("Failed to save review completed state:", error);
      }
    }

    // Trigger native review prompt
    try {
      const isAvailable = await StoreReview.isAvailableAsync();
      if (isAvailable) {
        await StoreReview.requestReview();
      } else if (__DEV__) {
        console.log(
          "Store review not available (dev build, simulator, or OS limit reached)",
        );
      }
    } catch (error) {
      // Silently fail - user already saw our prompt, native one is a bonus
      if (__DEV__) {
        console.warn("Store review request failed:", error);
      }
    }
  }, []);

  /**
   * User tapped "Not really" - dismiss for this version
   */
  const handleNotReally = useCallback(async () => {
    setShowReviewModal(false);

    try {
      await AsyncStorage.setItem(
        `${REVIEW_DISMISSED_PREFIX}${appVersion}`,
        "true",
      );
    } catch (error) {
      if (__DEV__) {
        console.warn("Failed to save review dismissed state:", error);
      }
    }
  }, [appVersion]);

  /**
   * User tapped "Maybe later" - close without setting any flags
   */
  const handleMaybeLater = useCallback(() => {
    setShowReviewModal(false);
  }, []);

  /**
   * Close modal (same as maybe later)
   */
  const closeReviewModal = useCallback(() => {
    setShowReviewModal(false);
  }, []);

  return {
    showReviewModal,
    checkAndPromptReview,
    handleLoveIt,
    handleNotReally,
    handleMaybeLater,
    closeReviewModal,
  };
};

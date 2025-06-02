import { UserProfile, getProfile } from "@/api/users";
import { useEffect, useState } from "react";
import offlineManager from "@/utils/offlineManager";

export const useUserInfo = () => {
  const [userInfo, setUserInfo] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<any>(null);

  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const user = await getProfile();
        setUserInfo(user);
        
        // Update offline manager with premium status
        const isPremium = user?.subscriptionPlan === 'monthly' || user?.subscriptionPlan === 'annual';
        offlineManager.setUserPremiumStatus(isPremium);
      } catch (err) {
        setError(err);
      } finally {
        setLoading(false);
      }
    };

    fetchUserInfo();
  }, []);

  return { userInfo, loading, error };
};

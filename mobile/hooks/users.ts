import { getProfile } from "@/api/users";
import offlineManager from "@/utils/offlineManager";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

export const useUserInfo = () => {

  const { data: user, isLoading,error } = useQuery({
    queryKey: ["userProfile"],
    queryFn: getProfile,
  });  

    // derive isPremium whenever `user` changes
  const isPremium = !!user && (user.subscriptionPlan === "monthly" || user.subscriptionPlan === "annual");

  useEffect(() => {
    if (user){
      offlineManager.setUserPremiumStatus(isPremium);
    }

  }, [user,isPremium]);

  return { userInfo: user, loading: isLoading, error, isPremium };
};

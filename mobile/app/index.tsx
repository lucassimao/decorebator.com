import { useUserInfo } from "@/hooks/users";
import { useSnackbar } from "@/hooks/useSnackbar";
import { router } from "expo-router";
import { useEffect } from "react";

export default function Index() {
  const { userInfo, error, loading } = useUserInfo();
  const snackbar = useSnackbar();

  useEffect(() => {
    if (loading) return;

    if (userInfo) {
      router.replace("/dashboard");
    } else {
      router.replace("/signin");
    }
  }, [userInfo, loading]);

  useEffect(() => {
    if (error) {
      snackbar.show(error.message, "error", 500);
    }
  }, [error, snackbar]);

  return null;
}

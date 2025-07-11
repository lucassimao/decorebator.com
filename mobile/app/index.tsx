import * as usersApi from "@/api/users";
import { useUserInfo } from "@/hooks/users";
import { Redirect, useRouter } from "expo-router";
import React, { useEffect } from "react";

export default function Index() {
  const { error } = useUserInfo();
  const router = useRouter();

  useEffect(() => {
    if (error) {
      console.error(error);
      usersApi.sigout();
      router.replace("/signin");
    }
  }, [error, router]);

  console.log(error);

  return <Redirect href="/dashboard" />;
}

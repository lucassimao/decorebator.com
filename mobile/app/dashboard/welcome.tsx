import EmptyDashboard from "@/components/dashboard/EmptyDashboard";
import { router } from "expo-router";
import * as React from "react";
import * as usersApi from "@/api/users";

export default function Welcome() {
  const user = usersApi.getUserInfo();
  if (!user) return null;

  return (
    <EmptyDashboard
      title={`Welcome, ${user.firstName}`}
      subtitle="All set for your learning journey !"
      onWordlistCreated={() => router.replace("dashboard")}
    />
  );
}

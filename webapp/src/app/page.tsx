"use client";
import { isAuthenticated } from "@/lib/auth";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function Home() {
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated()) {
      router.push("/wordlists");
    } else {
      router.push("/login");
    }
  }, []);

  return null;
}

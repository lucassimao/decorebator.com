"use client";

import { AppShell, Burger } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";

import { Group, Skeleton } from "@mantine/core";
import { useRouter } from "next/navigation";
import { isAuthenticated } from "@/lib/auth";
import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { getWordlists } from "@/lib/wordlists";
import { Badge, NavLink } from "@mantine/core";

export default function LayoutDashboard({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const [opened, { toggle }] = useDisclosure();
  const router = useRouter();

  useEffect(() => {
    if (!isAuthenticated()) {
      router.push("/login");
    }
  }, []);

  const {
    data: wordlists,
    isLoading: isLoadingWordlists,
    isError,
    error,
  } = useQuery({
    queryKey: ["wordlists"],
    queryFn: getWordlists,
  });

  // TODO handle errors

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{ width: 300, breakpoint: "sm", collapsed: { mobile: !opened } }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md">
          <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" />
          <img src="" />
        </Group>
      </AppShell.Header>
      <AppShell.Navbar p="md">
        Wordlists
        {isLoadingWordlists
          ? Array(15)
              .fill(0)
              .map((_, index) => (
                <Skeleton key={index} h={28} mt="sm" animate={true} />
              ))
          : wordlists?.map((wordlist) => (
              <NavLink
                href={`/wordlists/${wordlist.id}`}
                label={wordlist.name}
                key={`wordlist${wordlist.id}`}
              />
            ))}
      </AppShell.Navbar>
      <AppShell.Main>{children}</AppShell.Main>
    </AppShell>
  );
}

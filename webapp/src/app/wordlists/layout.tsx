"use client";

import { AppShell, Burger } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";

import { isAuthenticated } from "@/lib/auth";
import { getWordlists } from "@/lib/wordlists";
import { Group, NavLink, Skeleton } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function LayoutDashboard({
  children,
  params,
}: Readonly<{
  children: React.ReactNode;
  params: { wordlistId: string };
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
    isFetched,
  } = useQuery({
    queryKey: ["wordlists"],
    queryFn: getWordlists,
  });

  useEffect(() => {
    if (!isFetched) return;
    if (!wordlists?.length) return;
    if (!params.wordlistId) return;

    router.push(`/wordlists/${wordlists[0].id}`);
  }, [isFetched, params.wordlistId]);

  // TODO handle errors

  // display default view if no wordlists

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

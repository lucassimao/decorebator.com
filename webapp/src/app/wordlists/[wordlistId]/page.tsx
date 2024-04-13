"use client";

import { getWords } from "@/lib/wordlists";
import { useQuery } from "@tanstack/react-query";
import { ScrollArea, UnstyledButton, TextInput } from "@mantine/core";
import { useRef, useState } from "react";

export default function WordPage({
  params,
}: {
  params: { wordlistId: string };
}) {
  // TODO handle errors
  const {
    data: words,
    isLoading: isLoadingWordlists,
    isError,
    isFetched,
    error,
  } = useQuery({
    queryKey: ["words", params.wordlistId],
    queryFn: () => getWords(parseInt(params.wordlistId)),
  });

  const viewportRef = useRef<HTMLDivElement>(null);
  const [query, setQuery] = useState("");
  const [hovered, setHovered] = useState(-1);

  const filtered = words?.filter((word) =>
    word.name.toLowerCase().includes(query.toLowerCase()),
  );

  const items = filtered?.map((word, index) => (
    <UnstyledButton
      data-list-item
      key={`word-${word.id}`}
      display="block"
      // bg={index === hovered ? 'var(--mantine-color-blue-light)' : undefined}
      w="100%"
      p={5}
    >
      {word.name}
    </UnstyledButton>
  ));

  return (
    <>
      <TextInput
        value={query}
        onChange={(event) => {
          setQuery(event.currentTarget.value);
          setHovered(-1);
        }}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown") {
            event.preventDefault();
            setHovered((current) => {
              const nextIndex =
                current + 1 >= (words?.length || 0) ? current : current + 1;
              viewportRef.current
                ?.querySelectorAll("[data-list-item]")
                ?.[nextIndex]?.scrollIntoView({ block: "nearest" });
              return nextIndex;
            });
          }

          if (event.key === "ArrowUp") {
            event.preventDefault();
            setHovered((current) => {
              const nextIndex = current - 1 < 0 ? current : current - 1;
              viewportRef.current
                ?.querySelectorAll("[data-list-item]")
                ?.[nextIndex]?.scrollIntoView({ block: "nearest" });
              return nextIndex;
            });
          }
        }}
        placeholder="Search words"
      />
      <ScrollArea type="always" mt="md" viewportRef={viewportRef}>
        {items}
      </ScrollArea>
    </>
  );
}

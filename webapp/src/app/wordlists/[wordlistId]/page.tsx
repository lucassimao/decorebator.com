"use client";

import { getWords } from "@/lib/wordlists";
import { ScrollArea, TextInput } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { useRef, useState } from "react";
import styles from "./styles.module.css";

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

  const items = (
    <ol
      style={{ "--length": filtered?.length || 0 } as React.CSSProperties}
      role="list"
    >
      {filtered?.map((word, index) => (
        <li style={{ "--i": index + 1 } as React.CSSProperties}>
          <h3>{word.name}</h3>
          <p>
            Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
            eiusmod tempor incididunt ut labore et dolore magna aliqua.
            Adipiscing diam donec adipiscing tristique risus.
          </p>
        </li>
      ))}
    </ol>
  );

  return (
    <div className={styles.wrapper}>
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
    </div>
  );
}

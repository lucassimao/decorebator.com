import * as wordlistsApi from "@/api/wordlists";
import { useQuery } from "@tanstack/react-query";
import React, { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import {
  ActivityIndicator,
  List,
  TouchableRipple,
  useTheme,
} from "react-native-paper";

import SnackBar, { SnackBarProps } from "@/components/SnackBar";
import BottonBar from "@/components/dashboard/BottonBar";
import DeleteWordDialog from "@/components/dashboard/DeleteWordDialog";
import EmptyDashboard from "@/components/dashboard/EmptyDashboard";
import LoadingIndicator from "@/components/dashboard/LoadingIndicator";

export default function Dashboard() {
  const theme = useTheme();
  const [expandedWordlistId, setExpandedWordlistId] = React.useState<
    number | null
  >(null);
  const [wordToDelete, setWordToDelete] =
    React.useState<wordlistsApi.Word | null>(null);
  const [snackBarProps, setSnackBarProps] =
    React.useState<SnackBarProps | null>(null);

  // useFocusEffect(
  //     React.useCallback(() => {
  //         // refetch wordlists when focusing
  //         refetchWordlists()
  //     }, [])
  //   );

  const closeSnackBar = () => setSnackBarProps(null);

  const {
    data: wordlists,
    error: wordlistsError,
    isLoading: isLoadingWordlists,
    refetch: getWordlists,
  } = useQuery<wordlistsApi.Wordlist[], Error>({
    queryFn: () => wordlistsApi.getUserWordlists(),
    queryKey: ["wordlists"],
  });

  const {
    data: words,
    refetch: getWords,
    isLoading: isLoadingWords,
    error: wordsError,
  } = useQuery<wordlistsApi.Word[], Error>({
    queryFn: () =>
      expandedWordlistId ? wordlistsApi.getWords(expandedWordlistId) : [],
    queryKey: ["words", `wordlist-${expandedWordlistId}`],
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: !!expandedWordlistId, // Only run the query if wordlistId is not null
  });

  React.useEffect(() => {
    if (wordlistsError) {
      setSnackBarProps({
        message: "Could not fetch your wordlists at this moment.",
        type: "error",
        onDismiss: closeSnackBar,
      });
      return;
    }

    if (wordsError) {
      setSnackBarProps({
        message: "Could not fetch words at this moment.",
        type: "error",
        onDismiss: closeSnackBar,
      });
      return;
    }
  }, [wordlistsError, wordsError]);

  if (isLoadingWordlists) return <LoadingIndicator />;

  if (!wordlists?.length) return <EmptyDashboard />;

  const onListAccordionPressed = (wordlistId: number) => {
    // user pressed same item. Closing it
    if (expandedWordlistId == wordlistId) {
      setExpandedWordlistId(null);
      return;
    }
    setExpandedWordlistId(wordlistId);
    getWords();
  };

  const currentWordlist = wordlists.find((w) => w.id == expandedWordlistId);

  const onWordAdded = () => {
    getWords();
    setSnackBarProps({
      message: "word added",
      type: "success",
      onDismiss: closeSnackBar,
    });
  };

  const onWordlistDeleted = () => {
    getWordlists();
    setSnackBarProps({
      message: "wordlist deleted",
      type: "success",
      onDismiss: closeSnackBar,
    });
  };

  const onDismissDeleteWordDialog = (success?: boolean) => {
    if (success) {
      setSnackBarProps({
        message: "word deleted",
        type: "success",
        onDismiss: closeSnackBar,
      });
      getWords();
    }
    setWordToDelete(null);
  };

  return (
    <View style={styles.container}>
      {wordToDelete && (
        <DeleteWordDialog
          word={wordToDelete}
          onDismiss={onDismissDeleteWordDialog}
        />
      )}
      {snackBarProps && <SnackBar {...snackBarProps} />}

      <ScrollView contentContainerStyle={styles.scrollViewContent}>
        <List.Section>
          {wordlists.map((wordlist) => {
            return (
              <List.Accordion
                expanded={wordlist.id == expandedWordlistId}
                onPress={() => onListAccordionPressed(wordlist.id)}
                key={`wordlists-${wordlist.id}`}
                title={wordlist.name}
                description={wordlist.description}
                left={(props) => <List.Icon {...props} icon="folder" />}
              >
                {isLoadingWords && (
                  <View style={styles.wordsActivityIndicator}>
                    <ActivityIndicator animating={true} theme={theme} />
                  </View>
                )}
                {words?.map((w) => (
                  <TouchableRipple
                    centered
                    key={`wordlist-${wordlist.id}-${w.id}`}
                    onLongPress={() => setWordToDelete(w)}
                  >
                    <List.Item title={w.name} />
                  </TouchableRipple>
                ))}
              </List.Accordion>
            );
          })}
        </List.Section>
      </ScrollView>
      {currentWordlist && (
        <BottonBar
          onWordAdded={onWordAdded}
          onWordlistDeleted={onWordlistDeleted}
          wordlist={currentWordlist}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
  container: {
    flex: 1,
  },
  scrollViewContent: {
    paddingBottom: 56, // Ensure there's enough space for the BottomNavigation
    margin: 10,
  },
  wordsActivityIndicator: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    height: 50,
  },
});

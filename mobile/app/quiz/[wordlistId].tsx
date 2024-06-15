import * as wordlistsApi from "@/api/wordlists";
import Snackbar, { SnackBarProps } from "@/components/SnackBar";
import Quiz from "@/components/quiz/Quiz";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams, useNavigation } from "expo-router";
import * as React from "react";
import { StyleSheet, View } from "react-native";
import { ActivityIndicator, useTheme } from "react-native-paper";

export default function QuizScreen() {
  const { wordlistId } = useLocalSearchParams();
  const theme = useTheme();

  const [snackBarProps, setSnackBarProps] =
    React.useState<SnackBarProps | null>(null);
  const closeSnackBar = () => setSnackBarProps(null);

  const navigation = useNavigation();

  React.useEffect(() => {
    navigation.setOptions({
      headerShown: true,
      headerTitle: "Quiz",
    });
  }, [navigation]);

  const {
    data: quiz,
    isError,
    error,
    isLoading,
    refetch,
  } = useQuery<wordlistsApi.Quiz, Error>({
    queryFn: () => wordlistsApi.newQuiz(Number(wordlistId)),
    staleTime: 0,
    queryKey: [],
    networkMode: `always`,
    refetchOnMount: false,
  });

  const { mutate: answerQuiz, isPending } = useMutation<void, Error, boolean>({
    mutationFn: (success) =>
      wordlistsApi.answerQuiz(Number(wordlistId), Number(quiz?.id), success),
    onError: (error) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    },
    onSuccess: () => {
      refetch();
    },
  });

  React.useEffect(() => {
    if (isError) {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    }
  }, [isError, error]);

  const onOptionSelected = (optionIndex: number) =>
    answerQuiz(optionIndex == quiz?.answerIndex);

  const isAnyRequestPending = isPending || isLoading;

  if (isAnyRequestPending) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size={"large"} animating={true} theme={theme} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {quiz && <Quiz onOptionSelected={onOptionSelected} quiz={quiz} />}

      {snackBarProps && <Snackbar {...snackBarProps} />}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
});

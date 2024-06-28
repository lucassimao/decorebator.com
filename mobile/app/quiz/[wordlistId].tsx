import * as wordlistsApi from "@/api/wordlists";
import Snackbar, { SnackBarProps } from "@/components/SnackBar";
import Quiz from "@/components/quiz/Quiz";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams, useNavigation } from "expo-router";
import * as React from "react";
import { StyleSheet, Text, View } from "react-native";
import {
  ActivityIndicator,
  IconButton,
  Surface,
  Tooltip,
  TouchableRipple,
  useTheme,
} from "react-native-paper";



export default function QuizScreen() {
  const { wordlistId } = useLocalSearchParams();
  const theme = useTheme();

  const [isNextButtonVisible, setButtonNextVisible] =
    React.useState<boolean>(false);

  const [snackBarProps, setSnackBarProps] =
    React.useState<SnackBarProps | null>(null);
  const closeSnackBar = () => setSnackBarProps(null);

  const [isFastMode, setFastMode] = React.useState<boolean>();

  React.useEffect(() => {
    
    if (typeof isFastMode != 'boolean') return 

    setSnackBarProps({
      message: `Fast mode ${isFastMode ?  'enabled':'disabled'}`,
      type: 'info',
      onDismiss: closeSnackBar
    })
  }, [isFastMode])

  const navigation = useNavigation();

  React.useEffect(() => {

    const options: React.JSX.Element[] = [];

    options.push(
      <Tooltip title="Fast mode">
        <IconButton
          icon={'clock-fast'}
          iconColor={isFastMode  ?  theme.colors.primary :'#ccc' }
          size={25}
          key={"fastMode"}
          onPress={() => setFastMode(isFastMode => !isFastMode)}
        /></Tooltip>,
    );

    navigation.setOptions({
      headerShown: true,
      headerTitle: "Quiz",
      headerRight: () => (
        <View style={{ display: "flex", flexDirection: "row" }}>{options}</View>
      ),
    });
  }, [navigation, isFastMode]);

  const {
    data: quiz,
    error: fetchQuizError,
    isLoading: isLoadingQuiz,
    isRefetching: isFetchingNewQuiz,
    refetch: getNewQuiz,
  } = useQuery<wordlistsApi.Quiz, Error>({
    queryFn: () => wordlistsApi.newQuiz(Number(wordlistId)),
    staleTime: 0,
    queryKey: ["quiz", wordlistId],
    refetchOnMount: false,
  });

  const { mutate: answerQuiz, isPending: isAnsweringQuiz } = useMutation<
    void,
    Error,
    boolean
  >({
    mutationFn: (success) =>
      wordlistsApi.answerQuiz(Number(wordlistId), Number(quiz?.id), success),
    onError: (error) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    },
    onSuccess: (_,isAnswerCorrect) => {
      if (isFastMode && isAnswerCorrect) {
        setTimeout(getNewQuiz, 500);
      } else {
        setButtonNextVisible(true)
      }
    },
  });

  React.useEffect(() => {
    if (!fetchQuizError) return;

    setSnackBarProps({
      onDismiss: closeSnackBar,
      message: fetchQuizError.message,
      type: "error",
    });
  }, [fetchQuizError]);

  const onOptionSelected = (optionIndex: number) =>
    answerQuiz(optionIndex == quiz?.answerIndex);

  const onNextQuizPressed = () => {
    getNewQuiz();
    setButtonNextVisible(false);
  };


  if (isLoadingQuiz || isFetchingNewQuiz) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size={"large"} animating={true} theme={theme} />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {quiz && (
        <Quiz
          isAnsweringQuiz={isAnsweringQuiz}
          onOptionSelected={onOptionSelected}
          quiz={quiz}
        />
      )}

      {isNextButtonVisible && (
        <Surface
          theme={theme}
          style={[{ backgroundColor: theme.colors.primary }, styles.option]}
          elevation={5}
        >
          <TouchableRipple
            style={{ padding: 30 }}
            theme={theme}
            rippleColor={theme.colors.inversePrimary}
            onPress={onNextQuizPressed}
          >
            <Text style={styles.buttonText}>Next Quizz</Text>
          </TouchableRipple>
        </Surface>
      )}

      {snackBarProps && <Snackbar {...snackBarProps} />}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    width: "100%",
    height: "100%",
  },
  buttonText: {
    textAlign: "center",
    fontSize: 20,
    color: "#000",
    width: "100%",
  },
  option: {
    marginTop: 10,
    marginBottom: 50,
    padding: 0,
    justifyContent: "center",
    borderRadius: 10,
    width: "90%",
  },
});

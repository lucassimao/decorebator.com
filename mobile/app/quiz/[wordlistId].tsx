import * as wordlistsApi from "@/api/wordlists";
import * as errorReportingApi from "@/api/errorReporting";
import { ErrorType } from "@/api/errorReporting";

import Snackbar, { SnackBarProps } from "@/components/SnackBar";
import Quiz from "@/components/quiz/Quiz";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams, useNavigation } from "expo-router";
import * as React from "react";
import {
  findNodeHandle,
  LayoutChangeEvent,
  StyleSheet,
  Text,
  UIManager,
  View,
} from "react-native";
import {
  ActivityIndicator,
  Appbar,
  Button,
  Divider,
  IconButton,
  Menu,
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
    if (typeof isFastMode != "boolean") return;

    setSnackBarProps({
      message: `Fast mode ${isFastMode ? "enabled" : "disabled"}`,
      type: "info",
      onDismiss: closeSnackBar,
    });
  }, [isFastMode]);

  const navigation = useNavigation();

  const [isMenuVisible, setMenuVisible] = React.useState(false);
  const menuAnchorRef = React.useRef<View>(null);

  const openMenu = () => {
    if (!menuAnchorRef.current) return;

    menuAnchorRef.current.measure((x, y, width, height, pageX, pageY) => {
      setMenuPosition({ x: pageX + width, y: pageY + height }); // Position menu below header button
      setMenuVisible(true);
    });
  };
  const closeMenu = () => setMenuVisible(false);

  const [menuPosition, setMenuPosition] = React.useState({ x: 0, y: 0 });

  React.useEffect(() => {
    const options: React.JSX.Element[] = [];

    options.push(
      <Tooltip title="Report error">
        <IconButton
          ref={menuAnchorRef}
          icon={"bug"}
          iconColor={theme.colors.primary}
          size={25}
          key={"reportError"}
          onPress={isMenuVisible ? closeMenu : openMenu}
        />
      </Tooltip>,
    );

    options.push(
      <Tooltip title="Fast mode">
        <IconButton
          icon={"clock-fast"}
          iconColor={isFastMode ? theme.colors.primary : "#ccc"}
          size={25}
          key={"fastMode"}
          onPress={() => setFastMode((isFastMode) => !isFastMode)}
        />
      </Tooltip>,
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
    mutationFn: (success) => {
      if (isFastMode && success) {
        setTimeout(getNewQuiz, 400);
      } else {
        setButtonNextVisible(true);
      }

      setSnackBarProps(null);

      return wordlistsApi.answerQuiz(
        Number(wordlistId),
        Number(quiz?.id),
        success,
      );
    },
    onError: (error) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    },
    // onSuccess: (_, isAnswerCorrect) => {
    //   if (isFastMode && isAnswerCorrect) {
    //     setTimeout(getNewQuiz, 300);
    //   } else {
    //     setButtonNextVisible(true);
    //   }
    // },
  });

  const { mutate: reportError, isPending: isReportErrorPending } = useMutation<
    void,
    Error,
    ErrorType
  >({
    mutationFn: (errorType) => {
      setMenuVisible(false);
      if (!quiz) throw new Error("No quiz");

      return errorReportingApi.reportError(errorType, quiz);
    },
    onError: (error) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    },
    onSuccess: () => {
      // and then jump to the next screen
      getNewQuiz();

      setSnackBarProps({
        duration: 5000,
        onDismiss: closeSnackBar,
        message:
          "Thanks, your report will be analyzed and you will hear back from us soon.",
        type: "info",
      });
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

      <Menu visible={isMenuVisible} onDismiss={closeMenu} anchor={menuPosition}>
        <Menu.Item
          onPress={() => reportError(ErrorType.UnrelatedImage)}
          title="Unrelated image"
        />
        <Menu.Item
          onPress={() => reportError(ErrorType.MissingImage)}
          title="Missing image"
        />
        <Menu.Item
          onPress={() => reportError(ErrorType.UnrelatedMeaning)}
          title="Unrelated meaning"
        />
        <Menu.Item
          onPress={() => reportError(ErrorType.UnrelatedExample)}
          title="Unrelated example"
        />
        <Menu.Item
          onPress={() => reportError(ErrorType.SoundNotPlaying)}
          title="Sound not playing"
        />
      </Menu>
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

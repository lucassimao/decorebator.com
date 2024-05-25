import * as wordlistsApi from "@/api/wordlists";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams } from "expo-router";
import * as React from "react";
import { StyleSheet, Text, View } from "react-native";
import { Button, useTheme, ActivityIndicator, Card } from "react-native-paper";
import Snackbar, { SnackBarProps } from "@/components/SnackBar";

type Props = {
  quiz: wordlistsApi.Quiz;
  onOptionSelected: (optionIndex: number) => void;
};

const Quiz = ({ quiz, onOptionSelected }: Props) => {
  const theme = useTheme();
  const [selectedIndex, setSelectedIndex] = React.useState<number | null>(null);

  let title: string;
  // Complete sentence quiz
  if (quiz.type == 1) {
    const answer = quiz.options[quiz.answerIndex];
    const [before, after] = quiz.value.split(answer);
    title = [
      before,
      answer
        .split("")
        .map(() => `_`)
        .join(""),
      after,
    ].join(" ");
  } else {
    title = quiz.value;
  }

  const onButtonPressed = (index: number) => {
    setSelectedIndex(index);
    // onOptionSelected(index)
  };

  return (
    <Card style={{ width: "90%", height: "90%", paddingTop: 40 }}>
      <Card.Title
        titleStyle={styles.title}
        title={title}
        titleNumberOfLines={5}
      />
      <Card.Content>
        {quiz.options.map((option, index) => (
          <Button
            id={`${option}-${index}`}
            style={[
              styles.optionButton,
              selectedIndex === index
                ? quiz.answerIndex == index
                  ? styles.correctOption
                  : styles.wrongOption
                : null,
            ]}
            // labelStyle={{ fontSize: 10 }}
            rippleColor={theme.colors.inversePrimary}
            onPress={() => onButtonPressed(index)}
            mode="contained"
          >
            <View style={styles.buttonContent}>
              <Text style={styles.buttonText}>{option}</Text>
            </View>
          </Button>
        ))}
      </Card.Content>
    </Card>
  );
};

export default function QuizScreen() {
  const { wordlistId } = useLocalSearchParams();
  const theme = useTheme();

  const [snackBarProps, setSnackBarProps] =
    React.useState<SnackBarProps | null>(null);
  const closeSnackBar = () => setSnackBarProps(null);

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
    onMutate: (success) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: success ? "Correct" : "Wrong",
        type: success ? "success" : "error",
      });
    },
    onError: (error) => {
      setSnackBarProps({
        onDismiss: closeSnackBar,
        message: error.message,
        type: "error",
      });
    },
    onSuccess: () => {
      refetch();
      // setSnackBar(null)
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

  return (
    <View style={styles.container}>
      {(isPending || isLoading) && (
        <ActivityIndicator size={"large"} animating={true} theme={theme} />
      )}

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
  title: {
    fontSize: 20,
    fontWeight: "bold",
    textTransform: "capitalize",
    textAlign: "center",
    marginBottom: 30,
  },
  optionButton: {
    marginTop: 10,
    marginBottom: 20,
    padding: 5,
    alignItems: "center",
    justifyContent: "center",

    // Shadow for iOS
    shadowColor: "#000",
    shadowOffset: {
      width: 0,
      height: 4,
    },
    shadowOpacity: 0.3,
    shadowRadius: 4.65,
    // Shadow for Android
    elevation: 8,
  },

  buttonContent: {
    alignItems: "center",
  },
  buttonText: {
    textAlign: "center",
    color: "white",
    fontSize: 20,
  },

  correctOption: {
    backgroundColor: "#4caf50", // Green color for success
  },
  wrongOption: {
    backgroundColor: "#f44336", // Red color for error
  },
});

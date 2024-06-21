import * as wordlistsApi from "@/api/wordlists";
import * as React from "react";
import {
  ScrollView,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from "react-native";
import { Surface, TouchableRipple, useTheme } from "react-native-paper";

type Props = {
  quiz: wordlistsApi.Quiz;
  onOptionSelected: (optionIndex: number) => void;
};

const Quiz = ({ quiz, onOptionSelected }: Props) => {
  const { fontScale } = useWindowDimensions();
  const styles = makeStyles(fontScale); // pass in fontScale to the StyleSheet
  const theme = useTheme();
  const [selectedIndex, setSelectedIndex] = React.useState<number | null>(null);

  React.useEffect(() => {
    setSelectedIndex(null);
  }, [quiz.id]);

  React.useEffect(() => {
    if (selectedIndex != null) onOptionSelected(selectedIndex);
  }, [selectedIndex]);

  let title: string;
  // Complete sentence quiz
  if (quiz.type == 1) {
    title = quiz.value.replace(/\[(.*?)\]/g, (_, p1) => "_".repeat(p1.length));
  } else {
    title = quiz.value;
  }

  return (
    <View style={styles.container1}>
      <Surface theme={theme} elevation={5} style={styles.container}>
        <ScrollView
          contentContainerStyle={{ alignItems: "stretch" }}
          style={styles.content}
        >
          <Text
            adjustsFontSizeToFit={true}
            numberOfLines={9}
            style={[
              styles.header,
              quiz.type == 1
                ? styles.completeSentenceQuiz
                : styles.guessMeaningQuiz,
            ]}
          >
            {title}
          </Text>

          {quiz.options.map((option, index) => (
            <Surface
              theme={theme}
              style={[
                {
                  backgroundColor:
                    selectedIndex == null ? theme.colors.primary : "#fff",
                },
                styles.option,
                selectedIndex === index
                  ? quiz.answerIndex == index
                    ? styles.correctOption
                    : styles.wrongOption
                  : null,
              ]}
              key={`${option}-${index}`}
              elevation={4}
            >
              <TouchableRipple
                style={{ padding: 20 }}
                disabled={selectedIndex != null}
                theme={theme}
                rippleColor={theme.colors.inversePrimary}
                onPress={() => setSelectedIndex(index)}
              >
                <Text
                  style={[
                    styles.buttonText,
                    selectedIndex != null &&
                    selectedIndex != index &&
                    quiz.answerIndex == index
                      ? { color: "green" }
                      : null,
                  ]}
                >
                  {option}
                </Text>
              </TouchableRipple>
            </Surface>
          ))}
        </ScrollView>
      </Surface>
    </View>
  );
};

const makeStyles = (fontScale: number) =>
  StyleSheet.create({
    container1: {
      flex: 1,
      marginTop: 20,
      marginBottom: 20,
      width: "90%",
      height: "90%",
    },

    container: {
      backgroundColor: "#fff",

      //flex: 1,//
      height: 620,
      width: "100%",
      //  alignItems: "center",
      //  justifyContent: "space-between",
      borderRadius: 10,
      padding: 20,
      paddingBottom: 0,
    },
    header: {
      fontWeight: "bold",
      textAlign: "center",
      paddingTop: 10,
      marginBottom: 10,
    },
    guessMeaningQuiz: {
      fontSize: 40 / fontScale,
      textTransform: "capitalize",
    },
    completeSentenceQuiz: {
      fontSize: 25 / fontScale,
      marginBottom: 45,
    },
    content: {
      flexGrow: 1,
      flex: 1,
      width: "100%",
      padding: 10,
    },
    option: {
      marginTop: 10,
      marginBottom: 20,
      padding: 0,
      justifyContent: "center",
      borderRadius: 10,
      width: "100%",
    },

    buttonText: {
      textAlign: "center",
      fontSize: 20,
      width: "100%",
      color: "#000",
    },

    correctOption: {
      backgroundColor: "#4caf50", // Green color for success
    },
    wrongOption: {
      backgroundColor: "#f44336", // Red color for error
    },
  });

export default Quiz;

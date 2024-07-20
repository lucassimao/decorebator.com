import * as wordlistsApi from "@/api/wordlists";
import * as React from "react";
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from "react-native";
import {
  ActivityIndicator,
  Surface,
  TouchableRipple,
  useTheme,
} from "react-native-paper";

type Props = {
  quiz: wordlistsApi.Quiz;
  onOptionSelected: (optionIndex: number) => void;
  isAnsweringQuiz: boolean;
};

const Quiz = ({ quiz, onOptionSelected, isAnsweringQuiz }: Props) => {
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
  if (quiz.type == "COMPLETE_SENTENCE") {
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

          {quiz.type == 'WORD_FROM_IMAGE' ? <Image
            source={{ uri: quiz.value }}
            style={styles.image}
          /> : <Text
            adjustsFontSizeToFit={true}
            numberOfLines={9}
            style={[
              styles.header,
              ["COMPLETE_SENTENCE", 'WORD_FROM_MEANING'].includes(quiz.type)
                ? styles.defaultQuizTitle
                : styles.biggerQuizTitle,
            ]}
          >
            {title}
          </Text>}

          {["GUESS_MEANING", 'WORD_FROM_MEANING'].includes(quiz.type) && (
            <Text style={{ textAlign: "center" }}>{quiz.pos}</Text>
          )}

          {/* rendering option buttons */}
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
                onPress={() => setSelectedIndex(index)}
              >
                {/* {selectedIndex == index && isAnsweringQuiz ? (
                  <ActivityIndicator animating={true} />
                ) : ( */}
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
                {/* )} */}
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
    biggerQuizTitle: {
      fontSize: 40 / fontScale,
      textTransform: "capitalize",
    },
    defaultQuizTitle: {
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
    image: {
      width: 200,
      height: 200,
      marginLeft: 'auto',
      marginRight:'auto'
    },
  });

export default Quiz;

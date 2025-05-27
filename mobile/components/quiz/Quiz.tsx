import * as wordlistsApi from "@/api/wordlists";
import { useAudioPlayer } from "expo-audio";
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
  Button,
  IconButton,
  Surface,
  TouchableRipple,
  useTheme
} from "react-native-paper";

type Props = {
  quiz: wordlistsApi.Quiz;
  onOptionSelected: (optionIndex: number) => void;
  isAnsweringQuiz: boolean;
};

const Quiz = ({ quiz, onOptionSelected }: Props) => {
  const { fontScale } = useWindowDimensions();
  const styles = makeStyles(fontScale); // pass in fontScale to the StyleSheet
  const theme = useTheme();
  const [selectedIndex, setSelectedIndex] = React.useState<number | null>(null);
  const player = useAudioPlayer();
  const [textWidth, setTextWidth] = React.useState(0);

  React.useEffect(() => {
    setSelectedIndex(null);

    const shouldLoadAudio =
      ["WORD_FROM_AUDIO", "MEANING_FROM_AUDIO", "GUESS_MEANING"].includes(
        quiz.type,
      ) && quiz.audioURL;

    if (!shouldLoadAudio || !quiz.audioURL) return;

    player.replace(quiz.audioURL);
    player.seekTo(0)

  }, [quiz.id,player]);


  React.useEffect(() => {
    if (selectedIndex != null) {
      onOptionSelected(selectedIndex);
    }
  }, [selectedIndex]);

  let title = "";
  let quizTitleStyle;
  let displayPlayButton = false;

  switch (quiz.type) {
    case "COMPLETE_SENTENCE":
      title = quiz.value.replace(/\[(.*?)\]/g, (_, p1) =>
        "_".repeat(p1.length),
      );
      quizTitleStyle = styles.defaultQuizTitle;
      break;
    case "WORD_FROM_MEANING":
      quizTitleStyle = styles.defaultQuizTitle;
      title = quiz.value;
      break;
    case "WORD_FROM_IMAGE":
      title = quiz.imageDescription?.replace(/\[(.*?)\]/g, (_, p1) =>
        "_".repeat(p1.length),
      );
      quizTitleStyle = styles.smallQuizTitle;
      break;
    case "GUESS_MEANING":
      title = quiz.value;
      quizTitleStyle = styles.biggerQuizTitle;
      displayPlayButton = true;
      break;
    case "MEANING_FROM_AUDIO":
    case "WORD_FROM_AUDIO":
      displayPlayButton = true;
      break;
  }

  const playSound = () => player.play()


  return (
    <View style={styles.container1}>
      <Surface theme={theme} elevation={5} style={styles.container}>
        <ScrollView
          contentContainerStyle={{ alignItems: "stretch" }}
          style={styles.content}
        >
          {quiz.type == "WORD_FROM_IMAGE" && (
            <Image source={{ uri: quiz.value }} style={styles.image} />
          )}

          {["WORD_FROM_AUDIO", "MEANING_FROM_AUDIO"].includes(quiz.type) && (
            <View
              style={{
                flex: 1,
                justifyContent: "center",
                alignItems: "center",
              }}
            >
              <Button
                icon={({ color }) => (
                  <IconButton icon="play-circle" size={60} iconColor={color} />
                )}
                onPress={playSound}
              >
                <View />
              </Button>
              <Text>Press to hear word</Text>
            </View>
          )}

          {title && (
            <View
              style={{
                justifyContent: "center",
                alignItems: "center",
                flexDirection: "row",
              }}
            >
              <View
                style={{
                  position: "relative", // Allows absolute positioning for the icon
                  flexDirection: "row",
                  alignItems: "center",
                }}
              >
                <Text
                  adjustsFontSizeToFit={true}
                  numberOfLines={9}
                  style={[styles.header, quizTitleStyle]}
                  onPress={playSound}
                  onLayout={(event) => {
                    const { width } = event.nativeEvent.layout;
                    setTextWidth(width); // Update text width dynamically
                  }}
                >
                  {title}
                </Text>
                {displayPlayButton && (
                  <IconButton
                    onPress={playSound}
                    icon="play-circle"
                    iconColor={theme.colors.primary}
                    size={30}
                    loading={false}
                    style={{
                      position: "absolute",
                      left: -(textWidth / title?.length + 25),
                    }}
                  />
                )}
              </View>
            </View>
          )}

          {["GUESS_MEANING", "WORD_FROM_MEANING"].includes(quiz.type) && (
            <Text style={{ textAlign: "center" }}>{quiz.pos}</Text>
          )}

          <View style={{ marginTop: 15 }}>
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
          </View>
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
    smallQuizTitle: {
      fontSize: 15 / fontScale,
      marginBottom: 0,
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
      marginLeft: "auto",
      marginRight: "auto",
    },
  });

export default Quiz;

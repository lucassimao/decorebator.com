import * as wordlistsApi from "@/api/wordlists";
import * as React from "react";
import { StyleSheet, Text, View } from "react-native";
import { Surface, TouchableRipple, useTheme } from "react-native-paper";

type Props = {
    quiz: wordlistsApi.Quiz;
    onOptionSelected: (optionIndex: number) => void;
};

const Quiz = ({ quiz, onOptionSelected }: Props) => {
    const theme = useTheme();
    const [selectedIndex, setSelectedIndex] = React.useState<number | null>(null);

    React.useEffect(() => {
        setSelectedIndex(null)
    }, [quiz.id])

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

    return (
        <Surface  theme={theme} elevation={5} style={styles.container}>
            <Text
                style={[styles.header, quiz.type == 1 ? styles.completeSentenceQuiz : styles.guessMeaningQuiz]}
            >
                {title}
            </Text>

            <View style={styles.content}>
                {quiz.options.map((option, index) => (
                    <Surface
                        theme={theme}
                        style={[
                              { backgroundColor: selectedIndex == null ? theme.colors.primary : '#FFF' },
                            styles.option,
                            selectedIndex === index
                                ? quiz.answerIndex == index
                                    ? styles.correctOption
                                    : styles.wrongOption
                                : null,
                        ]}
                        key={`${option}-${index}`}
                        elevation={5}>
                        <TouchableRipple disabled={selectedIndex != null} theme={theme} rippleColor={theme.colors.inversePrimary}
                            onPress={() => setSelectedIndex(index)}>
                            <Text style={styles.buttonText}>{option}</Text>
                        </TouchableRipple>
                    </Surface>
                ))}
            </View>

               {selectedIndex!=null && <Surface
                    theme={theme}
                    style={[
                        { backgroundColor: theme.colors.primary },
                        styles.option,
                        styles.bottom
                    ]}
                    elevation={5}>
                    <TouchableRipple theme={theme} rippleColor={theme.colors.inversePrimary}
                        onPress={() => onOptionSelected(selectedIndex)}>
                        <Text style={styles.buttonText}>Next Quizz</Text>
                    </TouchableRipple>
                </Surface>}

        </Surface>
    );
};

const styles = StyleSheet.create({

    container: {
         flex: 1,
         marginTop: 20,
         marginBottom: 20,
         alignItems: "center",
         justifyContent:'space-between',
        width: "90%", 
        borderRadius: 10,
        padding: 20,
        paddingBottom:0
    },
    header: {
        fontWeight: "bold",
        textAlign: "center",
         paddingTop: 10,
         marginBottom:10
    },
    guessMeaningQuiz:{
        fontSize: 40,
        textTransform: "capitalize",
    },
    completeSentenceQuiz:{
        fontSize: 25,
        marginBottom:25,
    },
    content: {
        flexGrow:1,
        flex:1,
        width:'100%',
        alignItems: "stretch",
        // justifyContent:'center',
    },
    option: {
        marginTop: 10,
        marginBottom: 20,
        padding: 20,
        justifyContent: "center",
        borderRadius: 10,
        width: '100%',
    },

    buttonText: {
        textAlign: "center",
        fontSize: 20,
        width: '100%',
        color: '#000',
    },
    bottom: {
    },

    correctOption: {
        backgroundColor: "#4caf50", // Green color for success
    },
    wrongOption: {
        backgroundColor: "#f44336", // Red color for error
    },
});

export default Quiz
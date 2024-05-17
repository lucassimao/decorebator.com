
import * as wordlistsApi from '@/api/wordlists';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useLocalSearchParams } from 'expo-router';
import * as React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Button, Snackbar, useTheme, ActivityIndicator, TextInput } from 'react-native-paper';

type Props = {
    quiz: wordlistsApi.Quiz
    onOptionSelected : (optionIndex:number)=>void
}

const GuessMeaningQuiz = ({ quiz,onOptionSelected }: Props) => {
    const theme = useTheme();

    return <View style={{padding:10}}>
        <Text style={styles.title}>{quiz?.value}</Text>

        {quiz.options.map((option, index) => <Button id={option} style={styles.optionButton}
            labelStyle={{ fontSize: 20 }}
            rippleColor={theme.colors.inversePrimary}
            onPress={() => onOptionSelected(index)}
            mode="contained" >
            <View style={styles.buttonContent}>
                <Text style={styles.buttonText}>
                    {option}
                </Text>
            </View>

        </Button>)}
    </View>
}


const CompleteSentenceQuiz = ({ quiz,onOptionSelected }: Props) => {
    const theme = useTheme();
    const answer = quiz.options[quiz.answerIndex]
    const [before,after] = quiz.value.split(answer)

    return <View style={{padding:10}}>
        <Text style={styles.title}>{before}</Text>
        <Text style={styles.title}>{answer.split('').map(() => `_`).join('')}</Text>
        <Text style={styles.title}>{after}</Text>


        {quiz.options.map((option, index) => <Button id={option} style={styles.optionButton}
            labelStyle={{ fontSize: 20 }}
            rippleColor={theme.colors.inversePrimary}
            onPress={() => onOptionSelected(index)}
            mode="contained" >
            <View style={styles.buttonContent}>
                <Text style={styles.buttonText}>
                    {option}
                </Text>
            </View>

        </Button>)}
    </View>
}

type SnackBarInfo = {
    message: string,
    type: 'success' | 'error'
}
export default function QuizScreen() {
    const { wordlistId } = useLocalSearchParams();
    const theme = useTheme();

    const [snackBar, setSnackBar] = React.useState<SnackBarInfo | null>(null)

    const { data: quiz, isError, error, isLoading, refetch } = useQuery<wordlistsApi.Quiz, Error>({
        queryFn: () => wordlistsApi.newQuiz(Number(wordlistId)),
        staleTime: 0,
        queryKey: [],
        networkMode: `always`,
        refetchOnMount: false,
    })

    const { mutate: answerQuiz, isPending } = useMutation<void, Error, boolean>({
        mutationFn: (success) => wordlistsApi.answerQuiz(Number(wordlistId), Number(quiz?.id), success),
        onMutate: (success) =>{
            setSnackBar({ message: success ? 'Correct' : 'Wrong', type: success ? 'success' : 'error' })
        },
        onError: (error) => {
            setSnackBar({ message: error.message, type: 'error' })
        },
        onSuccess: () => {
            refetch()
            // setSnackBar(null)
        },
    })

    React.useEffect(() => {
        if (isError) {
            setSnackBar({ message: error.message, type: 'error' })
        }
    }, [isError, error])


    if (isPending || isLoading) {
        return <View style={styles.container}>
            <ActivityIndicator size={'large'} animating={true} theme={theme} />
        </View>
    }


    const onOptionSelected = (optionIndex:number) => answerQuiz(optionIndex == quiz?.answerIndex)

    return (
        <View  style={styles.container}>

            {quiz?.type == 0 && <GuessMeaningQuiz onOptionSelected={onOptionSelected} quiz={quiz}/>}
            {quiz?.type == 1 && <CompleteSentenceQuiz onOptionSelected={onOptionSelected} quiz={quiz}/>}

            {snackBar && <Snackbar
                visible={true}
                duration={1000}
                onDismiss={() => setSnackBar(null)}
                style={[styles.snackbar, snackBar.type === 'success' ? styles.success : styles.error]}
                action={{
                    label: 'Hide',
                }}>
                <Text style={styles.snackbarText}>{snackBar.message}</Text>
            </Snackbar>}

        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
    },
    title: {
        fontSize: 20,
        fontWeight: 'bold',
        marginBottom: 20
    },
    optionButton: {
        width: '80%',
        marginTop: 10,
        marginBottom: 20,
        padding: 10,
        alignItems: 'center',
        justifyContent: 'center',

        // Shadow for iOS
        shadowColor: '#000',
        shadowOffset: {
            width: 0,
            height: 4,
        },
        shadowOpacity: 0.3,
        shadowRadius: 4.65,
        // Shadow for Android
        elevation: 8,

    },
    signInHelper: {
        fontSize: 20,
    },

    buttonContent: {
        alignItems: 'center',
    },
    buttonText: {
        textAlign: 'center',
        color: 'white',
        fontSize: 20,
    },

    snackbar: {
        borderRadius: 4,
    },
    success: {
        backgroundColor: '#4caf50', // Green color for success
    },
    error: {
        backgroundColor: '#f44336', // Red color for error
    },
    snackbarText: {
        color: '#ffffff', // White text color
    },
});

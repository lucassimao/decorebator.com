import * as wordlistsApi from '@/api/wordlists';
import { useQuery } from "@tanstack/react-query";
import { router, useFocusEffect } from 'expo-router';
import * as React from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';
import { ActivityIndicator, List, Snackbar, Text, TouchableRipple, useTheme } from "react-native-paper";

import { BottomNavigation } from 'react-native-paper';
import AddWordDialog from './addWordDialog';
import DeleteWordlistDialog from './deleteWordlistDialog';
import DeleteWordDialog from './deleteWordDialog';

enum BottomNavigationRouteKey {
    New = 'new',
    Practice = 'practice',
    Edit = 'edit',
    Delete = 'delete'
}
const BOTTOM_NAVIGATION_ROUTES = [
    { key: BottomNavigationRouteKey.New, title: 'New word', focusedIcon: 'notebook-plus', unfocusedIcon: 'notebook-plus-outline' },
    { key: BottomNavigationRouteKey.Practice, title: 'Practice', focusedIcon: 'play-circle', unfocusedIcon: 'play-circle-outline' },
    { key: BottomNavigationRouteKey.Edit, title: 'Edit', focusedIcon: 'notebook-edit', unfocusedIcon: 'notebook-edit-outline' },
    { key: BottomNavigationRouteKey.Delete, title: 'Delete', focusedIcon: 'delete-forever', unfocusedIcon: 'delete-forever-outline' },
]

export default function Dashboard() {

    const [index, setIndex] = React.useState<number>(-1);

    const theme = useTheme();
    const [expandedWordlistId, setExpandedWordlistId] = React.useState<number | null>(null);
    const [selectedWord, setSelectedWord] = React.useState<wordlistsApi.Word | null>(null);


    useFocusEffect(
        React.useCallback(() => {
            refetchWordlists()
          return function onBlur() {
            setIndex(-1)
          };
        }, [])
      );

    const { data: wordlists, isError, error, isLoading: isLoadingWordlists,refetch: refetchWordlists } = useQuery<wordlistsApi.Wordlist[], Error>({
        queryFn: () => wordlistsApi.getUserWordlists(),
        queryKey: [
            'wordlists'
        ],
    })

    const { data: words, refetch: refetchWords, isLoading: isLoadingWords } = useQuery<wordlistsApi.Word[], Error>({
        queryFn: () => expandedWordlistId ? wordlistsApi.getWords(expandedWordlistId) : [],
        queryKey: [
            'words', `wordlist-${expandedWordlistId}`
        ],
        staleTime: 5 * 60 * 1000, // 5 minutes
        enabled: !!expandedWordlistId // Only run the query if wordlistId is not null
    })

    if (isError) {
        return <Snackbar
            onDismiss={() => router.replace('/signin')}
            visible
            action={{
                label: 'Return',
            }}>
            {error.message}
        </Snackbar>
    }

    if (isLoadingWordlists) return <ActivityIndicator size={'large'} animating={true} theme={theme} />
    if (!wordlists?.length) return <Text>No wordlists - improve</Text>

    const onListAccordionPressed = (wordlistId: number) => {
        // user pressed same item. Closing it
        if (expandedWordlistId == wordlistId) {
            setExpandedWordlistId(null);
            return
        }
        setExpandedWordlistId(wordlistId);
        refetchWords();
    };


    const currentWordlist = wordlists.find(w => w.id == expandedWordlistId)
    const routeKeySelected = index >= 0 ? BOTTOM_NAVIGATION_ROUTES[index]?.key : undefined
    const isAddWordDialogVisible = expandedWordlistId && routeKeySelected == BottomNavigationRouteKey.New
    const isDeleteWordlistDialogVisible = currentWordlist && expandedWordlistId && routeKeySelected == BottomNavigationRouteKey.Delete
    const isDeleteWordDialogVisible = !!selectedWord

    if (routeKeySelected == BottomNavigationRouteKey.Practice){
        router.push(`/dashboard/quiz/${expandedWordlistId}`)
    }

    const onDismissAddWordDialog = (success?: boolean) => {
        if (success) {
            refetchWords();
        }
        setIndex(-1)
    }

    const onDismissDeleteWordlistDialog = (success?: boolean) => {
        if (success) {
            refetchWordlists();
        }
        setIndex(-1)
    }

    const onDismissDeleteWordDialog = (success?: boolean) => {
        if (success) {
            refetchWords();
        }
        setSelectedWord(null)
    }

    return <View style={styles.container}>

        {isAddWordDialogVisible && <AddWordDialog wordlistId={expandedWordlistId} onDismiss={onDismissAddWordDialog} />}
        {isDeleteWordlistDialogVisible && <DeleteWordlistDialog wordlist={currentWordlist} onDismiss={onDismissDeleteWordlistDialog} />}
        {isDeleteWordDialogVisible && <DeleteWordDialog word={selectedWord} onDismiss={onDismissDeleteWordDialog} />}

        <ScrollView contentContainerStyle={styles.scrollViewContent} >
            <List.Section>
                {wordlists.map(wordlist => {
                    return <List.Accordion expanded={wordlist.id == expandedWordlistId} onPress={() => onListAccordionPressed(wordlist.id)}
                        key={`wordlists-${wordlist.id}`}
                        title={wordlist.name}
                        description={wordlist.description}
                        left={props => <List.Icon {...props} icon="folder" />}>
                        {isLoadingWords && <View style={styles.wordsActivityIndicator}>
                            <ActivityIndicator animating={true} theme={theme} />
                        </View>
                        }
                        {words?.map(w => <TouchableRipple
                            centered
                            key={`wordlist-${wordlist.id}-${w.id}`}
                            onLongPress={() => setSelectedWord(w)}
                        >
                            <List.Item title={w.name} />
                        </TouchableRipple>)}
                    </List.Accordion>
                })}

            </List.Section>

        </ScrollView>
        {expandedWordlistId && <BottomNavigation.Bar shifting={false}
            style={styles.bottomNavigation}
            navigationState={{ index, routes: BOTTOM_NAVIGATION_ROUTES }}
            onTabPress={({ route }) => setIndex(BOTTOM_NAVIGATION_ROUTES.findIndex(r => r.key == route.key))}
        />}
    </View>
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
    },
    scrollViewContent: {
        paddingBottom: 56, // Ensure there's enough space for the BottomNavigation
        margin: 10
    },
    bottomNavigation: {
        position: 'absolute',
        left: 0,
        right: 0,
        bottom: 0,
    },
    wordsActivityIndicator: {
        flex: 1,
        alignItems: 'center',
        justifyContent: 'center',
        height: 50
    },
});


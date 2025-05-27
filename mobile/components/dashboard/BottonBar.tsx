import * as wordlistsApi from "@/api/wordlists";
import * as usersApi from "@/api/users";
import { router, useFocusEffect } from "expo-router";
import * as React from "react";
import { StyleSheet, View } from "react-native";

import { BottomNavigation } from "react-native-paper";
import AddWordDialog from "./AddWordDialog";
import DeleteWordlistDialog from "./DeleteWordlistDialog";

enum NavigationRouteKey {
  New = "new",
  Practice = "practice",
  Edit = "edit",
  Delete = "delete",
}

const NAVIGATION_ROUTES = [
  {
    key: NavigationRouteKey.New,
    title: "New word",
    focusedIcon: "notebook-edit",
    unfocusedIcon: "notebook-edit-outline",
  },
  {
    key: NavigationRouteKey.Practice,
    title: "Practice",
    focusedIcon: "play-circle",
    unfocusedIcon: "play-circle-outline",
  },
  {
    key: NavigationRouteKey.Edit,
    title: "Edit",
    focusedIcon: "notebook-edit",
    unfocusedIcon: "notebook-edit-outline",
  },
  {
    key: NavigationRouteKey.Delete,
    title: "Delete",
    focusedIcon: "delete-forever",
    unfocusedIcon: "delete-forever-outline",
  },
];

type Props = {
  wordlist: wordlistsApi.Wordlist;
  onWordAdded: () => void;
  onWordlistDeleted: () => void;
  words?: wordlistsApi.Word[];
  onUpgradePrompt?: () => void;
  user?: usersApi.UserInfo|null;
};

export default function BottonBar({
  wordlist,
  onWordAdded,
  onWordlistDeleted,
  words,
  onUpgradePrompt,
  user,
}: Props) {
  const [selectedRoute, setSelectedRoute] =
    React.useState<NavigationRouteKey | null>(null);

  // clearing selected route whenever mounting the component
  useFocusEffect(React.useCallback(() => setSelectedRoute(null), []));

  if (selectedRoute == NavigationRouteKey.Practice) {
    router.push(`/quiz/${wordlist.id}`);
  }

  const displayAddWordDialog = selectedRoute == NavigationRouteKey.New;
  const displayDeleteWordlistDialog =
    selectedRoute == NavigationRouteKey.Delete;

  const onDismissAddWordDialog = (dismiss?: boolean) => {
    onWordAdded();
    if (dismiss) {
      setSelectedRoute(null);
    }
  };

  const onDismissDeleteWordlistDialog = (success?: boolean) => {
    if (success) {
      onWordlistDeleted();
    }
    setSelectedRoute(null);
  };

  const index =
    selectedRoute != null
      ? NAVIGATION_ROUTES.findIndex((r) => r.key == selectedRoute)
      : -1;

  return (
    <View style={styles.container}>
      {displayAddWordDialog && (
        <AddWordDialog
          wordlistId={wordlist.id}
          onDismiss={() => setSelectedRoute(null)}
          onWordAdded={onDismissAddWordDialog}
        />
      )}
      {displayDeleteWordlistDialog && (
        <DeleteWordlistDialog
          wordlist={wordlist}
          onDismiss={onDismissDeleteWordlistDialog}
        />
      )}

      <BottomNavigation.Bar
        shifting={false}
        navigationState={{ index, routes: NAVIGATION_ROUTES }}
        onTabPress={({ route }) => {
          // Check word limit for free users
          if (route.key === NavigationRouteKey.New) {
            const currentUser = user || usersApi.getUserInfo();
            const isFreePlan = !currentUser?.subscriptionPlan || currentUser.subscriptionPlan === 'free';
            const wordCount = words?.length || 0;
            
            if (isFreePlan && wordCount >= 10) {
              // Free users can only have 10 words per wordlist
              onUpgradePrompt?.();
              return;
            }
          }
          setSelectedRoute(route.key);
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    position: "absolute",
    left: 0,
    right: 0,
    bottom: 0,
  },

  snackbar: {
    borderRadius: 4,
  },
  success: {
    backgroundColor: "#4caf50", // Green color for success
  },
  error: {
    backgroundColor: "#f44336", // Red color for error
  },
  snackbarText: {
    color: "#ffffff", // White text color
  },
});

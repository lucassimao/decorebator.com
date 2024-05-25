import * as React from "react";
import { StyleSheet, View } from "react-native";
import { ActivityIndicator, useTheme } from "react-native-paper";

export default function LoadingIndicator() {
  const theme = useTheme();

  return (
    <View style={styles.container}>
      <ActivityIndicator size={"large"} animating={true} theme={theme} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
});

import { useUserInfo } from "@/hooks/users";
import { router } from "expo-router";
import React, { useEffect } from "react";
import {
  ActivityIndicator,
  Dimensions,
  ImageBackground,
  SafeAreaView,
  StyleSheet,
  View,
} from "react-native";
import * as usersApi from "@/api/users";

export default function Index() {
  const { userInfo, error, loading } = useUserInfo();

  useEffect(() => {
    if (loading) return;

    if (userInfo) {
      router.replace("/dashboard");
    } else {
      usersApi.sigout();
      router.replace("/signin");
    }
  }, [userInfo, loading]);

  useEffect(() => {
    if (error) {
      console.error(error);

      usersApi.sigout();
      router.replace("/signin");
    }
  }, [error]);

  return (
    <ImageBackground
      source={require("@/assets/images/dashboard-bg.png")}
      style={styles.backgroundImage}
      resizeMode="cover"
    >
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#FF7B54" />
        </View>
      </SafeAreaView>
    </ImageBackground>
  );
}
const { width, height } = Dimensions.get("window");

const styles = StyleSheet.create({
  backgroundImage: {
    flex: 1,
    width: width,
    height: height,
  },
  container: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
});

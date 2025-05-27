import * as subscriptionsApi from "@/api/subscriptions";
import * as usersApi from "@/api/users";
import { useFocusEffect } from "@react-navigation/native";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import * as WebBrowser from "expo-web-browser";
import React, { useState } from "react";
import { ScrollView, StyleSheet, View } from "react-native";
import {
  ActivityIndicator,
  Button,
  Card,
  Chip,
  List,
  Snackbar,
  Text,
  useTheme,
} from "react-native-paper";
import Icon from "react-native-vector-icons/MaterialCommunityIcons";
import * as AuthSession from "expo-auth-session";

export default function SubscriptionScreen() {
  const theme = useTheme();
  const queryClient = useQueryClient();
  const [user, setUser] = useState(usersApi.getUserInfo());
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showSuccessMessage, setShowSuccessMessage] = useState(false);

  // Fetch current subscription status
  const {
    data: subscriptionStatus,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ["subscriptionStatus"],
    queryFn: subscriptionsApi.getSubscriptionStatus,
  });

  // Refresh user session when screen comes into focus
  useFocusEffect(
    React.useCallback(() => {
      const refreshUserSession = async () => {
        try {
          setIsRefreshing(true);
          const updatedUser = await usersApi.refreshToken();
          setUser(updatedUser);

          // Refetch subscription status
          await refetch();

          // Invalidate queries that depend on subscription status
          queryClient.invalidateQueries({ queryKey: ["wordlists"] });

          // Check if user just upgraded
          if (
            updatedUser.subscriptionPlan !== "free" &&
            user?.subscriptionPlan === "free"
          ) {
            setShowSuccessMessage(true);
          }
        } catch (error) {
          console.error("Failed to refresh session:", error);
        } finally {
          setIsRefreshing(false);
        }
      };

      refreshUserSession();
    }, []),
  );

  const handleSubscribe = async (plan: "monthly" | "annual") => {
    try {
      const redirectUri = AuthSession.makeRedirectUri({
        scheme: "decorerbator",
        path: "stripe",
      });

      const { checkoutUrl } = await subscriptionsApi.createCheckoutSession(
        plan,
        redirectUri,
      );
      const result = await WebBrowser.openAuthSessionAsync(
        checkoutUrl,
        redirectUri,
      );

      // When browser is dismissed, refresh the session
      if (result.type === "success") {
        const updatedUser = await usersApi.refreshToken();
        setShowSuccessMessage(true);
        setUser(updatedUser);
        await refetch();
      }
    } catch (error) {
      console.error("Error opening checkout:", error);
      alert("Failed to open checkout. Please try again.");
    }
  };

  const features = [
    "Unlimited wordlists",
    "AI-powered word enrichment",
    "All quiz modes unlocked",
    "Progress tracking",
    "Priority support",
  ];

  const freeFeatures = ["1 wordlist", "Up to 10 words", "Basic quiz modes"];

  if (isLoading || isRefreshing) {
    return (
      <View style={[styles.container, styles.centerContent]}>
        <ActivityIndicator size="large" />
        <Text style={styles.loadingText}>Loading subscription info...</Text>
      </View>
    );
  }

  // Check if user is already subscribed
  const isSubscribed =
    user?.subscriptionPlan && user.subscriptionPlan !== "free";

  return (
    <>
      <ScrollView style={styles.container}>
        <View style={styles.header}>
          <Icon name="crown" size={80} color={theme.colors.primary} />
          <Text variant="headlineMedium" style={styles.title}>
            {isSubscribed
              ? "Manage Your Subscription"
              : "Unlock Full Learning Potential"}
          </Text>
          <Text variant="bodyLarge" style={styles.subtitle}>
            {isSubscribed
              ? `Current plan: ${user?.subscriptionPlan === "monthly" ? "Monthly" : "Annual"} subscription`
              : "Expand your vocabulary without limits"}
          </Text>
        </View>

        {/* Free Plan Info */}
        <Card style={[styles.card, styles.freeCard]}>
          <Card.Content>
            <View style={styles.planHeader}>
              <Text variant="titleLarge">Free Plan</Text>
              <Text variant="bodyMedium" style={styles.currentPlan}>
                Current Plan
              </Text>
            </View>
            <List.Section>
              {freeFeatures.map((feature, index) => (
                <List.Item
                  key={index}
                  title={feature}
                  titleStyle={styles.featureText}
                  left={() => (
                    <List.Icon
                      icon="check-circle"
                      color={theme.colors.outline}
                    />
                  )}
                />
              ))}
            </List.Section>
          </Card.Content>
        </Card>

        {/* Premium Plans */}
        <View style={styles.premiumSection}>
          <Text variant="titleLarge" style={styles.sectionTitle}>
            Choose Your Plan
          </Text>

          {/* Monthly Plan */}
          <Card style={[styles.card, styles.planCard]}>
            <Card.Content>
              <View style={styles.planHeader}>
                <Text variant="titleLarge">Monthly</Text>
                <View style={styles.priceContainer}>
                  <Text variant="headlineMedium" style={styles.price}>
                    $6.99
                  </Text>
                  <Text variant="bodyMedium" style={styles.period}>
                    /month
                  </Text>
                </View>
              </View>
              <List.Section>
                {features.map((feature, index) => (
                  <List.Item
                    key={index}
                    title={feature}
                    titleStyle={styles.featureText}
                    left={() => (
                      <List.Icon
                        icon="check-circle"
                        color={theme.colors.primary}
                      />
                    )}
                  />
                ))}
              </List.Section>
              <Button
                mode="contained"
                onPress={() => handleSubscribe("monthly")}
                style={styles.subscribeButton}
              >
                Subscribe Monthly
              </Button>
            </Card.Content>
          </Card>

          {/* Annual Plan */}
          <Card style={[styles.card, styles.planCard, styles.bestValue]}>
            <View style={styles.bestValueBadge}>
              <Chip style={styles.chip} textStyle={styles.chipText} mode="flat">
                BEST VALUE
              </Chip>
            </View>
            <Card.Content>
              <View style={styles.planHeader}>
                <Text variant="titleLarge">Annual</Text>
                <View style={styles.priceContainer}>
                  <Text variant="headlineMedium" style={styles.price}>
                    $69.9
                  </Text>
                  <Text variant="bodyMedium" style={styles.period}>
                    /year
                  </Text>
                </View>
              </View>
              <Text variant="bodyMedium" style={styles.savings}>
                Save $13.98 compared to monthly!
              </Text>
              <List.Section>
                {[...features, "Early access to new features"].map(
                  (feature, index) => (
                    <List.Item
                      key={index}
                      title={feature}
                      titleStyle={styles.featureText}
                      left={() => (
                        <List.Icon
                          icon="check-circle"
                          color={theme.colors.primary}
                        />
                      )}
                    />
                  ),
                )}
              </List.Section>
              <Button
                mode="contained"
                onPress={() => handleSubscribe("annual")}
                style={styles.subscribeButton}
              >
                Subscribe Annually
              </Button>
            </Card.Content>
          </Card>
        </View>

        <Text variant="bodySmall" style={styles.disclaimer}>
         • Cancel anytime{"\n"}• Prices in USD
        </Text>
      </ScrollView>

      <Snackbar
        visible={showSuccessMessage}
        onDismiss={() => setShowSuccessMessage(false)}
        duration={5000}
        style={styles.successSnackbar}
        action={{
          label: "Go to Dashboard",
          onPress: () => {
            setShowSuccessMessage(false);
            router.push("/dashboard");
          },
        }}
      >
        🎉 Subscription activated! You now have unlimited access.
      </Snackbar>
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f5f5f5",
  },
  header: {
    alignItems: "center",
    paddingVertical: 30,
    paddingHorizontal: 20,
  },
  title: {
    marginTop: 10,
    textAlign: "center",
    fontWeight: "bold",
  },
  subtitle: {
    marginTop: 5,
    textAlign: "center",
    opacity: 0.7,
  },
  card: {
    marginHorizontal: 20,
    marginBottom: 15,
    elevation: 2,
  },
  freeCard: {
    backgroundColor: "#f9f9f9",
  },
  planCard: {
    position: "relative",
  },
  bestValue: {
    borderWidth: 2,
    borderColor: "#6200ee",
  },
  bestValueBadge: {
    position: "absolute",
    top: -12,
    alignSelf: "center",
    zIndex: 1,
  },
  chip: {
    backgroundColor: "#6200ee",
  },
  chipText: {
    color: "white",
    fontSize: 11,
    fontWeight: "bold",
  },
  planHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 10,
  },
  currentPlan: {
    color: "#666",
    fontStyle: "italic",
  },
  priceContainer: {
    flexDirection: "row",
    alignItems: "baseline",
  },
  price: {
    fontWeight: "bold",
    color: "#6200ee",
  },
  period: {
    marginLeft: 5,
    opacity: 0.7,
  },
  savings: {
    color: "#4caf50",
    fontWeight: "600",
    marginBottom: 10,
  },
  featureText: {
    fontSize: 14,
  },
  subscribeButton: {
    marginTop: 15,
    paddingVertical: 5,
  },
  premiumSection: {
    marginTop: 20,
  },
  sectionTitle: {
    marginLeft: 20,
    marginBottom: 15,
    fontWeight: "bold",
  },
  disclaimer: {
    textAlign: "center",
    margin: 20,
    opacity: 0.6,
    lineHeight: 20,
  },
  centerContent: {
    justifyContent: "center",
    alignItems: "center",
  },
  loadingText: {
    marginTop: 10,
    opacity: 0.6,
  },
  successSnackbar: {
    backgroundColor: "#4caf50",
  },
});

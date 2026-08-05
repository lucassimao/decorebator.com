import React from "react";
import {
  AccessibilityInfo,
  Linking,
  Pressable,
  ScrollView,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import Animated, {
  useAnimatedStyle,
  useReducedMotion,
  useSharedValue,
  withDelay,
  withSpring,
} from "react-native-reanimated";
import { SafeAreaView } from "react-native-safe-area-context";

import { Button, Card, UiText } from "@/components/ui";
import { useTheme } from "@/contexts/ThemeContext";
import { useNativeIAP, type NativeIAPStatus } from "@/hooks/useNativeIAP";
import i18n from "@/i18n";
import type { NativeIAPPlan } from "@/utils/iapProducts";

interface NativeIAPPaywallProps {
  visible: boolean;
  source?: string;
  onClose: () => void;
  onEntitled: () => void | Promise<void>;
}

const LOCKED_STATUSES = new Set<NativeIAPStatus>([
  "purchasing",
  "pending",
  "entitled",
  "restoring",
]);

export default function NativeIAPPaywall({
  visible,
  source = "settings",
  onClose,
  onEntitled,
}: NativeIAPPaywallProps) {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const { state, store, selectPlan, purchase, restore, retry } = useNativeIAP(
    visible,
    source,
  );
  const entitledNotified = React.useRef(false);
  const entitledAnnounced = React.useRef(false);
  const locked = LOCKED_STATUSES.has(state.status);
  const commerceVisible = [
    "ready",
    "purchasing",
    "pending",
    "entitled",
    "restoring",
  ].includes(state.status);
  const selectedPlan = commerceVisible
    ? state.plans.find((plan) => plan.productId === state.selectedProductId)
    : undefined;

  React.useEffect(() => {
    if (state.status !== "entitled") return;
    if (!entitledAnnounced.current) {
      entitledAnnounced.current = true;
      AccessibilityInfo.announceForAccessibility(
        t("settings.subscription.nativeIap.entitled"),
      );
    }
    if (!state.entitlementOrigin || entitledNotified.current) return;
    entitledNotified.current = true;
    void onEntitled();
  }, [onEntitled, state.entitlementOrigin, state.status, t]);

  React.useEffect(() => {
    if (!visible) {
      entitledNotified.current = false;
      entitledAnnounced.current = false;
    }
  }, [visible]);

  const openLegal = (path: "terms" | "privacy") =>
    Linking.openURL(`https://decorebator.com/${i18n.language}/${path}`);

  return (
    <SafeAreaView
      style={{ flex: 1, backgroundColor: theme.colors.background.default }}
    >
      <View
        style={{
          minHeight: theme.geometry.touchTarget,
          paddingHorizontal: theme.spacing.md,
          flexDirection: "row",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <UiText variant="label" tone="secondary">
          {store === "google"
            ? t("settings.subscription.nativeIap.googlePlay")
            : t("settings.subscription.nativeIap.appStore")}
        </UiText>
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={t("common.close")}
          hitSlop={8}
          onPress={onClose}
          style={{
            width: theme.geometry.touchTarget,
            height: theme.geometry.touchTarget,
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          <Ionicons name="close" size={26} color={theme.colors.text.primary} />
        </Pressable>
      </View>

      <ScrollView
        testID="native-iap-scroll"
        contentContainerStyle={{
          paddingHorizontal: theme.spacing.md,
          paddingBottom: theme.spacing.xl,
          gap: theme.spacing.comfortable,
        }}
      >
        <View style={{ gap: theme.spacing.sm }}>
          <UiText variant="display">
            {t("settings.subscription.nativeIap.title")}
          </UiText>
          <UiText variant="body" tone="secondary">
            {t("settings.subscription.nativeIap.subtitle")}
          </UiText>
        </View>

        <StudyShelf ready={state.status !== "loading"} />

        <StatusMessage status={state.status} />

        {state.status === "loading" ? <PlanSkeletons /> : null}

        {commerceVisible && state.plans.length > 0 ? (
          <View
            accessibilityRole="radiogroup"
            accessibilityLabel={t(
              "settings.subscription.nativeIap.planGroupLabel",
            )}
            style={{ gap: theme.spacing.compact }}
          >
            {state.plans.map((plan) => (
              <PlanCard
                key={plan.productId}
                plan={plan}
                selected={plan.productId === state.selectedProductId}
                disabled={locked}
                onSelect={() => selectPlan(plan.productId)}
              />
            ))}
          </View>
        ) : null}

        {selectedPlan ? <RenewalDisclosure plan={selectedPlan} /> : null}

        <View style={{ gap: theme.spacing.sm }}>
          <Button
            testID="native-iap-purchase"
            disabled={state.status !== "ready" || !selectedPlan}
            loading={state.status === "purchasing"}
            loadingLabel={t("settings.subscription.nativeIap.purchasing")}
            onPress={() => void purchase()}
          >
            {state.status === "entitled"
              ? t("settings.subscription.nativeIap.entitled")
              : state.status === "pending"
                ? t("settings.subscription.nativeIap.pending")
                : selectedPlan
                  ? t("settings.subscription.nativeIap.continue")
                  : t("settings.subscription.nativeIap.choosePlan")}
          </Button>

          {(state.status === "failed" ||
            state.status === "unavailable" ||
            state.status === "empty") &&
          state.retryable ? (
            <Button variant="secondary" onPress={() => void retry()}>
              {t("settings.subscription.nativeIap.retry")}
            </Button>
          ) : null}

          <Button
            testID="native-iap-restore"
            variant="quiet"
            loading={state.status === "restoring"}
            loadingLabel={t("settings.subscription.nativeIap.restoring")}
            disabled={
              !store ||
              state.status === "purchasing" ||
              state.status === "pending"
            }
            onPress={() => void restore()}
          >
            {t("settings.subscription.restorePurchases")}
          </Button>
        </View>

        <View style={{ alignItems: "center", gap: theme.spacing.xs }}>
          <UiText
            variant="small"
            tone="secondary"
            style={{ textAlign: "center" }}
          >
            {t("settings.subscription.nativeIap.legal")}
          </UiText>
          <View
            style={{
              flexDirection: "row",
              flexWrap: "wrap",
              justifyContent: "center",
            }}
          >
            <Button variant="quiet" onPress={() => void openLegal("terms")}>
              {t("settings.termsOfService")}
            </Button>
            <Button variant="quiet" onPress={() => void openLegal("privacy")}>
              {t("settings.privacyPolicy")}
            </Button>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

function PlanCard({
  plan,
  selected,
  disabled,
  onSelect,
}: {
  plan: NativeIAPPlan;
  selected: boolean;
  disabled: boolean;
  onSelect: () => void;
}) {
  const { t } = useTranslation();
  const { theme } = useTheme();
  const period = formatPeriod(plan, t);
  return (
    <Pressable
      accessibilityRole="radio"
      accessibilityState={{ selected, disabled }}
      accessibilityLabel={`${plan.title}, ${plan.displayPrice}, ${period}`}
      disabled={disabled}
      onPress={onSelect}
      style={({ pressed }) => ({
        minHeight: 88,
        padding: theme.spacing.md,
        borderWidth: selected ? 3 : 1,
        borderColor: selected
          ? theme.colors.roles.action
          : theme.colors.roles.controlBorder,
        borderRadius: theme.borderRadius.lg,
        backgroundColor: pressed
          ? theme.colors.background.elevated
          : theme.colors.background.surface,
        opacity: disabled && !selected ? 0.62 : 1,
        flexDirection: "row",
        alignItems: "center",
        gap: theme.spacing.compact,
      })}
    >
      <View
        style={{
          minWidth: "58%",
          flex: 1,
          flexShrink: 1,
          gap: theme.spacing.xs,
        }}
      >
        <UiText variant="heading">{plan.title}</UiText>
        <UiText variant="small" tone="secondary">
          {period}
        </UiText>
        {plan.offerDisplay ? (
          <UiText variant="label" tone="action">
            {t("settings.subscription.nativeIap.offer", {
              price: plan.offerDisplay.displayPrice,
              period: t(
                `settings.subscription.nativeIap.period.${plan.offerDisplay.period.unit}`,
                { count: plan.offerDisplay.period.value },
              ),
            })}
          </UiText>
        ) : null}
      </View>
      <View
        style={{
          maxWidth: "100%",
          flexShrink: 1,
          alignItems: "flex-end",
          gap: theme.spacing.sm,
        }}
      >
        <UiText variant="title" style={{ flexShrink: 1, textAlign: "right" }}>
          {plan.displayPrice}
        </UiText>
        <View
          testID={`native-iap-radio-${plan.productId}`}
          style={{
            width: 24,
            height: 24,
            borderRadius: 12,
            borderWidth: 2,
            borderColor: selected
              ? theme.colors.roles.action
              : theme.colors.roles.controlBorder,
            alignItems: "center",
            justifyContent: "center",
          }}
        >
          {selected ? (
            <View
              style={{
                width: 12,
                height: 12,
                borderRadius: 6,
                backgroundColor: theme.colors.roles.action,
              }}
            />
          ) : null}
        </View>
      </View>
    </Pressable>
  );
}

function RenewalDisclosure({ plan }: { plan: NativeIAPPlan }) {
  const { t } = useTranslation();
  return (
    <Card variant="emphasis" accessibilityLiveRegion="polite">
      <UiText variant="small">
        {t("settings.subscription.nativeIap.renewal", {
          price: plan.displayPrice,
          period: formatPeriod(plan, t),
        })}
      </UiText>
    </Card>
  );
}

function StatusMessage({ status }: { status: NativeIAPStatus }) {
  const { t } = useTranslation();
  const message = (() => {
    switch (status) {
      case "pending":
        return t("settings.subscription.nativeIap.pendingMessage");
      case "entitled":
        return t("settings.subscription.nativeIap.entitledMessage");
      case "empty":
        return t("settings.subscription.nativeIap.empty");
      case "unavailable":
        return t("settings.subscription.nativeIap.unavailable");
      case "failed":
        return t("settings.subscription.nativeIap.failure");
      default:
        return null;
    }
  })();
  return message ? (
    <Card
      variant={status === "entitled" ? "emphasis" : "surface"}
      accessibilityLiveRegion="polite"
      accessibilityRole="alert"
    >
      <UiText>{message}</UiText>
    </Card>
  ) : null;
}

function PlanSkeletons() {
  const { t } = useTranslation();
  const { theme } = useTheme();
  return (
    <View
      testID="native-iap-loading"
      accessibilityLabel={t("settings.subscription.nativeIap.loading")}
      style={{ gap: theme.spacing.compact }}
    >
      {[0, 1].map((item) => (
        <View
          key={item}
          style={{
            height: 88,
            borderRadius: theme.borderRadius.lg,
            backgroundColor: theme.colors.roles.disabledBackground,
          }}
        />
      ))}
    </View>
  );
}

function StudyShelf({ ready }: { ready: boolean }) {
  const { theme } = useTheme();
  return (
    <View
      importantForAccessibility="no-hide-descendants"
      style={{ height: 92, justifyContent: "flex-end" }}
    >
      <View
        style={{
          height: 8,
          borderRadius: theme.borderRadius.full,
          backgroundColor: theme.colors.text.primary,
        }}
      />
      <View
        style={{
          position: "absolute",
          left: theme.spacing.lg,
          right: theme.spacing.lg,
          bottom: 8,
          flexDirection: "row",
          alignItems: "flex-end",
          justifyContent: "center",
          gap: theme.spacing.sm,
        }}
      >
        {[54, 70, 60].map((height, index) => (
          <ShelfCard key={height} height={height} index={index} ready={ready} />
        ))}
        <View
          style={{
            width: 12,
            height: 62,
            marginLeft: -18,
            backgroundColor: theme.colors.roles.brandAccent,
            borderTopLeftRadius: 3,
            borderTopRightRadius: 3,
          }}
        />
      </View>
    </View>
  );
}

function ShelfCard({
  height,
  index,
  ready,
}: {
  height: number;
  index: number;
  ready: boolean;
}) {
  const { theme } = useTheme();
  const reduceMotion = useReducedMotion();
  const settled = useSharedValue(reduceMotion || !ready ? 1 : 0);
  React.useEffect(() => {
    if (reduceMotion || !ready) {
      settled.value = 1;
      return;
    }
    settled.value = 0;
    settled.value = withDelay(
      index * 55,
      withSpring(1, { damping: 16, stiffness: 210 }),
    );
  }, [index, ready, reduceMotion, settled]);
  const style = useAnimatedStyle(() => ({
    opacity: settled.value,
    transform: [
      { translateY: (1 - settled.value) * -14 },
      { rotate: `${(1 - settled.value) * (index - 1) * 3}deg` },
    ],
  }));
  return (
    <Animated.View
      style={[
        {
          width: 58,
          height,
          borderWidth: 2,
          borderColor: theme.colors.text.primary,
          borderRadius: theme.borderRadius.sm,
          backgroundColor:
            index === 1
              ? theme.colors.roles.emphasisSurface
              : theme.colors.background.surface,
        },
        style,
      ]}
    />
  );
}

function formatPeriod(
  plan: NativeIAPPlan,
  t: ReturnType<typeof useTranslation>["t"],
) {
  return t(`settings.subscription.nativeIap.period.${plan.period.unit}`, {
    count: plan.period.value,
  });
}

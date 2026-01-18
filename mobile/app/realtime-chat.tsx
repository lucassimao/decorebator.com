import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";
import { useTheme } from "@/contexts/ThemeContext";
import { ConnectionState, useRealtimeChat } from "@/hooks/useRealtimeChat";
import { SafeAreaView } from "react-native-safe-area-context";
import {
  EventCallbacks,
  RealtimeEventHandler,
} from "@/utils/realtimeEventHandler";
import { MaterialIcons } from "@expo/vector-icons";
import { useQuery } from "@tanstack/react-query";
import { Stack, useLocalSearchParams, useRouter } from "expo-router";
import LottieView from "lottie-react-native";
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import {
  Animated,
  Platform,
  ScrollView,
  StatusBar,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { requestRecordingPermissionsAsync } from "expo-audio";
import {
  ChatSessionData,
  createChatSession,
  getDefinitionsForWords,
  getWordlist,
} from "../api/wordlists";
import type { WordWithDefinitions as ApiWordWithDefinitions } from "../api/wordlists";
import ScreenHeader from "@/components/common/ScreenHeader";
import { RealtimeChatTelemetryBuilder } from "@/utils/realtimeTelemetry";
import { sendRealtimeChatTelemetry } from "@/api/telemetry";
import { LinearGradient } from "expo-linear-gradient";

export default function RealtimeChatScreen() {
  const { wordlistId, wordlistName, selectedWordIds } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
    selectedWordIds: string;
  }>();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  // Animation refs
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const pulseAnim = useRef(new Animated.Value(1)).current;
  const orbGlowAnim = useRef(new Animated.Value(0.3)).current;

  // State
  const [sessionData, setSessionData] = useState<ChatSessionData | null>(null);
  const [selectedWords, setSelectedWords] = useState<ApiWordWithDefinitions[]>(
    [],
  );
  const [loading, setLoading] = useState(true);
  const [hasTimeout, setHasTimeout] = useState(false);
  const [initError, setInitError] = useState<Error | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>({
    status: "disconnected",
  });
  const [transcript, setTranscript] = useState<string>("");
  const [transcriptHistory, setTranscriptHistory] = useState<string[]>([]);
  const [showTranscript, setShowTranscript] = useState<boolean>(true);
  const [isSpeaking, setIsSpeaking] = useState(false);
  const [isAssistantSpeaking, setIsAssistantSpeaking] = useState(false);
  const transcriptRef = useRef<string>("");
  const transcriptScrollRef = useRef<ScrollView | null>(null);
  const audioFinishedRef = useRef<boolean>(false);
  const telemetryRef = useRef<RealtimeChatTelemetryBuilder | null>(null);
  const telemetrySentRef = useRef(false);

  // Entrance animation
  useEffect(() => {
    Animated.timing(fadeAnim, {
      toValue: 1,
      duration: 600,
      useNativeDriver: true,
    }).start();
  }, [fadeAnim]);

  // Pulsing animation for the orb glow
  useEffect(() => {
    const pulse = Animated.loop(
      Animated.sequence([
        Animated.timing(orbGlowAnim, {
          toValue: 0.6,
          duration: 1500,
          useNativeDriver: true,
        }),
        Animated.timing(orbGlowAnim, {
          toValue: 0.3,
          duration: 1500,
          useNativeDriver: true,
        }),
      ]),
    );
    pulse.start();
    return () => pulse.stop();
  }, [orbGlowAnim]);

  // Active speaking pulse
  useEffect(() => {
    if (isAssistantSpeaking || isSpeaking) {
      const activePulse = Animated.loop(
        Animated.sequence([
          Animated.timing(pulseAnim, {
            toValue: 1.05,
            duration: 300,
            useNativeDriver: true,
          }),
          Animated.timing(pulseAnim, {
            toValue: 1,
            duration: 300,
            useNativeDriver: true,
          }),
        ]),
      );
      activePulse.start();
      return () => activePulse.stop();
    } else {
      pulseAnim.setValue(1);
    }
  }, [isAssistantSpeaking, isSpeaking, pulseAnim]);

  // Parse selected word IDs
  const selectedWordIdList = useMemo(() => {
    if (!selectedWordIds) return [];
    return selectedWordIds
      .split(",")
      .map((id) => parseInt(id.trim(), 10))
      .filter((id) => !isNaN(id));
  }, [selectedWordIds]);

  // Redirect to word selection if no words are selected
  useEffect(() => {
    if (selectedWordIds === undefined && wordlistId) {
      router.replace(
        `/word-selection?wordlistId=${wordlistId}&wordlistName=${encodeURIComponent(wordlistName || "")}`,
      );
    }
  }, [selectedWordIds, wordlistId, wordlistName, router]);

  // Fetch selected words and their definitions (batched)
  const uniqueSelectedIds = useMemo(() => {
    return Array.from(new Set(selectedWordIdList)).sort((a, b) => a - b);
  }, [selectedWordIdList]);

  const { data: wordsData, isLoading: wordsLoading } = useQuery({
    queryKey: ["selected-words-batch", wordlistId, uniqueSelectedIds.join("-")],
    queryFn: async () => {
      if (uniqueSelectedIds.length === 0) return [];
      const results = await getDefinitionsForWords(
        Number(wordlistId),
        uniqueSelectedIds,
      );
      return results;
    },
    enabled: !!wordlistId && uniqueSelectedIds.length > 0,
    staleTime: 5 * 60 * 1000,
    gcTime: 10 * 60 * 1000,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    retry: 2,
  });

  // Fetch wordlist metadata to obtain languageCode for better assistant responses
  const { data: wordlistMeta } = useQuery({
    queryKey: ["wordlist-meta", wordlistId],
    queryFn: () => getWordlist(Number(wordlistId)),
    enabled: !!wordlistId,
    staleTime: 5 * 60 * 1000,
  });

  // Update selected words when data is fetched
  useEffect(() => {
    if (wordsData) {
      setSelectedWords(wordsData);
    }
  }, [wordsData]);

  // Event callbacks for handling OpenAI server events
  const eventCallbacks: EventCallbacks = useMemo(
    () => ({
      onSessionCreated: (event: any) => {
        console.log("Session created successfully");
        setTranscript("");
        setTranscriptHistory([]);
        telemetrySentRef.current = false;
        telemetryRef.current?.markConversationId(
          event?.session?.conversation_id,
        );
      },

      onResponseCreated: (event: any) => {
        console.log("AI response started");
        setIsAssistantSpeaking(true);
        telemetryRef.current?.recordResponseCreated(event?.response?.id);
        telemetryRef.current?.markConversationId(
          event?.response?.conversation_id,
        );
      },

      onInputAudioBufferSpeechStarted: (event: any) => {
        setIsSpeaking(true);
        telemetryRef.current?.recordUserSpeechStart();
        const finalText = (transcriptRef.current || "")
          .trim()
          .replace(/[\x00-\x1F\x7F-\x9F]/g, "");
        if (finalText) {
          setTranscriptHistory((prev) =>
            prev.length > 0 && prev[0] === finalText
              ? prev
              : [finalText, ...prev],
          );
          setTranscript("");
        }
        setIsAssistantSpeaking(false);
        audioFinishedRef.current = false;
      },

      onInputAudioBufferSpeechStopped: (event: any) => {
        setIsSpeaking(false);
        telemetryRef.current?.recordUserSpeechStop();
      },

      onResponseAudioTranscriptDelta: (event: any) => {
        if (event.delta && typeof event.delta === "string") {
          setTranscript((prev) => prev + event.delta);
        }
      },

      onResponseAudioDelta: () => {
        setIsAssistantSpeaking(true);
      },

      onResponseTextDelta: (event: any) => {
        if (event?.delta && typeof event.delta === "string") {
          setTranscript((prev) => prev + event.delta);
        }
      },

      onResponseAudioTranscriptDone: (event: any) => {
        if (event.transcript && typeof event.transcript === "string") {
          const cleanTranscript = event.transcript
            .replace(/[\x00-\x1F\x7F-\x9F]/g, "")
            .replace(/[\u2018\u2019]/g, "'")
            .replace(/[\u201C\u201D]/g, '"')
            .trim();
          setTranscript(cleanTranscript);
        }
      },

      onResponseDone: (event: any) => {
        console.log("Response completed:", event);
        telemetryRef.current?.recordResponseCompleted(
          event?.response?.id,
          event?.response?.usage,
        );
        telemetryRef.current?.markConversationId(
          event?.response?.conversation_id,
        );
      },

      onResponseAudioDone: (event: any) => {
        console.log("Audio playback finished");
        audioFinishedRef.current = true;
        setIsAssistantSpeaking(false);
      },

      onResponseOutputItemDone: (_event: any) => {},

      onError: (event: any) => {
        console.error("OpenAI API error:", event);
        telemetryRef.current?.recordError();
        setConnectionState({
          status: "error",
          error: event.error?.message || "Unknown error",
        });
      },

      onRateLimitsUpdated: (event: any) => {
        telemetryRef.current?.recordRateLimits(event?.rate_limits);
      },
    }),
    [
      setTranscript,
      setTranscriptHistory,
      setIsSpeaking,
      setIsAssistantSpeaking,
      setConnectionState,
    ],
  );

  // Event handler instance
  const eventHandler = useMemo(
    () => new RealtimeEventHandler(eventCallbacks),
    [eventCallbacks],
  );

  // Handle connection state changes
  const handleConnectionStateChange = useCallback((state: ConnectionState) => {
    setConnectionState(state);
  }, []);

  // Handle server events
  const handleServerEvent = useCallback(
    (event: any) => {
      eventHandler.handleServerEvent(event);
    },
    [eventHandler],
  );

  const ensureAudioPermission = useCallback(async () => {
    const { granted } = await requestRecordingPermissionsAsync();
    return granted;
  }, []);

  const chatConfig = useMemo(() => {
    if (!sessionData || selectedWords.length === 0) {
      return null;
    }

    return {
      sessionData,
      selectedWords,
      wordlistName: wordlistName || "",
      languageCode: wordlistMeta?.languageCode || "en",
      onConnectionStateChange: handleConnectionStateChange,
      onServerEvent: handleServerEvent,
    } as const;
  }, [
    sessionData,
    selectedWords,
    wordlistName,
    wordlistMeta?.languageCode,
    handleConnectionStateChange,
    handleServerEvent,
  ]);

  // Initialize WebRTC chat using the hook
  const realtimeChat = useRealtimeChat(chatConfig);

  useEffect(() => {
    if (chatConfig && !telemetryRef.current) {
      const parsedWordlistId = Number(wordlistId);
      telemetryRef.current = new RealtimeChatTelemetryBuilder({
        wordlistId: Number.isFinite(parsedWordlistId)
          ? parsedWordlistId
          : undefined,
        languageCode: chatConfig.languageCode,
        model: chatConfig.sessionData.webrtcConfig.model,
      });
      telemetrySentRef.current = false;
    }
  }, [chatConfig, wordlistId]);

  useEffect(() => {
    if (telemetryRef.current) {
      telemetryRef.current.updateLanguage(wordlistMeta?.languageCode);
      telemetryRef.current.updateModel(
        chatConfig?.sessionData.webrtcConfig.model,
      );
    }
  }, [wordlistMeta?.languageCode, chatConfig?.sessionData.webrtcConfig.model]);

  // Destructure to stabilize hook dependencies
  const initializeConnection = realtimeChat.initializeConnection;
  const cleanupConnection = realtimeChat.cleanup;
  const toggleMute = realtimeChat.toggleMute;

  // Initialize session
  const initializeSession = useCallback(async () => {
    try {
      setLoading(true);
      setHasTimeout(false);
      setInitError(null);
      setConnectionState({ status: "disconnected" });

      if (selectedWordIdList.length === 0) {
        throw new Error("No words selected for chat practice");
      }

      const permissionGranted = await ensureAudioPermission();
      if (!permissionGranted) {
        throw new Error(
          t(
            "realtimeChat.microphonePermission",
            "Microphone access is required for voice chat.",
          ),
        );
      }

      const timeoutId = setTimeout(() => {
        setHasTimeout(true);
      }, 10000);

      const response = await createChatSession(parseInt(wordlistId));
      clearTimeout(timeoutId);
      setSessionData(response);
    } catch (error: any) {
      console.error("Failed to initialize chat session:", error);
      setInitError(error);
      setConnectionState({
        status: "error",
        error: error.message || "Failed to start chat session",
      });
    } finally {
      setLoading(false);
    }
  }, [wordlistId, selectedWordIdList, ensureAudioPermission, t]);

  const finalizeTelemetry = useCallback(() => {
    if (telemetrySentRef.current) {
      return;
    }
    const builder = telemetryRef.current;
    if (!builder) {
      return;
    }
    telemetrySentRef.current = true;
    const payload = builder.finalize();
    telemetryRef.current = null;
    if (!payload) {
      return;
    }
    sendRealtimeChatTelemetry(payload).catch((error) => {
      console.warn("Failed to send realtime telemetry", error);
    });
  }, []);

  // Initialize connection when configuration is ready
  useEffect(() => {
    if (chatConfig && !realtimeChat.isConnected) {
      initializeConnection();
    }
  }, [chatConfig, initializeConnection, realtimeChat.isConnected]);

  // Initialize session on mount, but only after we have selected words
  useEffect(() => {
    if (selectedWordIdList.length > 0 && !wordsLoading) {
      initializeSession();
    }

    return () => {
      if (cleanupConnection) {
        cleanupConnection();
      }
    };
  }, [initializeSession, selectedWordIdList, wordsLoading, cleanupConnection]);

  useEffect(() => {
    return () => {
      finalizeTelemetry();
    };
  }, [finalizeTelemetry]);

  // UI event handlers
  const handleToggleMute = useCallback(() => {
    toggleMute?.();
  }, [toggleMute]);

  const handleEndCall = useCallback(() => {
    cleanupConnection?.();
    finalizeTelemetry();
    router.back();
  }, [cleanupConnection, finalizeTelemetry, router]);

  const handleRetry = useCallback(() => {
    setTranscript("");
    setTranscriptHistory([]);
    setIsSpeaking(false);
    setHasTimeout(false);
    setInitError(null);
    finalizeTelemetry();
    initializeSession();
  }, [finalizeTelemetry, initializeSession]);

  const handleToggleTranscript = useCallback(() => {
    setShowTranscript((prev) => !prev);
  }, []);

  // Keep a ref in sync with latest transcript for event callbacks
  useEffect(() => {
    transcriptRef.current = transcript;
  }, [transcript]);

  // Auto-scroll transcript area to bottom when new content arrives
  useEffect(() => {
    if (showTranscript) {
      const id = setTimeout(() => {
        transcriptScrollRef.current?.scrollToEnd({ animated: true });
      }, 10);
      return () => clearTimeout(id);
    }
  }, [showTranscript, transcript, transcriptHistory]);

  const getStatusConfig = () => {
    switch (connectionState.status) {
      case "connected":
        return {
          text: t("realtimeChat.connected", "Connected"),
          color: theme.colors.success,
          icon: "wifi" as const,
        };
      case "connecting":
        return {
          text: t("realtimeChat.connecting", "Connecting..."),
          color: theme.colors.semantic.info,
          icon: "sync" as const,
        };
      case "error":
        return {
          text: t("realtimeChat.connectionError", "Connection error"),
          color: theme.colors.error,
          icon: "error-outline" as const,
        };
      default:
        return {
          text: t("realtimeChat.disconnected", "Disconnected"),
          color: theme.colors.text.secondary,
          icon: "wifi-off" as const,
        };
    }
  };

  // Glassmorphic wrapper component (semi-transparent background with blur effect simulation)
  const GlassContainer: React.FC<{
    children: React.ReactNode;
    style?: any;
    intensity?: number;
  }> = ({ children, style }) => {
    return (
      <View style={[styles.glassBase, styles.glassEffect, style]}>
        {children}
      </View>
    );
  };

  const headerTitle = wordlistName || t("realtimeChat.title");

  const isInitialLoading = loading || wordsLoading;

  if (isInitialLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <Stack.Screen options={{ headerShown: false }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <ScreenHeader
          title={headerTitle}
          subtitle={t("realtimeChat.focusWords", {
            count: selectedWords.length,
          })}
          onBackPress={() => router.back()}
        />
        <LoadingWithTimeout
          isLoading={isInitialLoading}
          hasTimeout={hasTimeout}
          error={initError}
          loadingMessage={t("realtimeChat.initializing")}
          timeoutMessage={t("realtimeChat.connectionSlow")}
          onRetry={handleRetry}
          onGoBack={() => router.back()}
        />
      </SafeAreaView>
    );
  }

  const statusConfig = getStatusConfig();

  return (
    <View style={styles.container}>
      {/* Background gradient */}
      <LinearGradient
        colors={
          theme.mode === "light"
            ? ["#FFF8F4", "#FDF6E3", "#FFE8DC"]
            : [
                theme.colors.background.default,
                theme.colors.background.surface,
                theme.colors.background.default,
              ]
        }
        style={StyleSheet.absoluteFill}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      />

      <SafeAreaView style={styles.safeArea}>
        <Stack.Screen options={{ headerShown: false }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <ScreenHeader
          title={headerTitle}
          subtitle={t("realtimeChat.focusWords", {
            count: selectedWords.length,
          })}
          onBackPress={() => router.back()}
        />

        {/* Main Chat Area */}
        <Animated.View style={[styles.chatArea, { opacity: fadeAnim }]}>
          {/* Floating Status Pill */}
          <GlassContainer style={styles.statusPill} intensity={70}>
            <View
              style={[
                styles.statusDot,
                { backgroundColor: statusConfig.color },
              ]}
            />
            <MaterialIcons
              name={statusConfig.icon}
              size={14}
              color={statusConfig.color}
            />
            <Text style={[styles.statusText, { color: statusConfig.color }]}>
              {statusConfig.text}
            </Text>
          </GlassContainer>

          {/* Central Orb Container */}
          <View style={styles.orbSection}>
            {/* Outer glow ring */}
            <Animated.View
              style={[
                styles.orbGlow,
                {
                  opacity: orbGlowAnim,
                  transform: [{ scale: pulseAnim }],
                },
              ]}
            />

            {/* Main Orb */}
            <Animated.View
              style={[
                styles.orbContainer,
                {
                  transform: [{ scale: pulseAnim }],
                },
              ]}
            >
              <LinearGradient
                colors={
                  realtimeChat.isConnected
                    ? [theme.colors.primary, "#FF9966", "#FFB88C"]
                    : [
                        theme.colors.text.secondary,
                        theme.colors.text.tertiary,
                        theme.colors.text.secondary,
                      ]
                }
                style={styles.orbGradient}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
              >
                {/* Lottie Wave inside orb */}
                {(isAssistantSpeaking || isSpeaking) && (
                  <View style={styles.orbWaveContainer}>
                    <LottieWave
                      color="#FFFFFF"
                      height={responsive.getValueForSize(80, 90, 100, 110)}
                      speed={1.2}
                      active={isAssistantSpeaking || isSpeaking}
                    />
                  </View>
                )}

                {/* Icon when not actively speaking */}
                {!isAssistantSpeaking && !isSpeaking && (
                  <MaterialIcons
                    name={realtimeChat.isMuted ? "mic-off" : "mic"}
                    size={responsive.getValueForSize(48, 56, 64, 72)}
                    color="rgba(255, 255, 255, 0.9)"
                  />
                )}
              </LinearGradient>
            </Animated.View>

            {/* Status Text Below Orb */}
            <Text style={styles.mainStatusText}>
              {isSpeaking
                ? t("realtimeChat.listening", "Listening...")
                : isAssistantSpeaking
                  ? t("realtimeChat.aiSpeaking", "AI is speaking...")
                  : realtimeChat.isConnected
                    ? t("realtimeChat.speakNow", "Speak to practice")
                    : t("realtimeChat.connecting", "Connecting...")}
            </Text>

            {/* Word chips */}
            <View style={styles.wordChips}>
              {selectedWords.slice(0, 3).map((word, idx) => (
                <View key={`word-${idx}-${word.name}`} style={styles.wordChip}>
                  <Text style={styles.wordChipText}>{word.name}</Text>
                </View>
              ))}
              {selectedWords.length > 3 && (
                <View style={[styles.wordChip, styles.wordChipMore]}>
                  <Text style={styles.wordChipText}>
                    +{selectedWords.length - 3}
                  </Text>
                </View>
              )}
            </View>
          </View>

          {/* Collapsible Transcript Panel */}
          {showTranscript && (transcript || transcriptHistory.length > 0) && (
            <GlassContainer style={styles.transcriptPanel} intensity={50}>
              <ScrollView
                ref={transcriptScrollRef}
                style={styles.transcriptScroll}
                contentContainerStyle={styles.transcriptScrollContent}
                keyboardShouldPersistTaps="handled"
                showsVerticalScrollIndicator={false}
              >
                {/* Previous responses (oldest first) */}
                {transcriptHistory.length > 0 && (
                  <View style={styles.transcriptHistoryContainer}>
                    {transcriptHistory
                      .slice()
                      .reverse()
                      .map((item, idx) => (
                        <View key={`hist-${idx}`} style={styles.historyBubble}>
                          <Text style={styles.transcriptHistoryItem}>
                            {item}
                          </Text>
                        </View>
                      ))}
                  </View>
                )}

                {/* Live transcript at the bottom */}
                {!!transcript && (
                  <View style={styles.liveBubble}>
                    <View style={styles.liveIndicator}>
                      <View style={styles.liveDot} />
                      <Text style={styles.liveLabel}>
                        {t("realtimeChat.live", "Live")}
                      </Text>
                    </View>
                    <Text style={styles.transcriptText}>{transcript}</Text>
                  </View>
                )}
              </ScrollView>
            </GlassContainer>
          )}

          {/* Error State */}
          {connectionState.status === "error" && (
            <GlassContainer style={styles.errorContainer} intensity={70}>
              <MaterialIcons
                name="error-outline"
                size={responsive.getValueForSize(32, 36, 40, 44)}
                color={theme.colors.error}
              />
              <Text style={styles.errorText}>
                {t("realtimeChat.genericError", "Something went wrong")}
              </Text>
              {!!connectionState.error && (
                <Text style={styles.errorSubtext}>{connectionState.error}</Text>
              )}
              <TouchableOpacity
                style={styles.retryButton}
                onPress={handleRetry}
                activeOpacity={0.8}
              >
                <LinearGradient
                  colors={[theme.colors.primary, "#FF9966"]}
                  style={styles.retryButtonGradient}
                  start={{ x: 0, y: 0 }}
                  end={{ x: 1, y: 1 }}
                >
                  <MaterialIcons
                    name="refresh"
                    size={18}
                    color={theme.colors.text.inverse}
                  />
                  <Text style={styles.retryButtonText}>
                    {t("common.retry", "Retry")}
                  </Text>
                </LinearGradient>
              </TouchableOpacity>
            </GlassContainer>
          )}
        </Animated.View>

        {/* Bottom Controls - Glassmorphic */}
        <GlassContainer style={styles.controls} intensity={80}>
          {/* Transcript Toggle */}
          <TouchableOpacity
            style={[
              styles.controlButton,
              showTranscript && styles.controlButtonActive,
            ]}
            onPress={handleToggleTranscript}
            activeOpacity={0.7}
            accessibilityRole="button"
            accessibilityLabel={
              showTranscript
                ? t("realtimeChat.hideTranscript", "Hide transcript")
                : t("realtimeChat.showTranscript", "Show transcript")
            }
          >
            <MaterialIcons
              name={showTranscript ? "subtitles" : "subtitles-off"}
              size={responsive.getValueForSize(22, 24, 26, 28)}
              color={
                showTranscript
                  ? theme.colors.primary
                  : theme.colors.text.secondary
              }
            />
          </TouchableOpacity>

          {/* Mute Toggle */}
          <TouchableOpacity
            style={[
              styles.controlButton,
              realtimeChat.isMuted && styles.controlButtonMuted,
            ]}
            onPress={handleToggleMute}
            disabled={!realtimeChat.isConnected}
            activeOpacity={0.7}
            accessibilityRole="button"
            accessibilityState={{ disabled: !realtimeChat.isConnected }}
            accessibilityLabel={
              realtimeChat.isMuted
                ? t("realtimeChat.unmute", "Unmute")
                : t("realtimeChat.mute", "Mute")
            }
          >
            <MaterialIcons
              name={realtimeChat.isMuted ? "mic-off" : "mic"}
              size={responsive.getValueForSize(22, 24, 26, 28)}
              color={
                realtimeChat.isMuted
                  ? theme.colors.text.inverse
                  : realtimeChat.isConnected
                    ? theme.colors.success
                    : theme.colors.text.disabled
              }
            />
          </TouchableOpacity>

          {/* End Call */}
          <TouchableOpacity
            style={styles.endCallButton}
            onPress={handleEndCall}
            activeOpacity={0.8}
            accessibilityRole="button"
            accessibilityLabel={t("realtimeChat.endCall", "End call")}
          >
            <LinearGradient
              colors={[theme.colors.error, "#FF6B6B"]}
              style={styles.endCallGradient}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
            >
              <MaterialIcons
                name="call-end"
                size={responsive.getValueForSize(26, 28, 30, 32)}
                color={theme.colors.text.inverse}
              />
            </LinearGradient>
          </TouchableOpacity>
        </GlassContainer>
      </SafeAreaView>
    </View>
  );
}

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) => {
  const bodyFontSize = responsive.getScaledFont("body");
  const lineHeightBody = Math.round(bodyFontSize * 1.4);

  return StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    safeArea: {
      flex: 1,
    },
    glassBase: {
      overflow: "hidden",
    },
    glassEffect: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.85)"
          : "rgba(30, 30, 30, 0.85)",
    },
    chatArea: {
      flex: 1,
      justifyContent: "flex-start",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingTop: responsive.spacing.vertical,
    },
    statusPill: {
      flexDirection: "row",
      alignItems: "center",
      gap: 6,
      paddingHorizontal: responsive.getValueForSize(14, 16, 18, 20),
      paddingVertical: responsive.getValueForSize(8, 10, 10, 12),
      borderRadius: theme.borderRadius.full,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.6)"
          : "rgba(255, 255, 255, 0.1)",
      marginBottom: responsive.spacing.vertical,
    },
    statusDot: {
      width: 8,
      height: 8,
      borderRadius: 4,
    },
    statusText: {
      fontSize: responsive.getScaledFont("label"),
      fontWeight: "600",
    },
    orbSection: {
      alignItems: "center",
      marginVertical: responsive.getValueForSize(20, 28, 36, 44),
    },
    orbGlow: {
      position: "absolute",
      width: responsive.getValueForSize(200, 240, 280, 320),
      height: responsive.getValueForSize(200, 240, 280, 320),
      borderRadius: responsive.getValueForSize(100, 120, 140, 160),
      backgroundColor: theme.colors.primary,
    },
    orbContainer: {
      width: responsive.getValueForSize(160, 190, 220, 250),
      height: responsive.getValueForSize(160, 190, 220, 250),
      borderRadius: responsive.getValueForSize(80, 95, 110, 125),
      overflow: "hidden",
      shadowColor: theme.colors.primary,
      shadowOffset: { width: 0, height: 12 },
      shadowOpacity: 0.4,
      shadowRadius: 24,
      elevation: 16,
    },
    orbGradient: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
    },
    orbWaveContainer: {
      width: "100%",
      alignItems: "center",
      justifyContent: "center",
    },
    mainStatusText: {
      marginTop: responsive.getValueForSize(20, 24, 28, 32),
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    wordChips: {
      flexDirection: "row",
      flexWrap: "wrap",
      justifyContent: "center",
      gap: 8,
      marginTop: responsive.getValueForSize(12, 16, 20, 24),
      paddingHorizontal: responsive.spacing.horizontal,
    },
    wordChip: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.12)"
          : "rgba(255, 123, 84, 0.2)",
      paddingHorizontal: responsive.getValueForSize(12, 14, 16, 18),
      paddingVertical: responsive.getValueForSize(6, 8, 8, 10),
      borderRadius: theme.borderRadius.full,
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.2)"
          : "rgba(255, 123, 84, 0.3)",
    },
    wordChipMore: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : "rgba(255, 255, 255, 0.1)",
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.08)"
          : "rgba(255, 255, 255, 0.15)",
    },
    wordChipText: {
      fontSize: responsive.getScaledFont("label"),
      fontWeight: "600",
      color: theme.colors.primary,
    },
    transcriptPanel: {
      width: "100%",
      maxHeight: responsive.getValueForSize(180, 220, 260, 300),
      borderRadius: responsive.getValueForSize(16, 18, 20, 22),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.6)"
          : "rgba(255, 255, 255, 0.1)",
      marginTop: responsive.spacing.vertical,
    },
    transcriptScroll: {
      flex: 1,
    },
    transcriptScrollContent: {
      padding: responsive.getValueForSize(12, 14, 16, 18),
      gap: responsive.getValueForSize(8, 10, 12, 14),
    },
    transcriptHistoryContainer: {
      gap: responsive.getValueForSize(8, 10, 12, 14),
    },
    historyBubble: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.04)"
          : "rgba(255, 255, 255, 0.06)",
      borderRadius: responsive.getValueForSize(12, 14, 16, 18),
      padding: responsive.getValueForSize(10, 12, 14, 16),
    },
    transcriptHistoryItem: {
      fontSize: bodyFontSize,
      color: theme.colors.text.secondary,
      lineHeight: lineHeightBody,
      ...(Platform.OS === "web"
        ? ({
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            overflowWrap: "anywhere",
          } as any)
        : {}),
    },
    liveBubble: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.08)"
          : "rgba(255, 123, 84, 0.12)",
      borderRadius: responsive.getValueForSize(12, 14, 16, 18),
      padding: responsive.getValueForSize(10, 12, 14, 16),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.15)"
          : "rgba(255, 123, 84, 0.2)",
    },
    liveIndicator: {
      flexDirection: "row",
      alignItems: "center",
      gap: 6,
      marginBottom: 6,
    },
    liveDot: {
      width: 6,
      height: 6,
      borderRadius: 3,
      backgroundColor: theme.colors.primary,
    },
    liveLabel: {
      fontSize: responsive.getScaledFont("micro"),
      fontWeight: "700",
      color: theme.colors.primary,
      textTransform: "uppercase",
      letterSpacing: 0.5,
    },
    transcriptText: {
      fontSize: bodyFontSize,
      color: theme.colors.text.primary,
      lineHeight: lineHeightBody,
      ...(Platform.OS === "web"
        ? ({
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            overflowWrap: "anywhere",
          } as any)
        : {}),
    },
    errorContainer: {
      alignItems: "center",
      gap: responsive.spacing.elementSpacing,
      padding: responsive.spacing.formPadding,
      marginTop: responsive.spacing.vertical,
      borderRadius: responsive.getValueForSize(16, 18, 20, 22),
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(255, 255, 255, 0.6)"
          : "rgba(255, 255, 255, 0.1)",
    },
    errorText: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.error,
      textAlign: "center",
    },
    errorSubtext: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    retryButton: {
      borderRadius: theme.borderRadius.full,
      overflow: "hidden",
      marginTop: responsive.spacing.elementSpacing / 2,
    },
    retryButtonGradient: {
      flexDirection: "row",
      alignItems: "center",
      gap: 8,
      paddingHorizontal: responsive.getValueForSize(20, 24, 28, 32),
      paddingVertical: responsive.getValueForSize(12, 14, 14, 16),
    },
    retryButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "700",
      color: theme.colors.text.inverse,
    },
    controls: {
      flexDirection: "row",
      justifyContent: "center",
      alignItems: "center",
      gap: responsive.getValueForSize(16, 20, 24, 28),
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.getValueForSize(16, 20, 24, 28),
      borderTopWidth: 1,
      borderTopColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : "rgba(255, 255, 255, 0.1)",
      borderTopLeftRadius: responsive.getValueForSize(24, 28, 32, 36),
      borderTopRightRadius: responsive.getValueForSize(24, 28, 32, 36),
    },
    controlButton: {
      width: responsive.getValueForSize(52, 56, 60, 64),
      height: responsive.getValueForSize(52, 56, 60, 64),
      borderRadius: responsive.getValueForSize(26, 28, 30, 32),
      backgroundColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.06)"
          : "rgba(255, 255, 255, 0.1)",
      justifyContent: "center",
      alignItems: "center",
      borderWidth: 1,
      borderColor:
        theme.mode === "light"
          ? "rgba(0, 0, 0, 0.08)"
          : "rgba(255, 255, 255, 0.15)",
    },
    controlButtonActive: {
      backgroundColor:
        theme.mode === "light"
          ? "rgba(255, 123, 84, 0.15)"
          : "rgba(255, 123, 84, 0.2)",
      borderColor: theme.colors.primary,
    },
    controlButtonMuted: {
      backgroundColor: theme.colors.error,
      borderColor: theme.colors.error,
    },
    endCallButton: {
      width: responsive.getValueForSize(64, 68, 72, 76),
      height: responsive.getValueForSize(64, 68, 72, 76),
      borderRadius: responsive.getValueForSize(32, 34, 36, 38),
      overflow: "hidden",
      shadowColor: theme.colors.error,
      shadowOffset: { width: 0, height: 6 },
      shadowOpacity: 0.35,
      shadowRadius: 12,
      elevation: 10,
    },
    endCallGradient: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
    },
  });
};

// Cross-platform wave using Lottie on native
type LottieWaveProps = {
  color: string;
  height: number;
  speed?: number;
  active?: boolean;
};
const LottieWave: React.FC<LottieWaveProps> = ({
  color,
  height,
  speed = 1.2,
  active = true,
}) => {
  const source = require("../assets/animations/voice-wave.json");
  const ref = useRef<LottieView>(null);

  useEffect(() => {
    if (!ref.current) return;
    if (active) (ref.current as any).play?.();
    else (ref.current as any).pause?.();
  }, [active]);

  return (
    <LottieView
      ref={ref}
      source={source}
      loop
      autoPlay={false}
      speed={speed}
      style={{
        width: "100%",
        height,
        alignSelf: "stretch",
      }}
      colorFilters={[{ keypath: "**", color }]}
      resizeMode="cover"
    />
  );
};

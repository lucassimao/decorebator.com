import { LoadingWithTimeout } from "@/components/LoadingWithTimeout";
import { useTheme } from "@/contexts/ThemeContext";
import { ConnectionState, useRealtimeChat } from "@/hooks/useRealtimeChat";
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
  Platform,
  SafeAreaView,
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

const RealtimeChatScreen: React.FC = () => {
  const { wordlistId, wordlistName, selectedWordIds } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
    selectedWordIds: string;
  }>();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

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
  // Track whether assistant audio finished; we now finalize on user's next speech start
  const audioFinishedRef = useRef<boolean>(false);
  const telemetryRef = useRef<RealtimeChatTelemetryBuilder | null>(null);
  const telemetrySentRef = useRef(false);

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
        // Clear any previous transcript when new session starts
        setTranscript("");
        setTranscriptHistory([]);
        telemetrySentRef.current = false;
        telemetryRef.current?.markConversationId(event?.session?.conversation_id);
      },

      onResponseCreated: (event: any) => {
        console.log("🤖 AI response started");
        // Assistant is about to speak; if transcript is hidden, show the wave
        setIsAssistantSpeaking(true);
        // Keep current transcript until the user starts speaking
        telemetryRef.current?.recordResponseCreated(event?.response?.id);
        telemetryRef.current?.markConversationId(
          event?.response?.conversation_id,
        );
      },

      onInputAudioBufferSpeechStarted: (event: any) => {
        setIsSpeaking(true);
        telemetryRef.current?.recordUserSpeechStart();
        // When user starts speaking, finalize the previous assistant utterance (if present)
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
        // Pause wave while the user is speaking
        setIsAssistantSpeaking(false);
        // Reset flag for next turn
        audioFinishedRef.current = false;
      },

      onInputAudioBufferSpeechStopped: (event: any) => {
        setIsSpeaking(false);
        telemetryRef.current?.recordUserSpeechStop();
      },

      onResponseAudioTranscriptDelta: (event: any) => {
        console.log(
          "🔍 Transcript delta received:",
          JSON.stringify(event.delta),
        );
        if (event.delta && typeof event.delta === "string") {
          setTranscript((prev) => prev + event.delta);
        }
      },

      // Audio chunks streaming from assistant
      onResponseAudioDelta: () => {
        setIsAssistantSpeaking(true);
      },

      // Fallback if text output is enabled in the future
      onResponseTextDelta: (event: any) => {
        if (event?.delta && typeof event.delta === "string") {
          setTranscript((prev) => prev + event.delta);
        }
      },

      onResponseAudioTranscriptDone: (event: any) => {
        console.log("🔍 FULL EVENT:", JSON.stringify(event, null, 2));
        console.log("🔍 Transcript type:", typeof event.transcript);
        console.log("🔍 Transcript length:", event.transcript?.length);
        console.log("🔍 Transcript raw:", event.transcript);
        // console.log("🔍 Transcript char codes:", event.transcript ? Array.from(event.transcript).map(c => c.charCodeAt(0)).join(',') : 'null');

        // Use the complete transcript from the done event
        if (event.transcript && typeof event.transcript === "string") {
          // Filter out non-printable characters and normalize Unicode quotes
          const cleanTranscript = event.transcript
            .replace(/[\x00-\x1F\x7F-\x9F]/g, "") // Remove control characters
            .replace(/[\u2018\u2019]/g, "'") // Replace smart quotes with regular apostrophes
            .replace(/[\u201C\u201D]/g, '"') // Replace smart quotes with regular quotes
            .trim();

          console.log("🔍 Clean transcript:", cleanTranscript);
          // Keep as the live transcript only; we'll move it when the user starts speaking
          setTranscript(cleanTranscript);
        }
      },

      onResponseDone: (event: any) => {
        console.log("Response completed:", event);
        // Keep transcript visible until audio playback finishes
        telemetryRef.current?.recordResponseCompleted(
          event?.response?.id,
          event?.response?.usage,
        );
        telemetryRef.current?.markConversationId(
          event?.response?.conversation_id,
        );
      },

      // Audio finished (if event is emitted): now move live transcript into history and clear the live slot
      onResponseAudioDone: (event: any) => {
        console.log("🎵 Audio playback finished");
        // Mark that assistant finished speaking; we will finalize on the user's next speech start
        audioFinishedRef.current = true;
        // Stop the animation when audio playback is actually done
        setIsAssistantSpeaking(false);
      },

      // Do not toggle speaking state on output item done to avoid early stop
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
      telemetryRef.current.updateModel(chatConfig?.sessionData.webrtcConfig.model);
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

      // Check if we have selected words
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

      // Set timeout for session initialization (10 seconds)
      const timeoutId = setTimeout(() => {
        setHasTimeout(true);
      }, 10000);

      // Get session token from backend (iOS will prompt for mic on getUserMedia)
      const response = await createChatSession(parseInt(wordlistId));
      clearTimeout(timeoutId);
      setSessionData(response);

      // The WebRTC connection will be initialized by the hook
      // when sessionData is set
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

    // Cleanup on unmount
    return () => {
      if (cleanupConnection) {
        cleanupConnection();
      }
      // no timers to clear in current turn-taking strategy
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
      // slight delay to ensure layout pass completes
      const id = setTimeout(() => {
        transcriptScrollRef.current?.scrollToEnd({ animated: true });
      }, 10);
      return () => clearTimeout(id);
    }
  }, [showTranscript, transcript, transcriptHistory]);

  const getStatusText = () => {
    switch (connectionState.status) {
      case "connecting":
        return t("realtimeChat.connecting", "Connecting...");
      case "connected":
        return t("realtimeChat.connected", "Connected - Start speaking!");
      case "error":
        return t("realtimeChat.connectionError", "Connection error");
      default:
        return t("realtimeChat.disconnected", "Disconnected");
    }
  };

  // Uniform status color/background using theme tokens
  const getStatusColors = () => {
    switch (connectionState.status) {
      case "connected":
        return {
          fg: theme.colors.success,
          bg: theme.colors.state.correctBackground,
          border: theme.colors.border.light,
        } as const;
      case "connecting":
        return {
          fg: theme.colors.semantic.info,
          bg: theme.colors.state.infoBackground,
          border: theme.colors.border.light,
        } as const;
      case "error":
        return {
          fg: theme.colors.error,
          bg: theme.colors.state.incorrectBackground,
          border: theme.colors.border.light,
        } as const;
      default:
        return {
          fg: theme.colors.text.secondary,
          bg: theme.colors.background.elevated,
          border: theme.colors.ui.divider,
        } as const;
    }
  };

  const headerTitle = `${t("realtimeChat.title")} • ${wordlistName}`;

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

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ headerShown: false }} />
      <StatusBar
        barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
      />
      <ScreenHeader
        title={headerTitle}
        subtitle={t("realtimeChat.focusWords", { count: selectedWords.length })}
        onBackPress={() => router.back()}
      />

      {/* Main Chat Area */}
      <View style={styles.chatArea}>
        {(() => {
          const status = getStatusColors();
          return (
            <View
              style={[
                styles.statusIndicator,
                {
                  backgroundColor: status.bg,
                  borderColor: status.border,
                },
              ]}
            >
              <View
                style={[styles.statusDot, { backgroundColor: status.fg }]}
              />
              <Text style={[styles.statusText, { color: status.fg }]}>
                {getStatusText()}
              </Text>
            </View>
          );
        })()}

        <View style={styles.conversationIndicator}>
          <MaterialIcons
            name="mic"
            size={responsive.getValueForSize(48, 52, 56, 60)}
            color={
              realtimeChat.isConnected && !realtimeChat.isMuted
                ? theme.colors.success
                : theme.colors.text.secondary
            }
          />
          <Text style={styles.conversationText}>
            {isSpeaking
              ? t("realtimeChat.listening")
              : realtimeChat.isConnected
                ? t("realtimeChat.speakNow")
                : t("realtimeChat.connecting")}
          </Text>
          <Text style={styles.wordsCountText}>
            {selectedWords.length > 0
              ? t("realtimeChat.focusWords", {
                  count: selectedWords.length,
                })
              : t("realtimeChat.wordsAvailable", {
                  count: 0,
                })}
          </Text>

          {/* Transcript & History (toggleable, scrollable) */}
          {showTranscript && (
            <View style={styles.transcriptPanel}>
              <ScrollView
                ref={transcriptScrollRef}
                style={styles.transcriptScroll}
                contentContainerStyle={styles.transcriptScrollContent}
                keyboardShouldPersistTaps="handled"
              >
                {/* Previous responses (oldest first) */}
                {transcriptHistory.length > 0 && (
                  <View style={styles.transcriptHistoryContainer}>
                    <Text style={styles.transcriptHistoryLabel}>
                      {t("realtimeChat.previousResponses")}
                    </Text>
                    {transcriptHistory
                      .slice() // clone
                      .reverse() // show oldest → newest
                      .map((item, idx) => (
                        <Text
                          key={`hist-${idx}`}
                          style={styles.transcriptHistoryItem}
                        >
                          {item}
                        </Text>
                      ))}
                  </View>
                )}

                {/* Live transcript at the bottom */}
                {!!transcript && (
                  <View style={styles.transcriptContainer}>
                    <Text style={styles.transcriptLabel}>
                      {t("realtimeChat.transcript")}
                    </Text>
                    <Text style={styles.transcriptText}>{transcript}</Text>
                  </View>
                )}
              </ScrollView>
            </View>
          )}

          {/* Audio wave when transcript is hidden and assistant is speaking */}
          {!showTranscript && isAssistantSpeaking && (
            <View style={styles.lottieContainer}>
              <LottieWave
                color={theme.colors.primary}
                height={responsive.getValueForSize(60, 70, 80, 90)}
                speed={1}
                active={isAssistantSpeaking}
              />
            </View>
          )}
        </View>

        {connectionState.status === "error" && (
          <View style={styles.errorContainer}>
            <MaterialIcons
              name="error-outline"
              size={responsive.getValueForSize(24, 26, 28, 30)}
              color={theme.colors.error}
            />
            <Text style={styles.errorText}>
              {t("realtimeChat.genericError")}
            </Text>
            {!!connectionState.error && (
              <Text style={styles.errorSubtext}>{connectionState.error}</Text>
            )}
            <TouchableOpacity style={styles.retryButton} onPress={handleRetry}>
              <Text style={styles.retryButtonText}>{t("common.retry")}</Text>
            </TouchableOpacity>
          </View>
        )}
      </View>

      {/* Controls */}
      <View style={styles.controls}>
        <TouchableOpacity
          style={[
            styles.controlButton,
            showTranscript && styles.controlButtonSelected,
          ]}
          onPress={handleToggleTranscript}
          accessibilityRole="button"
          accessibilityLabel={
            showTranscript
              ? t("realtimeChat.hideTranscript")
              : t("realtimeChat.showTranscript")
          }
        >
          <MaterialIcons
            name={showTranscript ? "subtitles" : "subtitles-off"}
            size={responsive.getValueForSize(24, 26, 28, 30)}
            color={
              showTranscript
                ? theme.colors.primary
                : theme.colors.text.secondary
            }
          />
        </TouchableOpacity>
        <TouchableOpacity
          style={[
            styles.controlButton,
            realtimeChat.isMuted && styles.controlButtonActive,
          ]}
          onPress={handleToggleMute}
          disabled={!realtimeChat.isConnected}
          accessibilityRole="button"
          accessibilityState={{ disabled: !realtimeChat.isConnected }}
          accessibilityLabel={
            realtimeChat.isMuted
              ? t("realtimeChat.unmute")
              : t("realtimeChat.mute")
          }
        >
          <MaterialIcons
            name={realtimeChat.isMuted ? "mic-off" : "mic"}
            size={responsive.getValueForSize(24, 26, 28, 30)}
            color={
              realtimeChat.isMuted
                ? theme.colors.text.inverse
                : realtimeChat.isConnected
                  ? theme.colors.success
                  : theme.colors.text.secondary
            }
          />
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.endCallButton]}
          onPress={handleEndCall}
          accessibilityRole="button"
          accessibilityLabel={t("realtimeChat.endCall")}
        >
          <MaterialIcons
            name="call-end"
            size={responsive.getValueForSize(28, 30, 32, 34)}
            color={theme.colors.text.inverse}
          />
        </TouchableOpacity>
      </View>
    </SafeAreaView>
  );
};

const createStyles = (
  theme: ReturnType<typeof useTheme>["theme"],
  responsive: ReturnType<typeof useTheme>["responsive"],
) => {
  // Derive a comfortable line height for multi-line captions
  const bodyFontSize = responsive.getScaledFont("body");
  const lineHeightBody = Math.round(bodyFontSize * 1.4);

  return StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.background.default,
    },
    // header removed in favor of ScreenHeader
    loadingContainer: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      gap: responsive.spacing.elementSpacing,
    },
    loadingText: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    chatArea: {
      flex: 1,
      justifyContent: "center",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
    },
    statusIndicator: {
      flexDirection: "row",
      alignItems: "center",
      gap: responsive.spacing.elementSpacing / 2,
      marginBottom: responsive.spacing.vertical,
      paddingHorizontal: responsive.getValueForSize(10, 12, 14, 16),
      paddingVertical: responsive.getValueForSize(6, 8, 8, 10),
      borderRadius: theme.borderRadius.full,
      borderWidth: 1,
      alignSelf: "center",
    },
    statusDot: {
      width: responsive.getValueForSize(8, 10, 12, 14),
      height: responsive.getValueForSize(8, 10, 12, 14),
      borderRadius: responsive.getValueForSize(4, 5, 6, 7),
    },
    statusText: {
      fontSize: responsive.getScaledFont("label"),
      fontWeight: "600",
    },
    conversationIndicator: {
      alignItems: "center",
      gap: responsive.spacing.elementSpacing,
      alignSelf: "stretch",
      width: "100%",
    },
    conversationText: {
      fontSize: responsive.getScaledFont("headline"),
      fontWeight: "600",
      color: theme.colors.text.primary,
      textAlign: "center",
    },
    wordsCountText: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    transcriptContainer: {
      marginTop: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      backgroundColor: theme.colors.background.elevated,
      borderRadius: theme.borderRadius.md,
      alignSelf: "stretch",
      // Ensure text can wrap within the container on all platforms
      flexDirection: "column",
      alignItems: "flex-start",
      minWidth: 0,
    },
    transcriptPanel: {
      marginTop: responsive.spacing.vertical,
      width: "100%",
      maxHeight: responsive.getValueForSize(220, 280, 340, 400),
      alignSelf: "stretch",
      minWidth: 0,
    },
    transcriptScroll: {
      width: "100%",
      borderRadius: theme.borderRadius.md,
      backgroundColor: theme.colors.background.surface,
      overflow: "hidden",
      minWidth: 0,
    },
    transcriptScrollContent: {
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      flexGrow: 1,
      alignItems: "stretch",
      minWidth: 0,
    },
    transcriptLabel: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      fontWeight: "600",
      marginBottom: responsive.spacing.elementSpacing / 2,
    },
    transcriptText: {
      fontSize: bodyFontSize,
      color: theme.colors.text.primary,
      lineHeight: lineHeightBody,
      // Use normal style for readability
      fontStyle: "normal",
      textAlign: "left",
      width: "100%",
      flexShrink: 1,
      flexWrap: "wrap",
      minWidth: 0,
      ...(Platform.OS === "web"
        ? ({
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            overflowWrap: "anywhere",
          } as any)
        : {}),
    },
    transcriptHistoryContainer: {
      marginTop: responsive.spacing.elementSpacing,
      gap: responsive.spacing.elementSpacing / 2,
      alignSelf: "stretch",
    },
    transcriptHistoryLabel: {
      fontSize: responsive.getScaledFont("label"),
      color: theme.colors.text.secondary,
      fontWeight: "600",
      marginBottom: responsive.spacing.elementSpacing / 4,
    },
    transcriptHistoryItem: {
      fontSize: bodyFontSize,
      color: theme.colors.text.secondary,
      lineHeight: lineHeightBody,
      textAlign: "left",
      width: "100%",
      flexShrink: 1,
      flexWrap: "wrap",
      minWidth: 0,
      ...(Platform.OS === "web"
        ? ({
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
            overflowWrap: "anywhere",
          } as any)
        : {}),
    },
    lottieContainer: {
      marginTop: responsive.spacing.vertical,
      alignItems: "stretch",
      justifyContent: "center",
      alignSelf: "stretch",
      width: "100%",
    },
    errorContainer: {
      alignItems: "center",
      gap: responsive.spacing.elementSpacing,
      paddingHorizontal: responsive.spacing.horizontal,
      marginTop: responsive.spacing.vertical,
    },
    errorText: {
      fontSize: responsive.getScaledFont("body"),
      color: theme.colors.error,
      textAlign: "center",
    },
    errorSubtext: {
      fontSize: responsive.getScaledFont("micro"),
      color: theme.colors.text.secondary,
      textAlign: "center",
    },
    retryButton: {
      backgroundColor: theme.colors.primary,
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.elementSpacing,
      borderRadius: theme.borderRadius.md,
    },
    retryButtonText: {
      fontSize: responsive.getScaledFont("body"),
      fontWeight: "600",
      color: theme.colors.text.inverse,
    },
    controls: {
      flexDirection: "row",
      justifyContent: "space-around",
      alignItems: "center",
      paddingHorizontal: responsive.spacing.horizontal,
      paddingVertical: responsive.spacing.vertical,
      backgroundColor: theme.colors.background.surface,
      borderTopWidth: 1,
      borderTopColor: theme.colors.ui.divider,
    },
    controlButton: {
      width: responsive.getValueForSize(56, 60, 64, 68),
      height: responsive.getValueForSize(56, 60, 64, 68),
      borderRadius: responsive.getValueForSize(28, 30, 32, 34),
      backgroundColor: theme.colors.background.elevated,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.sm,
    },
    controlButtonSelected: {
      borderWidth: 2,
      borderColor: theme.colors.primary,
    },
    controlButtonActive: {
      backgroundColor: theme.colors.error,
    },
    endCallButton: {
      width: responsive.getValueForSize(64, 68, 72, 76),
      height: responsive.getValueForSize(64, 68, 72, 76),
      borderRadius: responsive.getValueForSize(32, 34, 36, 38),
      backgroundColor: theme.colors.error,
      justifyContent: "center",
      alignItems: "center",
      ...theme.shadows.md,
    },
  });
};

export default RealtimeChatScreen;

// Cross-platform wave using Lottie on native and ActivityIndicator on web
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
    // play or pause based on active flag
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

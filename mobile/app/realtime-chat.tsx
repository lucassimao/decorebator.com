import React, {
  useEffect,
  useState,
  useCallback,
  useMemo,
  useRef,
} from "react";
import { useLocalSearchParams, useRouter, Stack } from "expo-router";
import {
  View,
  StyleSheet,
  Text,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  SafeAreaView,
  StatusBar,
  Platform,
} from "react-native";
import LottieView from "lottie-react-native";
import { MaterialIcons } from "@expo/vector-icons";
import { useTranslation } from "react-i18next";
import { useTheme } from "@/contexts/ThemeContext";
import { createChatSession, ChatSessionData } from "../api/wordlists";
import { useRealtimeChat, ConnectionState } from "@/hooks/useRealtimeChat";
import {
  RealtimeEventHandler,
  EventCallbacks,
} from "@/utils/realtimeEventHandler";

const RealtimeChatScreen: React.FC = () => {
  const { wordlistId, wordlistName } = useLocalSearchParams<{
    wordlistId: string;
    wordlistName: string;
  }>();
  const router = useRouter();
  const { t } = useTranslation();
  const { theme, responsive } = useTheme();
  const styles = createStyles(theme, responsive);

  // State
  const [sessionData, setSessionData] = useState<ChatSessionData | null>(null);
  const [loading, setLoading] = useState(true);
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

  // Event callbacks for handling OpenAI server events
  const eventCallbacks: EventCallbacks = useMemo(
    () => ({
      onSessionCreated: (event: any) => {
        console.log("Session created successfully");
        // Clear any previous transcript when new session starts
        setTranscript("");
        setTranscriptHistory([]);
      },

      onResponseCreated: (event: any) => {
        console.log("🤖 AI response started");
        // Assistant is about to speak; if transcript is hidden, show the wave
        setIsAssistantSpeaking(true);
        // Keep current transcript until the user starts speaking
      },

      onInputAudioBufferSpeechStarted: (event: any) => {
        setIsSpeaking(true);
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
        setConnectionState({
          status: "error",
          error: event.error?.message || "Unknown error",
        });
      },
    }),
    [setTranscript, setTranscriptHistory, setIsSpeaking, setConnectionState],
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

  // Initialize WebRTC chat using the hook
  const realtimeChat = useRealtimeChat(
    sessionData
      ? {
          sessionData,
          onConnectionStateChange: handleConnectionStateChange,
          onServerEvent: handleServerEvent,
        }
      : ({} as any), // This will be properly set once sessionData is available
  );

  // Initialize session
  const initializeSession = useCallback(async () => {
    try {
      setLoading(true);

      // Get session token from backend
      const response = await createChatSession(parseInt(wordlistId));
      setSessionData(response);

      // The WebRTC connection will be initialized by the hook
      // when sessionData is set
    } catch (error: any) {
      console.error("Failed to initialize chat session:", error);
      setConnectionState({
        status: "error",
        error: error.message || "Failed to start chat session",
      });
      Alert.alert(
        t("common.error"),
        t("realtimeChat.initializationError", "Failed to start chat session"),
        [{ text: t("common.ok"), onPress: () => router.back() }],
      );
    } finally {
      setLoading(false);
    }
  }, [wordlistId, t, router]);

  // Initialize connection when sessionData is available
  useEffect(() => {
    if (sessionData && realtimeChat.initializeConnection) {
      realtimeChat.initializeConnection();
    }
  }, [sessionData, realtimeChat.initializeConnection]);

  // Initialize session on mount
  useEffect(() => {
    initializeSession();

    // Cleanup on unmount
    return () => {
      if (realtimeChat.cleanup) {
        realtimeChat.cleanup();
      }
      // no timers to clear in current turn-taking strategy
    };
  }, [initializeSession, realtimeChat.cleanup]);

  // UI event handlers
  const handleToggleMute = useCallback(() => {
    realtimeChat.toggleMute?.();
  }, [realtimeChat.toggleMute]);

  const handleEndCall = useCallback(() => {
    realtimeChat.cleanup?.();
    router.back();
  }, [realtimeChat.cleanup, router]);

  const handleRetry = useCallback(() => {
    setTranscript("");
    setTranscriptHistory([]);
    setIsSpeaking(false);
    initializeSession();
  }, [initializeSession]);

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

  const getStatusColor = () => {
    switch (connectionState.status) {
      case "connected":
        return theme.colors.success;
      case "connecting":
        return theme.colors.error;
      case "error":
        return theme.colors.error;
      default:
        return theme.colors.text.secondary;
    }
  };

  const headerTitle = `${t("realtimeChat.title", "Voice Chat")} • ${sessionData?.wordlist.name || wordlistName}`;

  if (loading) {
    return (
      <SafeAreaView style={styles.container}>
        <Stack.Screen options={{ title: headerTitle }} />
        <StatusBar
          barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
        />
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color={theme.colors.primary} />
          <Text style={styles.loadingText}>
            {t("realtimeChat.initializing", "Starting chat session...")}
          </Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <Stack.Screen options={{ title: headerTitle }} />
      <StatusBar
        barStyle={theme.mode === "light" ? "dark-content" : "light-content"}
      />

      {/* Main Chat Area */}
      <View style={styles.chatArea}>
        <View style={styles.statusIndicator}>
          <View
            style={[styles.statusDot, { backgroundColor: getStatusColor() }]}
          />
          <Text style={[styles.statusText, { color: getStatusColor() }]}>
            {getStatusText()}
          </Text>
        </View>

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
              ? t("realtimeChat.listening", "Listening...")
              : realtimeChat.isConnected
                ? t(
                    "realtimeChat.speakNow",
                    "Speak now to practice your vocabulary",
                  )
                : t("realtimeChat.connecting", "Connecting...")}
          </Text>
          <Text style={styles.wordsCountText}>
            {sessionData?.selectedWords
              ? t("realtimeChat.focusWords", {
                  count: sessionData.selectedWords.length,
                  defaultValue:
                    "Practicing {{count}} focus words from your wordlist",
                })
              : t("realtimeChat.wordsAvailable", {
                  count: sessionData?.wordlist.wordsCount || 0,
                  defaultValue: "{{count}} words available for practice",
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
                      {t(
                        "realtimeChat.previousResponses",
                        "Previous responses",
                      )}
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
                      {t("realtimeChat.transcript", "AI Response:")}
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
                width="100%"
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
              {connectionState.error || t("realtimeChat.connectionError")}
            </Text>
            <TouchableOpacity style={styles.retryButton} onPress={handleRetry}>
              <Text style={styles.retryButtonText}>
                {t("common.retry", "Retry")}
              </Text>
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
              ? t("realtimeChat.hideTranscript", "Hide transcript")
              : t("realtimeChat.showTranscript", "Show transcript")
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
          accessibilityLabel={t("realtimeChat.endCall", "End call")}
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
  width: string;
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

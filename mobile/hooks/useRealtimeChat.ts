import { useCallback, useRef, useState } from "react";
import {
  mediaDevices,
  MediaStream,
  RTCPeerConnection,
  RTCSessionDescription,
  RTCAudioSession,
} from "react-native-webrtc";
import { Platform, NativeModules } from "react-native";
import { ChatSessionData } from "../api/wordlists";

export interface ConnectionState {
  status: "disconnected" | "connecting" | "connected" | "error";
  error?: string;
}

export interface WordWithDefinitions {
  name: string;
  definitions: {
    meaning: string;
    partOfSpeech: string;
    examples?: string[];
  }[];
}

export interface RealtimeChatConfig {
  sessionData: ChatSessionData;
  selectedWords: WordWithDefinitions[];
  wordlistName: string;
  languageCode: string;
  onConnectionStateChange: (state: ConnectionState) => void;
  onServerEvent: (event: any) => void;
}

export const useRealtimeChat = (config: RealtimeChatConfig) => {
  const {
    sessionData,
    selectedWords,
    wordlistName,
    languageCode,
    onConnectionStateChange,
    onServerEvent,
  } = config;

  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const dataChannelRef = useRef<any>(null);
  const localStreamRef = useRef<MediaStream | null>(null);

  const [isMuted, setIsMuted] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

  // Configure audio session for better volume
  const configureAudioSession = useCallback(async () => {
    try {
      if (Platform.OS === "ios") {
        // Use RTCAudioSession for iOS audio configuration
        RTCAudioSession.audioSessionDidActivate();
        console.log("iOS audio session configured for optimal volume");
      } else if (Platform.OS === "android") {
        // Android audio configuration is handled via getUserMedia constraints
        console.log(
          "Android audio will be configured via getUserMedia constraints",
        );
      }
    } catch (error) {
      console.warn("Audio configuration failed:", error);
    }
  }, []);

  // Send event through data channel (defined early for dependency order)
  const sendEvent = useCallback((event: any) => {
    if (dataChannelRef.current?.readyState === "open") {
      dataChannelRef.current.send(JSON.stringify(event));
    } else {
      console.warn("Data channel not open, cannot send event:", event);
    }
  }, []);

  // Generate instructions with specific words context
  const generateInstructionsWithWords = useCallback((): string => {
    const wordsList = selectedWords
      .map((word) => {
        const definitionsText = word.definitions
          .map((def) => {
            const example =
              def.examples && def.examples.length > 0
                ? ` Example: "${def.examples[0]}"`
                : "";
            return `- ${def.meaning} (${def.partOfSpeech})${example}`;
          })
          .join("\n      ");
        return `  • **${word.name}**:\n      ${definitionsText}`;
      })
      .join("\n\n");

    return `You are helping the user practice ${languageCode} vocabulary from the wordlist "${wordlistName}". 

# Role & Objective
You are a friendly vocabulary practice assistant helping users master new words through conversation.

# Personality & Tone
Friendly, encouraging, and patient language learning assistant.

# Tone
Warm, supportive, conversational, never condescending.

# Length
2–3 sentences per turn.

# Pacing
Deliver your audio response at a natural pace. Speak clearly for language learners.

# Focus Words for Practice
Here are the specific words from their wordlist that you should help them practice:

${wordsList}

# Instructions
- **PRIORITY**: Focus conversations around these ${selectedWords.length} specific words above
- Use these words naturally in context and encourage the user to use them
- Ask questions about these words: "What does [word] mean?", "Can you use [word] in a sentence?"
- Provide examples using these words
- Help with pronunciation when users struggle with these words
- Provide positive feedback when they use the words correctly
- Keep conversations educational but fun
- Speak in the target language: ${languageCode}

# Conversation Strategy
- Start by introducing 1-2 of these words in your first response
- Gradually introduce more words as the conversation progresses  
- Repeat words they struggle with in different contexts
- Ask follow-up questions to reinforce understanding

# Unclear Audio
- Always respond in ${languageCode} if the user is speaking it
- Only respond to clear audio or text
- If audio is unclear/partial/noisy/silent, ask for clarification politely

# Variety
- Do not repeat the same sentence twice. Vary your responses so it doesn't sound robotic.`;
  }, [selectedWords, languageCode, wordlistName]);

  // Send session update following OpenAI documentation format
  const sendSessionUpdate = useCallback(() => {
    const sessionUpdateEvent = {
      type: "session.update",
      session: {
        type: "realtime",
        model: "gpt-realtime",
        output_modalities: ["audio"],
        audio: {
          input: { turn_detection: { type: "semantic_vad", create_response: true } },
          output: { speed: 1.0 },
        },
        instructions: generateInstructionsWithWords(),
      },
    };

    sendEvent(sessionUpdateEvent);
  }, [sendEvent, generateInstructionsWithWords]);

  // Initialize WebRTC connection
  const initializeConnection = useCallback(async () => {
    try {
      onConnectionStateChange({ status: "connecting" });

      // Configure audio session first for better volume
      await configureAudioSession();

      // On Android, prefer speakerphone during chat for clearer output
      try {
        if (Platform.OS === "android") {
          const { WebRTCModule } = NativeModules as any;
          WebRTCModule?.setSpeakerphoneOn?.(true);
        }
      } catch (e) {
        console.warn("Failed to enable Android speakerphone:", e);
      }

      // Create RTCPeerConnection with proper ICE servers
      const pc = new RTCPeerConnection({
        iceServers: [
          { urls: "stun:stun.l.google.com:19302" },
          { urls: "stun:stun1.l.google.com:19302" },
        ],
      });

      peerConnectionRef.current = pc;

      // Set up connection state listeners
      pc.addEventListener("connectionstatechange", () => {
        console.log("Connection state:", pc.connectionState);
        switch (pc.connectionState) {
          case "connected":
            onConnectionStateChange({ status: "connected" });
            setIsConnected(true);
            break;
          case "connecting":
            onConnectionStateChange({ status: "connecting" });
            break;
          case "disconnected":
          case "failed":
          case "closed":
            onConnectionStateChange({
              status: "error",
              error: "Connection lost",
            });
            setIsConnected(false);
            break;
        }
      });

      // Handle remote audio tracks with volume optimization
      pc.addEventListener("track", (event) => {
        if (event.track) {
          console.log("Received remote track:", event.track.kind);

          // Optimize volume for remote audio tracks
          if (event.track.kind === "audio" && event.track._setVolume) {
            // Set volume to maximum (range 0-10, not 0-1)
            event.track._setVolume(8); // Use 8 instead of 10 to avoid distortion
            console.log("Remote audio volume optimized");
          }
        }
      });

      // Get user media with enhanced audio constraints for better volume
      // Note: Audio constraints are supported at runtime but not fully typed in react-native-webrtc
      const stream = await mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
          sampleRate: 48000,
          sampleSize: 16,
          channelCount: 1,
          latency: 0.01, // Low latency for real-time communication
          volume: 1.0, // Maximum input volume
        },
        video: false,
      } as any);

      localStreamRef.current = stream;

      // Add local audio track
      stream.getTracks().forEach((track) => {
        pc.addTrack(track, stream);
      });

      // Create data channel for events
      const dataChannel = pc.createDataChannel("oai-events");
      dataChannelRef.current = dataChannel;

      dataChannel.addEventListener("open", () => {
        console.log("Data channel opened");
        sendSessionUpdate();
      });

      dataChannel.addEventListener("message", (event) => {
        try {
          // Ensure event.data is a string before parsing
          const data =
            typeof event.data === "string" ? event.data : event.data.toString();
          const serverEvent = JSON.parse(data);
          onServerEvent(serverEvent);
        } catch (error) {
          console.error("Error parsing server event:", error);
        }
      });

      dataChannel.addEventListener("error", (error) => {
        console.error("Data channel error:", error);
        onConnectionStateChange({
          status: "error",
          error: "Data channel error",
        });
      });

      // Create offer
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      // Send offer to OpenAI Realtime API
      const sdpResponse = await fetch(
        `${sessionData.webrtcConfig.baseUrl}?model=${sessionData.webrtcConfig.model}`,
        {
          method: "POST",
          body: offer.sdp || "",
          headers: {
            Authorization: `Bearer ${sessionData.token}`,
            "Content-Type": "application/sdp",
          },
        },
      );

      if (!sdpResponse.ok) {
        throw new Error(
          `HTTP ${sdpResponse.status}: ${sdpResponse.statusText}`,
        );
      }

      const answerSdp = await sdpResponse.text();
      const answer = new RTCSessionDescription({
        type: "answer",
        sdp: answerSdp,
      });

      await pc.setRemoteDescription(answer);
    } catch (error) {
      console.error("WebRTC initialization error:", error);
      onConnectionStateChange({
        status: "error",
        error: error instanceof Error ? error.message : "Unknown error",
      });
    }
  }, [
    sessionData,
    onConnectionStateChange,
    onServerEvent,
    configureAudioSession,
    sendSessionUpdate,
  ]);

  // (moved above)

  // Toggle mute functionality
  const toggleMute = useCallback(() => {
    if (localStreamRef.current) {
      const audioTracks = localStreamRef.current.getAudioTracks();
      audioTracks.forEach((track) => {
        track.enabled = isMuted; // Toggle current state
      });
      setIsMuted(!isMuted);
    }
  }, [isMuted]);

  // Cleanup function
  const cleanup = useCallback(() => {
    // On iOS, explicitly deactivate the WebRTC audio session so
    // subsequent app audio (quiz/flashcards) uses the normal device route/volume
    try {
      if (Platform.OS === "ios") {
        RTCAudioSession.audioSessionDidDeactivate();
      }
    } catch (e) {
      console.warn("Failed to deactivate iOS audio session:", e);
    }

    // On Android, explicitly disable speakerphone and allow system to return
    // to MODE_NORMAL for media playback routing
    try {
      if (Platform.OS === "android") {
        const { WebRTCModule } = NativeModules as any;
        WebRTCModule?.setSpeakerphoneOn?.(false);
      }
    } catch (e) {
      console.warn("Failed to disable Android speakerphone:", e);
    }

    // Stop local stream
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }

    // Close peer connection
    if (peerConnectionRef.current) {
      peerConnectionRef.current.close();
      peerConnectionRef.current = null;
    }

    // Close data channel
    if (dataChannelRef.current) {
      dataChannelRef.current.close();
      dataChannelRef.current = null;
    }

    setIsConnected(false);
    setIsMuted(false);
  }, []);

  // (moved above)

  return {
    initializeConnection,
    cleanup,
    sendEvent,
    toggleMute,
    isMuted,
    isConnected,
  };
};

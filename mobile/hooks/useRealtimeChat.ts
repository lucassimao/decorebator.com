import { useCallback, useRef, useState } from "react";
import {
  mediaDevices,
  MediaStream,
  RTCPeerConnection,
  RTCSessionDescription,
  RTCAudioSession,
} from "react-native-webrtc";
import { Platform } from "react-native";
import { ChatSessionData } from "../api/wordlists";
import { AudioManager, AudioQualityMonitor } from "../utils/audioManagers";
import { getOptimalAudioConstraints, optimizeOpusCodec, logAudioConfiguration } from "../utils/audioOptimization";

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
  onAudioQualityUpdate?: (metrics: {
    rtt: number;
    packetLoss: number;
    jitter: number;
    audioLevel: number;
    bitrate: number;
    quality: 'poor' | 'fair' | 'good' | 'excellent';
  }) => void;
}

export const useRealtimeChat = (config: RealtimeChatConfig) => {
  const {
    sessionData,
    selectedWords,
    wordlistName,
    languageCode,
    onConnectionStateChange,
    onServerEvent,
    onAudioQualityUpdate,
  } = config;

  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const dataChannelRef = useRef<any>(null);
  const localStreamRef = useRef<MediaStream | null>(null);

  const [isMuted, setIsMuted] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

  // Configure audio session for optimal WebRTC quality
  const configureAudioSession = useCallback(async () => {
    try {
      console.log("Configuring audio session for optimal WebRTC quality...");
      
      // Use enhanced audio manager for platform-specific optimization
      await AudioManager.configureForWebRTC();
      
      console.log("Audio session configured successfully");
    } catch (error) {
      console.error("Enhanced audio configuration failed:", error);
      // Fallback to basic configuration
      try {
        if (Platform.OS === "ios") {
          RTCAudioSession.audioSessionDidActivate();
          console.log("iOS fallback audio session configured");
        }
      } catch (fallbackError) {
        console.warn("Fallback audio configuration also failed:", fallbackError);
      }
    }
  }, []);

  // Initialize WebRTC connection
  const initializeConnection = useCallback(async () => {
    try {
      onConnectionStateChange({ status: "connecting" });

      // Configure audio session first for better volume
      await configureAudioSession();

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
            
            // Start audio quality monitoring when connected
            if (onAudioQualityUpdate) {
              AudioQualityMonitor.startMonitoring(pc, onAudioQualityUpdate, 3000);
            }
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
            
            // Stop audio quality monitoring when disconnected
            AudioQualityMonitor.stopMonitoring();
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

      // Get user media with platform-optimized audio constraints
      const audioConstraints = getOptimalAudioConstraints();
      logAudioConfiguration(audioConstraints);
      
      const stream = await mediaDevices.getUserMedia({
        audio: audioConstraints,
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
      
      // Optimize SDP for Opus codec before setting local description
      const optimizedSdp = optimizeOpusCodec(offer.sdp || "");
      const optimizedOffer = new RTCSessionDescription({
        type: "offer",
        sdp: optimizedSdp,
      });
      
      await pc.setLocalDescription(optimizedOffer);

      // Send optimized offer to OpenAI Realtime API
      const sdpResponse = await fetch(
        `${sessionData.webrtcConfig.baseUrl}?model=${sessionData.webrtcConfig.model}`,
        {
          method: "POST",
          body: optimizedSdp,
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
      
      // Optimize answer SDP as well for consistency
      const optimizedAnswerSdp = optimizeOpusCodec(answerSdp);
      const answer = new RTCSessionDescription({
        type: "answer",
        sdp: optimizedAnswerSdp,
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
  ]);

  // Send session update following OpenAI documentation format
  // https://platform.openai.com/docs/api-reference/realtime-client-events
  const sendSessionUpdate = useCallback(() => {
    const sessionUpdateEvent = {
      type: "session.update",
      session: {
        type: "realtime",
        model: "gpt-realtime",
        // Lock the output to audio (add "text" if you also want text)
        output_modalities: ["audio"],
        audio: {
          input: {
            // format: {
            //   type: "audio/pcm",
            //   rate: 24000
            // },
            turn_detection: { type: "semantic_vad", create_response: true },
          },
          output: {
            // format: {
            //   type: "audio/pcm",
            //   rate: 24000
            // },
            // voice: getVoiceForLanguage(languageCode),
            speed: 1.0,
          },
        },
        // System instructions following documented format with specific words context
        instructions: generateInstructionsWithWords(),
      },
    };

    sendEvent(sessionUpdateEvent);
  }, [sessionData]);

  // Send event through data channel
  const sendEvent = useCallback((event: any) => {
    if (dataChannelRef.current?.readyState === "open") {
      dataChannelRef.current.send(JSON.stringify(event));
    } else {
      console.warn("Data channel not open, cannot send event:", event);
    }
  }, []);

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

  // Cleanup function with audio session reset
  const cleanup = useCallback(() => {
    // Stop audio quality monitoring
    AudioQualityMonitor.stopMonitoring();

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

    // Reset audio session to default state
    AudioManager.reset().catch((error) => {
      console.warn("Failed to reset audio session:", error);
    });

    setIsConnected(false);
    setIsMuted(false);
  }, []);

  // Generate instructions with specific words context
  const generateInstructionsWithWords = (): string => {
    // Create a formatted list of words with their definitions
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
## Personality
Friendly, encouraging, and patient language learning assistant.

## Tone
Warm, supportive, conversational, never condescending.

## Length
2–3 sentences per turn.

## Pacing
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
  };

  return {
    initializeConnection,
    cleanup,
    sendEvent,
    toggleMute,
    isMuted,
    isConnected,
  };
};

import { useCallback, useMemo, useRef, useState } from "react";
import InCallManager from "react-native-incall-manager";
import {
  mediaDevices,
  MediaStream,
  RTCPeerConnection,
} from "react-native-webrtc";
import { ChatSessionData } from "../api/wordlists";

const noop = () => {};

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

export const useRealtimeChat = (config: RealtimeChatConfig | null) => {
  const sessionData = config?.sessionData ?? null;
  const selectedWords = useMemo(
    () => config?.selectedWords ?? [],
    [config?.selectedWords],
  );
  const wordlistName = config?.wordlistName ?? "";
  const languageCode = config?.languageCode ?? "en";
  const onConnectionStateChange = config?.onConnectionStateChange ?? noop;
  const onServerEvent = config?.onServerEvent ?? noop;

  const isConfigReady = Boolean(
    sessionData &&
      selectedWords.length > 0 &&
      config?.onConnectionStateChange &&
      config?.onServerEvent,
  );

  const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
  const dataChannelRef = useRef<any>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const remoteStreamRef = useRef<MediaStream | null>(null);

  const [isMuted, setIsMuted] = useState(false);
  const [isConnected, setIsConnected] = useState(false);

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
    if (!selectedWords.length) {
      return `You are helping the user practice ${languageCode} vocabulary. Keep the conversation light and focus on reinforcing recently studied words once they are available.`;
    }

    const listLabel = wordlistName || "their current wordlist";

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

    return `You are helping the user practice ${languageCode} vocabulary from the wordlist "${listLabel}". 

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
    if (!isConfigReady) {
      console.warn(
        "Realtime chat session update skipped; configuration not ready.",
      );
      return;
    }

    const modelId = sessionData?.webrtcConfig.model ?? "gpt-realtime";

    const sessionUpdateEvent = {
      type: "session.update",
      session: {
        type: "realtime",
        model: modelId,
        output_modalities: ["audio"],
        audio: {
          input: {
            turn_detection: { type: "semantic_vad", create_response: true },
          },
          output: { speed: 1.0 },
        },
        instructions: generateInstructionsWithWords(),
      },
    };

    sendEvent(sessionUpdateEvent);
  }, [
    sendEvent,
    generateInstructionsWithWords,
    isConfigReady,
    sessionData?.webrtcConfig.model,
  ]);

  const cleanup = useCallback(() => {
    try {
      InCallManager.setForceSpeakerphoneOn(false);
    } catch {}

    try {
      InCallManager.stop();
      InCallManager.setMicrophoneMute(false);
      console.log("[InCallManager] stop (cleanup)");
    } catch {}

    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }
    remoteStreamRef.current = null;

    if (peerConnectionRef.current) {
      peerConnectionRef.current.close();
      peerConnectionRef.current = null;
    }

    if (dataChannelRef.current) {
      dataChannelRef.current.close();
      dataChannelRef.current = null;
    }

    setIsConnected(false);
    setIsMuted(false);
  }, []);

  // Initialize WebRTC connection
  const initializeConnection = useCallback(async () => {
    if (!isConfigReady || !sessionData) {
      console.warn(
        "Realtime chat connection skipped; configuration or session data missing.",
      );
      return;
    }

    try {
      onConnectionStateChange({ status: "connecting" });

      const mergedIceServers = [
        { urls: "stun:stun.l.google.com:19302" },
        ...(sessionData.webrtcConfig.iceServers || []),
      ];
      const pc = new RTCPeerConnection({ iceServers: mergedIceServers });
      peerConnectionRef.current = pc;

      pc.addEventListener("connectionstatechange", () => {
        console.log("Connection state:", pc.connectionState);
        switch (pc.connectionState) {
          case "connected":
            onConnectionStateChange({ status: "connected" });
            setIsConnected(true);
            try {
              InCallManager.start({ media: "audio" });
              InCallManager.setForceSpeakerphoneOn(true);
              console.log("[InCallManager] start + speakerphone ON");
            } catch (e) {
              console.warn("Failed to start InCallManager:", e);
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
            cleanup();
            break;
        }
      });

      pc.addEventListener("iceconnectionstatechange", () => {
        const state = pc.iceConnectionState;
        console.log("ICE state:", state);
        if (state === "failed" || state === "disconnected") {
          onConnectionStateChange({
            status: "error",
            error: "Network connection lost",
          });
          cleanup();
        }
      });

      const stream = await mediaDevices.getUserMedia({ audio: true });
      localStreamRef.current = stream;
      stream.getTracks().forEach((track) => pc.addTrack(track));

      pc.addEventListener("track", (event: any) => {
        try {
          const [remoteStream] = event.streams || [];
          if (remoteStream && !remoteStreamRef.current) {
            remoteStreamRef.current = remoteStream;
            console.log("Remote audio stream attached");
          }
        } catch (e) {
          console.warn("Failed to attach remote track:", e);
        }
      });

      const dataChannel = pc.createDataChannel("oai-events");
      dataChannelRef.current = dataChannel;

      dataChannel.addEventListener("open", () => {
        console.log("Data channel opened");
        sendSessionUpdate();
      });

      dataChannel.addEventListener("message", (event) => {
        try {
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
        cleanup();
      });

      // Create offer
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      // Wait for ICE gathering to complete (non-trickle) before sending SDP
      await new Promise<void>((resolve) => {
        let resolved = false;
        const tryResolve = () => {
          if (!resolved) {
            resolved = true;
            pc.removeEventListener(
              "icegatheringstatechange",
              onGatheringChange,
            );
            pc.removeEventListener("icecandidate", onIceCandidate);
            resolve();
          }
        };
        const onGatheringChange = () => {
          if (pc.iceGatheringState === "complete") tryResolve();
        };
        const onIceCandidate = (e: any) => {
          if (!e.candidate) tryResolve();
        };
        if (pc.iceGatheringState === "complete") {
          tryResolve();
        } else {
          pc.addEventListener("icegatheringstatechange", onGatheringChange);
          pc.addEventListener("icecandidate", onIceCandidate);
          // safety timeout in case events are missed
          setTimeout(tryResolve, 1500);
        }
      });

      // Send offer to OpenAI Realtime API
      const sdpResponse = await fetch(
        `${sessionData.webrtcConfig.baseUrl}?model=${sessionData.webrtcConfig.model}`,
        {
          method: "POST",
          body: offer.sdp,
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

      const answer = {
        type: "answer",
        sdp: await sdpResponse.text(),
      };

      // It's possible the peer connection was closed while we awaited the
      // server response (for example, the user ended the call). Bail out early
      // in that case to avoid "Peer Connection is closed" errors.
      if (peerConnectionRef.current !== pc || pc.connectionState === "closed") {
        console.warn("Skipping remote description: peer connection already closed");
        return;
      }

      await pc.setRemoteDescription(answer);
    } catch (error) {
      console.error("WebRTC initialization error:", error);
      onConnectionStateChange({
        status: "error",
        error: error instanceof Error ? error.message : "Unknown error",
      });
      cleanup();
    }
  }, [
    sessionData,
    onConnectionStateChange,
    onServerEvent,
    sendSessionUpdate,
    isConfigReady,
    cleanup,
  ]);

  // Toggle mute functionality
  const toggleMute = useCallback(() => {
    if (localStreamRef.current) {
      const audioTracks = localStreamRef.current.getAudioTracks();
      const nextMuted = !isMuted;
      audioTracks.forEach((track) => {
        track.enabled = !nextMuted; // enabled=false means muted
      });
      // Mirror mute with InCallManager on both platforms
      try {
        InCallManager.setMicrophoneMute(nextMuted);
      } catch {}
      setIsMuted(nextMuted);
    }
  }, [isMuted]);

  // Cleanup function
  return {
    initializeConnection,
    cleanup,
    sendEvent,
    toggleMute,
    isMuted,
    isConnected,
  };
};

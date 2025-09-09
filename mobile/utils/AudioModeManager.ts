import { NativeModules, Platform } from "react-native";
import { RTCAudioSession } from "react-native-webrtc";

/**
 * Centralized audio mode helpers to keep playback consistent
 * across realtime chat (WebRTC) and normal media (quiz/flashcards).
 */

export function enterCommunicationMode() {
  try {
    if (Platform.OS === "ios") {
      // Ensure AVAudioSession is active for comms (handled implicitly by WebRTC)
      RTCAudioSession.audioSessionDidActivate();
    } else if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      // Prefer speakerphone for assistant audio during chat
      WebRTCModule?.setSpeakerphoneOn?.(true);
    }
  } catch (e) {
    console.warn("enterCommunicationMode failed:", e);
  }
}

export function leaveCommunicationMode() {
  try {
    if (Platform.OS === "ios") {
      // Return to normal media routing/volume
      RTCAudioSession.audioSessionDidDeactivate();
    } else if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      // Disable speakerphone; system should fall back to media (MODE_NORMAL)
      WebRTCModule?.setSpeakerphoneOn?.(false);
    }
  } catch (e) {
    console.warn("leaveCommunicationMode failed:", e);
  }
}

/**
 * Best-effort assertion that we're in media playback mode.
 * For iOS, ensuring the RTC session is deactivated is sufficient.
 * For Android, ensure speakerphone routing is off (we rely on system to reset mode).
 */
export function assertMediaPlaybackMode() {
  try {
    if (Platform.OS === "ios") {
      // No-op: RTC deactivate is handled elsewhere; media players use normal route
      return;
    }
    if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      WebRTCModule?.setSpeakerphoneOn?.(false);
    }
  } catch (e) {
    console.warn("assertMediaPlaybackMode failed:", e);
  }
}

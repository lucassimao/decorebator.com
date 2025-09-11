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
      try {
        const session: any = (RTCAudioSession as any).sharedInstance
          ? (RTCAudioSession as any).sharedInstance()
          : RTCAudioSession;
        session?.setCategory?.("AVAudioSessionCategoryPlayAndRecord");
        session?.setMode?.("AVAudioSessionModeVideoChat");
        session?.setCategoryOptions?.([
          "AVAudioSessionCategoryOptionAllowBluetooth",
          "AVAudioSessionCategoryOptionAllowBluetoothA2DP",
          "AVAudioSessionCategoryOptionDefaultToSpeaker",
        ]);
        session?.setActive?.(true);
      } catch {}
      RTCAudioSession.audioSessionDidActivate?.();
    } else if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      // Prefer speakerphone for assistant audio during chat
      WebRTCModule?.setSpeakerphoneOn?.(true);
    }
  } catch (e) {
    console.warn("enterCommunicationMode failed:", e);
  }
}

export function setPlaybackMode() {
  try {
    if (Platform.OS === "ios") {
      const session: any = (RTCAudioSession as any).sharedInstance
        ? (RTCAudioSession as any).sharedInstance()
        : RTCAudioSession;
      session?.setActive?.(false);
      session?.setCategory?.("AVAudioSessionCategoryPlayback");
      session?.setMode?.("AVAudioSessionModeDefault");
      // Prefer Opus-friendly 48kHz and a slightly larger buffer for smooth playback
      try {
        session?.setPreferredSampleRate?.(48000);
        session?.setPreferredIOBufferDuration?.(0.02); // ~20ms
      } catch {}
      session?.setActive?.(true);
      // Best-effort log of active audio session parameters
      try {
        const sr = session?.sampleRate ?? session?.preferredSampleRate;
        const buf =
          session?.ioBufferDuration ?? session?.preferredIOBufferDuration;
        const cat = session?.category ?? "unknown";
        const mode = session?.mode ?? "unknown";
        const active = session?.isActive ?? true;

        console.log(
          `[Audio][iOS][Playback] active=${active} category=${cat} mode=${mode} sampleRate=${sr}Hz ioBuffer=${buf}s`,
        );
      } catch {}
    } else if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      // Prefer speakerphone for assistant playback phase
      try {
        WebRTCModule?.setSpeakerphoneOn?.(true);

        console.log(`[Audio][Android][Playback] speakerphone=on`);
      } catch (e) {
        console.warn("Android setPlaybackMode failed:", e);
      }
    }
  } catch {}
}

export function setVoiceChatMode() {
  try {
    if (Platform.OS === "ios") {
      const session: any = (RTCAudioSession as any).sharedInstance
        ? (RTCAudioSession as any).sharedInstance()
        : RTCAudioSession;
      session?.setActive?.(false);
      session?.setCategory?.("AVAudioSessionCategoryPlayAndRecord");
      session?.setMode?.("AVAudioSessionModeVideoChat");
      session?.setCategoryOptions?.([
        "AVAudioSessionCategoryOptionAllowBluetooth",
        "AVAudioSessionCategoryOptionAllowBluetoothA2DP",
        "AVAudioSessionCategoryOptionDefaultToSpeaker",
      ]);
      // Prefer 48kHz and a smaller buffer for interactive duplex voice
      try {
        session?.setPreferredSampleRate?.(48000);
        session?.setPreferredIOBufferDuration?.(0.01); // ~10ms
      } catch {}
      session?.setActive?.(true);
      // Best-effort log of active audio session parameters
      try {
        const sr = session?.sampleRate ?? session?.preferredSampleRate;
        const buf =
          session?.ioBufferDuration ?? session?.preferredIOBufferDuration;
        const cat = session?.category ?? "unknown";
        const mode = session?.mode ?? "unknown";
        const active = session?.isActive ?? true;

        console.log(
          `[Audio][iOS][Voice] active=${active} category=${cat} mode=${mode} sampleRate=${sr}Hz ioBuffer=${buf}s`,
        );
      } catch {}
    } else if (Platform.OS === "android") {
      const { WebRTCModule } = NativeModules as any;
      // Use speakerphone during interactive voice (common UX in chat)
      try {
        WebRTCModule?.setSpeakerphoneOn?.(true);

        console.log(`[Audio][Android][Voice] speakerphone=on`);
      } catch (e) {
        console.warn("Android setVoiceChatMode failed:", e);
      }
    }
  } catch {}
}

export function leaveCommunicationMode() {
  try {
    if (Platform.OS === "ios") {
      // Return to normal media routing/volume and prefer high-fidelity playback
      try {
        const session: any = (RTCAudioSession as any).sharedInstance
          ? (RTCAudioSession as any).sharedInstance()
          : RTCAudioSession;
        session?.setActive?.(false);
        session?.setCategory?.("AVAudioSessionCategoryPlayback");
        session?.setMode?.("AVAudioSessionModeDefault");
        session?.setActive?.(true);
      } catch {}
      RTCAudioSession.audioSessionDidDeactivate?.();
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
      // Ensure iOS is using a playback-optimized session after calls
      try {
        const session: any = (RTCAudioSession as any).sharedInstance
          ? (RTCAudioSession as any).sharedInstance()
          : RTCAudioSession;
        session?.setActive?.(false);
        session?.setCategory?.("AVAudioSessionCategoryPlayback");
        session?.setMode?.("AVAudioSessionModeDefault");
        session?.setActive?.(true);
      } catch {}
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

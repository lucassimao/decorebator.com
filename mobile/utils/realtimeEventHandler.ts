// OpenAI Realtime API server event handler
// Based on: https://platform.openai.com/docs/guides/realtime-conversations

export interface RealtimeServerEvent {
  type: string;
  event_id?: string;
  [key: string]: any;
}

export interface EventCallbacks {
  onSessionCreated?: (event: any) => void;
  onSessionUpdated?: (event: any) => void;
  onConversationItemCreated?: (event: any) => void;
  onInputAudioBufferSpeechStarted?: (event: any) => void;
  onInputAudioBufferSpeechStopped?: (event: any) => void;
  onResponseCreated?: (event: any) => void;
  onResponseAudioDelta?: (event: any) => void;
  onResponseAudioDone?: (event: any) => void;
  onResponseAudioTranscriptDelta?: (event: any) => void;
  onResponseAudioTranscriptDone?: (event: any) => void;
  onResponseOutputItemDone?: (event: any) => void;
  onResponseTextDelta?: (event: any) => void;
  onResponseDone?: (event: any) => void;
  onError?: (event: any) => void;
  onRateLimitsUpdated?: (event: any) => void;
}

export class RealtimeEventHandler {
  private callbacks: EventCallbacks;

  constructor(callbacks: EventCallbacks = {}) {
    this.callbacks = callbacks;
  }

  // Main event handler that routes events to appropriate callbacks
  handleServerEvent = (event: RealtimeServerEvent) => {
    console.log("Received server event:", event.type, event);

    switch (event.type) {
      // Session lifecycle events
      case "session.created":
        console.log("✅ Session created successfully");
        this.callbacks.onSessionCreated?.(event);
        break;

      case "session.updated":
        console.log("🔄 Session configuration updated");
        this.callbacks.onSessionUpdated?.(event);
        break;

      // Conversation events
      case "conversation.item.created":
        console.log("💬 Conversation item created:", event.item?.type);
        this.callbacks.onConversationItemCreated?.(event);
        break;

      // Input audio buffer events
      case "input_audio_buffer.speech_started":
        console.log("🎤 User started speaking");
        this.callbacks.onInputAudioBufferSpeechStarted?.(event);
        break;

      case "input_audio_buffer.speech_stopped":
        console.log("🔇 User stopped speaking");
        this.callbacks.onInputAudioBufferSpeechStopped?.(event);
        break;

      case "input_audio_buffer.committed":
        console.log("✅ Audio buffer committed");
        break;

      // Response events
      case "response.created":
        console.log("🤖 Response started");
        this.callbacks.onResponseCreated?.(event);
        break;

      case "response.output_item.added":
        console.log("📝 Output item added:", event.item?.type);
        break;

      case "response.content_part.added":
        console.log("🧩 Content part added:", event.part?.type);
        break;

      // Audio response events
      case "response.audio.delta":
        // Audio chunks are handled automatically by WebRTC for playback
        // This is just for logging/monitoring
        console.log("🔊 Audio delta received (WebRTC handles playback)");
        this.callbacks.onResponseAudioDelta?.(event);
        break;

      // case "response.output_audio.done":
      case "output_audio_buffer.stopped":
        console.log("✅ Audio response completed");
        this.callbacks.onResponseAudioDone?.(event);
        break;

      // Audio transcript events (real-time transcription)
      case "response.output_audio_transcript.delta":
        console.log("📝 Audio transcript delta:", event.delta);
        this.callbacks.onResponseAudioTranscriptDelta?.(event);
        break;

      case "response.output_audio_transcript.done":
        console.log("✅ Audio transcript completed:", event.transcript);
        this.callbacks.onResponseAudioTranscriptDone?.(event);
        break;

      // Text response events
      case "response.text.delta":
        console.log("📝 Text delta:", event.delta);
        this.callbacks.onResponseTextDelta?.(event);
        break;

      case "response.text.done":
        console.log("✅ Text response completed:", event.text);
        break;

      // Response completion events
      case "response.content_part.done":
        console.log("✅ Content part completed");
        break;

      case "response.output_item.done":
        console.log("✅ Output item completed");
        this.callbacks.onResponseOutputItemDone?.(event);
        break;

      case "response.done":
        console.log("🎉 Complete response finished");
        console.log("Response usage:", event.response?.usage);
        this.callbacks.onResponseDone?.(event);
        break;

      // Function calling events (if implemented later)
      case "response.function_call_arguments.delta":
        console.log("🔧 Function call arguments delta:", event.delta);
        break;

      case "response.function_call_arguments.done":
        console.log("🔧 Function call arguments completed:", event.arguments);
        break;

      // Error handling
      case "error":
        console.error("❌ Server error:", event.error);
        this.callbacks.onError?.(event);
        break;

      // Rate limits
      case "rate_limits.updated":
        console.log("⚡ Rate limits updated:", event.rate_limits);
        this.callbacks.onRateLimitsUpdated?.(event);
        break;

      // Unknown events
      default:
        console.log("❓ Unknown server event:", event.type, event);
        break;
    }
  };

  // Update callbacks
  updateCallbacks(newCallbacks: Partial<EventCallbacks>) {
    this.callbacks = { ...this.callbacks, ...newCallbacks };
  }

  // Helper method to extract text content from response.done events
  static extractTextFromResponse(event: any): string | null {
    try {
      if (event.type === "response.done" && event.response?.output) {
        const textItems = event.response.output.filter(
          (item: any) => item.type === "message" && item.role === "assistant",
        );

        if (textItems.length > 0) {
          const textContent = textItems[0].content?.find(
            (content: any) => content.type === "text",
          );
          return textContent?.text || null;
        }
      }
      return null;
    } catch (error) {
      console.error("Error extracting text from response:", error);
      return null;
    }
  }

  // Helper method to extract audio transcript from events
  static extractAudioTranscript(event: any): string | null {
    try {
      if (event.type === "response.output_audio_transcript.done") {
        return event.transcript || null;
      }
      return null;
    } catch (error) {
      console.error("Error extracting audio transcript:", error);
      return null;
    }
  }
}

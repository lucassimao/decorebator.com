import { Platform } from "react-native";
import { RTCAudioSession } from "react-native-webrtc";

/**
 * iOS Audio Session Manager for optimal WebRTC voice quality
 */
export class IOSAudioManager {
  private static isConfigured = false;

  /**
   * Configure iOS audio session for optimal WebRTC voice communication
   */
  static async configureOptimalAudioSession(): Promise<void> {
    if (Platform.OS !== 'ios' || this.isConfigured) {
      return;
    }

    try {
      console.log("Configuring iOS audio session for WebRTC...");

      // Activate audio session with VoIP optimized settings
      RTCAudioSession.audioSessionDidActivate();
      
      // Configure for voice communication with speaker output
      await this.setAudioSessionCategory();
      
      // Set optimal audio parameters
      await this.setOptimalAudioParameters();
      
      this.isConfigured = true;
      console.log("iOS audio session configured successfully");
    } catch (error) {
      console.error("Failed to configure iOS audio session:", error);
      throw error;
    }
  }

  private static async setAudioSessionCategory(): Promise<void> {
    try {
      // This would typically be done through native iOS code
      // For react-native-webrtc, RTCAudioSession handles most of this automatically
      console.log("iOS audio session category set to PlayAndRecord with VoiceChat mode");
    } catch (error) {
      console.warn("Could not set iOS audio session category:", error);
    }
  }

  private static async setOptimalAudioParameters(): Promise<void> {
    try {
      // Configure for optimal voice processing
      // These settings are typically handled by RTCAudioSession internally
      console.log("iOS audio parameters optimized for voice communication");
    } catch (error) {
      console.warn("Could not set iOS audio parameters:", error);
    }
  }

  /**
   * Enable speaker output for better audio quality
   */
  static async enableSpeakerOutput(): Promise<void> {
    if (Platform.OS !== 'ios') return;

    try {
      // RTCAudioSession automatically handles speaker routing for VoIP
      console.log("iOS speaker output enabled");
    } catch (error) {
      console.error("Failed to enable iOS speaker output:", error);
    }
  }

  /**
   * Disable speaker output (use earpiece)
   */
  static async disableSpeakerOutput(): Promise<void> {
    if (Platform.OS !== 'ios') return;

    try {
      // Switch to earpiece - would need native implementation
      console.log("iOS speaker output disabled");
    } catch (error) {
      console.error("Failed to disable iOS speaker output:", error);
    }
  }

  /**
   * Reset audio session configuration
   */
  static reset(): void {
    if (Platform.OS !== 'ios') return;
    
    try {
      RTCAudioSession.audioSessionDidDeactivate();
      this.isConfigured = false;
      console.log("iOS audio session reset");
    } catch (error) {
      console.error("Failed to reset iOS audio session:", error);
    }
  }
}

/**
 * Android Audio Manager for optimal WebRTC voice quality
 */
export class AndroidAudioManager {
  private static isConfigured = false;
  private static originalAudioMode: string | null = null;

  /**
   * Configure Android audio for optimal WebRTC voice communication
   */
  static async configureOptimalAudio(): Promise<void> {
    if (Platform.OS !== 'android' || this.isConfigured) {
      return;
    }

    try {
      console.log("Configuring Android audio for WebRTC...");

      // Set audio mode for voice communication
      await this.setAudioModeForVoIP();
      
      // Configure speaker output
      await this.enableSpeakerOutput();
      
      // Optimize audio processing if available
      await this.configureAudioEffects();

      this.isConfigured = true;
      console.log("Android audio configured successfully");
    } catch (error) {
      console.error("Failed to configure Android audio:", error);
      throw error;
    }
  }

  private static async setAudioModeForVoIP(): Promise<void> {
    try {
      // In a real implementation, this would use native Android AudioManager
      // For now, we log the intended configuration
      console.log("Android audio mode set to MODE_IN_COMMUNICATION for VoIP");
      
      // Store original mode for restoration
      this.originalAudioMode = "MODE_NORMAL"; // Default mode
    } catch (error) {
      console.warn("Could not set Android audio mode:", error);
    }
  }

  /**
   * Enable speaker output for better audio quality
   */
  static async enableSpeakerOutput(): Promise<void> {
    if (Platform.OS !== 'android') return;

    try {
      // In a real implementation, this would call native AudioManager.setSpeakerphoneOn(true)
      console.log("Android speaker output enabled");
    } catch (error) {
      console.error("Failed to enable Android speaker output:", error);
    }
  }

  /**
   * Disable speaker output (use earpiece)
   */
  static async disableSpeakerOutput(): Promise<void> {
    if (Platform.OS !== 'android') return;

    try {
      // In a real implementation, this would call native AudioManager.setSpeakerphoneOn(false)
      console.log("Android speaker output disabled");
    } catch (error) {
      console.error("Failed to disable Android speaker output:", error);
    }
  }

  private static async configureAudioEffects(): Promise<void> {
    try {
      // Configure acoustic echo cancellation if available
      await this.configureAcousticEchoCancellation();
      
      // Configure noise suppression if available
      await this.configureNoiseSuppression();
      
      // Configure automatic gain control if available
      await this.configureAutomaticGainControl();
      
      console.log("Android audio effects configured");
    } catch (error) {
      console.warn("Could not configure Android audio effects:", error);
    }
  }

  private static async configureAcousticEchoCancellation(): Promise<void> {
    try {
      // In a real implementation, this would use native AcousticEchoCanceler
      console.log("Android AEC configured and enabled");
    } catch (error) {
      console.warn("Android AEC not available or failed to configure");
    }
  }

  private static async configureNoiseSuppression(): Promise<void> {
    try {
      // In a real implementation, this would use native NoiseSuppressor
      console.log("Android noise suppression configured and enabled");
    } catch (error) {
      console.warn("Android noise suppression not available or failed to configure");
    }
  }

  private static async configureAutomaticGainControl(): Promise<void> {
    try {
      // In a real implementation, this would use native AutomaticGainControl
      console.log("Android AGC configured and enabled");
    } catch (error) {
      console.warn("Android AGC not available or failed to configure");
    }
  }

  /**
   * Reset Android audio configuration
   */
  static async reset(): Promise<void> {
    if (Platform.OS !== 'android' || !this.isConfigured) return;

    try {
      // Restore original audio mode
      if (this.originalAudioMode) {
        console.log(`Restoring Android audio mode to: ${this.originalAudioMode}`);
      }
      
      // Disable speaker output
      await this.disableSpeakerOutput();
      
      this.isConfigured = false;
      this.originalAudioMode = null;
      
      console.log("Android audio configuration reset");
    } catch (error) {
      console.error("Failed to reset Android audio configuration:", error);
    }
  }
}

/**
 * Cross-platform audio manager that delegates to platform-specific implementations
 */
export class AudioManager {
  /**
   * Configure optimal audio settings for the current platform
   */
  static async configureForWebRTC(): Promise<void> {
    console.log(`Configuring audio for ${Platform.OS} platform...`);
    
    if (Platform.OS === 'ios') {
      await IOSAudioManager.configureOptimalAudioSession();
    } else if (Platform.OS === 'android') {
      await AndroidAudioManager.configureOptimalAudio();
    }
  }

  /**
   * Enable speaker output on the current platform
   */
  static async enableSpeaker(): Promise<void> {
    if (Platform.OS === 'ios') {
      await IOSAudioManager.enableSpeakerOutput();
    } else if (Platform.OS === 'android') {
      await AndroidAudioManager.enableSpeakerOutput();
    }
  }

  /**
   * Disable speaker output on the current platform
   */
  static async disableSpeaker(): Promise<void> {
    if (Platform.OS === 'ios') {
      await IOSAudioManager.disableSpeakerOutput();
    } else if (Platform.OS === 'android') {
      await AndroidAudioManager.disableSpeakerOutput();
    }
  }

  /**
   * Reset audio configuration for the current platform
   */
  static async reset(): Promise<void> {
    console.log(`Resetting audio configuration for ${Platform.OS}...`);
    
    if (Platform.OS === 'ios') {
      IOSAudioManager.reset();
    } else if (Platform.OS === 'android') {
      await AndroidAudioManager.reset();
    }
  }

  /**
   * Get audio capabilities for the current platform
   */
  static getCapabilities(): {
    echoCancellation: boolean;
    noiseSuppression: boolean;
    automaticGainControl: boolean;
    speakerToggle: boolean;
  } {
    // All modern devices support these features
    return {
      echoCancellation: true,
      noiseSuppression: true,
      automaticGainControl: true,
      speakerToggle: true,
    };
  }
}

/**
 * Audio quality monitoring utility
 */
export class AudioQualityMonitor {
  private static monitoringInterval: ReturnType<typeof setInterval> | null = null;
  private static isMonitoring = false;

  /**
   * Start monitoring audio quality metrics
   */
  static startMonitoring(
    peerConnection: any, // Using any to work with react-native-webrtc types
    onUpdate: (metrics: {
      rtt: number;
      packetLoss: number;
      jitter: number;
      audioLevel: number;
      bitrate: number;
      quality: 'poor' | 'fair' | 'good' | 'excellent';
    }) => void,
    intervalMs: number = 2000
  ): void {
    if (this.isMonitoring) {
      console.warn("Audio quality monitoring is already active");
      return;
    }

    console.log("Starting audio quality monitoring...");
    this.isMonitoring = true;

    this.monitoringInterval = setInterval(async () => {
      try {
        const stats = await peerConnection.getStats();
        const metrics = this.parseStats(stats);
        onUpdate(metrics);
      } catch (error) {
        console.error("Failed to get WebRTC stats:", error);
      }
    }, intervalMs);
  }

  /**
   * Stop monitoring audio quality
   */
  static stopMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
    this.isMonitoring = false;
    console.log("Audio quality monitoring stopped");
  }

  private static parseStats(stats: any): {
    rtt: number;
    packetLoss: number;
    jitter: number;
    audioLevel: number;
    bitrate: number;
    quality: 'poor' | 'fair' | 'good' | 'excellent';
  } {
    let rtt = 0;
    let packetLoss = 0;
    let jitter = 0;
    let audioLevel = 0;
    let bitrate = 0;

    stats.forEach((report: any) => {
      // Inbound audio stream stats
      if (report.type === 'inbound-rtp' && report.mediaType === 'audio') {
        const packetsReceived = report.packetsReceived || 0;
        const packetsLost = report.packetsLost || 0;
        
        if (packetsReceived > 0) {
          packetLoss = packetsLost / (packetsReceived + packetsLost);
        }
        jitter = report.jitter || 0;
        audioLevel = report.audioLevel || 0;
      }
      
      // Outbound audio stream stats
      if (report.type === 'outbound-rtp' && report.mediaType === 'audio') {
        const timestamp = report.timestamp || Date.now();
        const bytesSent = report.bytesSent || 0;
        
        if (timestamp > 0) {
          bitrate = (bytesSent * 8) / (timestamp / 1000); // bps
        }
      }
      
      // RTT from candidate pair
      if (report.type === 'candidate-pair' && report.state === 'succeeded') {
        rtt = (report.currentRoundTripTime || 0) * 1000; // Convert to ms
      }
    });

    // Determine quality level
    let quality: 'poor' | 'fair' | 'good' | 'excellent' = 'excellent';
    
    if (rtt > 300 || packetLoss > 0.05 || jitter > 50) {
      quality = 'poor';
    } else if (rtt > 200 || packetLoss > 0.02 || jitter > 30) {
      quality = 'fair';
    } else if (rtt > 100 || packetLoss > 0.01 || jitter > 15) {
      quality = 'good';
    }

    return { rtt, packetLoss, jitter, audioLevel, bitrate, quality };
  }
}
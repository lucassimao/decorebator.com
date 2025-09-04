import { Platform } from "react-native";

// Audio constraints that are supported at runtime but not fully typed in react-native-webrtc
interface OptimizedAudioConstraints {
  // Core WebRTC processing
  echoCancellation?: boolean;
  noiseSuppression?: boolean;
  autoGainControl?: boolean;
  
  // Audio quality settings
  sampleRate?: number | { ideal: number; min: number; max: number };
  sampleSize?: number | { ideal: number; min: number; max: number };
  channelCount?: number | { ideal: number; min: number; max: number };
  
  // Mobile-optimized settings
  latency?: number | { ideal: number; max: number };
  volume?: number | { ideal: number; min: number; max: number };
  
  // Advanced processing (Google-specific)
  googEchoCancellation?: boolean;
  googEchoCancellation2?: boolean;
  googAutoGainControl?: boolean;
  googAutoGainControl2?: boolean;
  googNoiseSuppression?: boolean;
  googNoiseSuppression2?: boolean;
  googHighpassFilter?: boolean;
  googTypingNoiseDetection?: boolean;
  googAudioMirroring?: boolean;
  
  // Experimental mobile features
  googAecExtendedFilter?: boolean;
  googAecm?: boolean;
  googAgcStartupMinVolume?: number;
  googDacTone?: boolean;
  googNoiseReduction?: boolean;
  googCpuOveruseDetection?: boolean;
  googCpuUnderuseThreshold?: number;
  googCpuOveruseThreshold?: number;
}

/**
 * Get optimal audio constraints for mobile WebRTC based on platform
 */
export const getOptimalAudioConstraints = (): OptimizedAudioConstraints => {
  const baseConstraints: OptimizedAudioConstraints = {
    // Core WebRTC processing - essential for mobile
    echoCancellation: true,
    noiseSuppression: true,
    autoGainControl: true,
    
    // Audio quality settings optimized for voice
    sampleRate: Platform.OS === 'ios' ? 48000 : 44100, // iOS handles 48kHz better
    sampleSize: 16, // 16-bit is optimal for voice
    channelCount: 1, // Mono for voice chat
    
    // Low latency for real-time communication
    latency: 0.01, // 10ms target
    volume: 1.0, // Maximum input volume
    
    // Advanced processing - Google WebRTC extensions
    googEchoCancellation: true,
    googEchoCancellation2: true,
    googAutoGainControl: true,
    googAutoGainControl2: true,
    googNoiseSuppression: true,
    googNoiseSuppression2: true,
    googHighpassFilter: true,
    googTypingNoiseDetection: true,
    googAudioMirroring: false,
    
    // Mobile-specific optimizations
    googAecExtendedFilter: true,
    googDacTone: false,
    googNoiseReduction: true,
    googCpuOveruseDetection: true,
    googCpuUnderuseThreshold: 55,
    googCpuOveruseThreshold: 85,
  };

  // Platform-specific optimizations
  if (Platform.OS === 'ios') {
    return {
      ...baseConstraints,
      // iOS-specific optimizations
      googAecm: false, // Disable mobile AEC for better quality on iOS
      googAgcStartupMinVolume: 15, // Higher startup volume for iOS
      sampleRate: 48000, // iOS handles 48kHz well
    };
  } else {
    return {
      ...baseConstraints,
      // Android-specific optimizations
      googAecm: true, // Enable mobile AEC for Android compatibility
      googAgcStartupMinVolume: 10, // Lower startup volume for Android
      sampleRate: 44100, // Android standard sample rate
    };
  }
};

/**
 * SDP manipulation for optimal Opus codec configuration
 */
export const optimizeOpusCodec = (sdp: string): string => {
  console.log("Optimizing SDP for Opus codec...");
  
  // Find Opus codec payload type
  const opusRegex = /a=rtpmap:(\d+)\s+opus\/48000(?:\/2)?/;
  const opusMatch = sdp.match(opusRegex);
  
  if (!opusMatch) {
    console.warn("Opus codec not found in SDP");
    return sdp;
  }
  
  const opusPayloadType = opusMatch[1];
  console.log(`Found Opus codec with payload type: ${opusPayloadType}`);
  
  // Remove any existing fmtp line for Opus
  sdp = sdp.replace(new RegExp(`a=fmtp:${opusPayloadType}.*\\r?\\n`, 'g'), '');
  
  // Optimal Opus parameters for mobile voice communication
  const opusParams = [
    `a=fmtp:${opusPayloadType} minptime=10; maxptime=60; sprop-maxcapturerate=48000; sprop-stereo=0; stereo=0; useinbandfec=1; usedtx=1; maxaveragebitrate=32000; maxplaybackrate=48000`,
  ].join('\\r\\n');
  
  // Insert optimized parameters after rtpmap line
  const rtpmapPattern = new RegExp(`(a=rtpmap:${opusPayloadType}\\s+opus\\/48000(?:\\/2)?\\r?\\n)`);
  sdp = sdp.replace(rtpmapPattern, `$1${opusParams}\\r\\n`);
  
  console.log("SDP optimized for Opus codec");
  return sdp;
};

/**
 * Audio quality metrics interface
 */
export interface AudioQualityMetrics {
  rtt: number; // Round trip time in ms
  packetLoss: number; // Packet loss percentage (0-1)
  jitter: number; // Jitter in ms
  audioLevel: number; // Audio level (0-1)
  bitrate: number; // Current bitrate in bps
}

/**
 * Parse WebRTC stats to extract audio quality metrics
 */
export const parseAudioStats = (stats: RTCStatsReport): AudioQualityMetrics => {
  const metrics: AudioQualityMetrics = {
    rtt: 0,
    packetLoss: 0,
    jitter: 0,
    audioLevel: 0,
    bitrate: 0,
  };

  stats.forEach((report) => {
    // Inbound audio stream stats
    if (report.type === 'inbound-rtp' && report.mediaType === 'audio') {
      const packetsReceived = report.packetsReceived || 0;
      const packetsLost = report.packetsLost || 0;
      
      if (packetsReceived > 0) {
        metrics.packetLoss = packetsLost / (packetsReceived + packetsLost);
      }
      metrics.jitter = report.jitter || 0;
      metrics.audioLevel = report.audioLevel || 0;
    }
    
    // Outbound audio stream stats
    if (report.type === 'outbound-rtp' && report.mediaType === 'audio') {
      const timestamp = report.timestamp || Date.now();
      const bytesSent = report.bytesSent || 0;
      
      if (timestamp > 0) {
        metrics.bitrate = (bytesSent * 8) / (timestamp / 1000); // bps
      }
    }
    
    // RTT from candidate pair
    if (report.type === 'candidate-pair' && report.state === 'succeeded') {
      metrics.rtt = (report.currentRoundTripTime || 0) * 1000; // Convert to ms
    }
  });

  return metrics;
};

/**
 * Determine audio quality level based on metrics
 */
export const getAudioQualityLevel = (metrics: AudioQualityMetrics): 'poor' | 'fair' | 'good' | 'excellent' => {
  const { rtt, packetLoss, jitter } = metrics;
  
  // Poor quality thresholds
  if (rtt > 300 || packetLoss > 0.05 || jitter > 50) {
    return 'poor';
  }
  
  // Fair quality thresholds
  if (rtt > 200 || packetLoss > 0.02 || jitter > 30) {
    return 'fair';
  }
  
  // Good quality thresholds
  if (rtt > 100 || packetLoss > 0.01 || jitter > 15) {
    return 'good';
  }
  
  return 'excellent';
};

/**
 * Log audio configuration for debugging
 */
export const logAudioConfiguration = (constraints: OptimizedAudioConstraints): void => {
  console.log("=== WebRTC Audio Configuration ===");
  console.log(`Platform: ${Platform.OS}`);
  console.log(`Echo Cancellation: ${constraints.echoCancellation}`);
  console.log(`Noise Suppression: ${constraints.noiseSuppression}`);
  console.log(`Auto Gain Control: ${constraints.autoGainControl}`);
  console.log(`Sample Rate: ${constraints.sampleRate}`);
  console.log(`Sample Size: ${constraints.sampleSize}`);
  console.log(`Channel Count: ${constraints.channelCount}`);
  console.log(`Latency: ${constraints.latency}ms`);
  console.log("===================================");
};
import { useRef, useCallback, useEffect } from "react";
import { Animated, Easing } from "react-native";
import { useFocusEffect } from "expo-router";

interface UseDashboardAnimationsProps {
  hasNoWordlist: boolean;
  isLoading: boolean;
  showCreateModal: boolean;
}

export const useDashboardAnimations = ({
  hasNoWordlist,
  isLoading,
  showCreateModal,
}: UseDashboardAnimationsProps) => {
  // Animation refs
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const slideAnim = useRef(new Animated.Value(50)).current;
  const pulseAnim = useRef(new Animated.Value(1)).current;
  const lockNudgeAnim = useRef(new Animated.Value(1)).current;
  const pulseLoopRef = useRef<Animated.CompositeAnimation | null>(null);

  const stopPulse = useCallback(() => {
    if (pulseLoopRef.current) {
      pulseLoopRef.current.stop();
      pulseLoopRef.current = null;
    }
    pulseAnim.stopAnimation();
    pulseAnim.setValue(1);
  }, [pulseAnim]);

  const startPulse = useCallback(() => {
    if (pulseLoopRef.current) return; // avoid duplicates
    const maxScale = hasNoWordlist ? 1.1 : 1.05;
    const loop = Animated.loop(
      Animated.sequence([
        Animated.timing(pulseAnim, {
          toValue: maxScale,
          duration: 1200,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
        Animated.timing(pulseAnim, {
          toValue: 1,
          duration: 1200,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
      ]),
    );
    pulseLoopRef.current = loop;
    loop.start();
  }, [hasNoWordlist, pulseAnim]);

  const triggerLockNudge = useCallback(() => {
    lockNudgeAnim.stopAnimation();
    lockNudgeAnim.setValue(1);
    Animated.sequence([
      Animated.timing(lockNudgeAnim, {
        toValue: 1.08,
        duration: 120,
        useNativeDriver: true,
      }),
      Animated.timing(lockNudgeAnim, {
        toValue: 1,
        duration: 160,
        useNativeDriver: true,
      }),
    ]).start();
  }, [lockNudgeAnim]);

  // Animate when wordlist state changes
  useEffect(() => {
    if (!isLoading) {
      fadeAnim.setValue(0);
      slideAnim.setValue(50);

      Animated.parallel([
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 400,
          useNativeDriver: true,
        }),
        Animated.timing(slideAnim, {
          toValue: 0,
          duration: 400,
          useNativeDriver: true,
        }),
      ]).start();
    }
  }, [hasNoWordlist, isLoading, fadeAnim, slideAnim]);

  // Start pulse when focused; stop on blur
  useFocusEffect(
    useCallback(() => {
      if (!isLoading) startPulse();
      return () => stopPulse();
    }, [isLoading, startPulse, stopPulse]),
  );

  // Restart pulse when empty state changes to update amplitude
  useEffect(() => {
    if (pulseLoopRef.current) {
      stopPulse();
      startPulse();
    }
  }, [hasNoWordlist, startPulse, stopPulse]);

  // Pause while the create modal is open
  useEffect(() => {
    if (showCreateModal) stopPulse();
    else startPulse();
  }, [showCreateModal, startPulse, stopPulse]);

  return {
    fadeAnim,
    slideAnim,
    pulseAnim,
    lockNudgeAnim,
    startPulse,
    stopPulse,
    triggerLockNudge,
  };
};

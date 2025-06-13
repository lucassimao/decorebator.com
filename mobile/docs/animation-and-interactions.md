# Animation & Interaction Patterns

## Overview

The Decorebator mobile app implements sophisticated animation and interaction patterns that enhance the learning experience through smooth, engaging, and accessible user interfaces. This document details the animation architectures, interaction patterns, and performance optimizations used throughout the application.

## Core Animation Principles

### 1. Performance-First Animation Strategy

All animations in the app prioritize performance and smoothness:

```typescript
// Always use native driver for transform animations
Animated.timing(animationValue, {
  toValue: 1,
  duration: 300,
  useNativeDriver: true, // Runs on UI thread (60fps)
});

// Non-native driver only for layout properties
Animated.timing(animationValue, {
  toValue: 100,
  duration: 300,
  useNativeDriver: false, // Required for width, height, etc.
});
```

### 2. Animation Lifecycle Management

```typescript
// Proper cleanup to prevent memory leaks
export function useAnimationCleanup() {
  const animationRef = useRef<Animated.CompositeAnimation | null>(null);
  
  useEffect(() => {
    return () => {
      // Stop animations on unmount
      animationRef.current?.stop();
    };
  }, []);
  
  const startAnimation = useCallback((animation: Animated.CompositeAnimation) => {
    animationRef.current = animation;
    animation.start();
  }, []);
  
  return { startAnimation };
}
```

## Complex Animation Systems

### 1. 3D Flashcard Flip Animation

The flashcard system implements sophisticated 3D flip animations with proper depth perception:

```typescript
export function FlashcardFlipAnimation() {
  const flipAnimation = useRef(new Animated.Value(0)).current;
  const [isFlipped, setIsFlipped] = useState(false);
  
  // Create rotation interpolations for 3D effect
  const frontInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["0deg", "180deg"],
  });
  
  const backInterpolate = flipAnimation.interpolate({
    inputRange: [0, 180],
    outputRange: ["180deg", "360deg"],
  });
  
  // Front card animation style
  const frontAnimatedStyle = {
    transform: [{ rotateY: frontInterpolate }],
  };
  
  // Back card animation style  
  const backAnimatedStyle = {
    transform: [{ rotateY: backInterpolate }],
  };
  
  // Flip function with physics-based animation
  const flipCard = useCallback(() => {
    Animated.timing(flipAnimation, {
      toValue: isFlipped ? 0 : 180,
      duration: 600,
      useNativeDriver: true,
    }).start();
    
    setIsFlipped(!isFlipped);
  }, [isFlipped, flipAnimation]);
  
  return (
    <View style={styles.cardContainer}>
      {/* Front of card */}
      <Animated.View
        style={[
          styles.card,
          frontAnimatedStyle,
          { backfaceVisibility: 'hidden' }
        ]}
      >
        <Text>Front Content</Text>
      </Animated.View>
      
      {/* Back of card */}
      <Animated.View
        style={[
          styles.card,
          styles.cardBack,
          backAnimatedStyle,
          { backfaceVisibility: 'hidden' }
        ]}
      >
        <Text>Back Content</Text>
      </Animated.View>
    </View>
  );
}
```

### 2. Multi-Layered Card Entry Animation

Cards enter with coordinated fade, scale, and slide animations:

```typescript
export function CardEntryAnimation({ children }: { children: React.ReactNode }) {
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.9)).current;
  const slideAnim = useRef(new Animated.Value(30)).current;
  
  useEffect(() => {
    // Parallel animations for smooth entry
    Animated.parallel([
      // Fade in
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 500,
        useNativeDriver: true,
      }),
      // Scale up with spring physics
      Animated.spring(scaleAnim, {
        toValue: 1,
        friction: 8,
        tension: 40,
        useNativeDriver: true,
      }),
      // Slide up
      Animated.timing(slideAnim, {
        toValue: 0,
        duration: 400,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);
  
  return (
    <Animated.View
      style={{
        opacity: fadeAnim,
        transform: [
          { scale: scaleAnim },
          { translateY: slideAnim }
        ],
      }}
    >
      {children}
    </Animated.View>
  );
}
```

### 3. Progressive Loading Animation

Staggered animations for loading multiple elements:

```typescript
export function ProgressiveLoadingAnimation({ items }: { items: any[] }) {
  const animValues = useRef(
    items.map(() => new Animated.Value(0))
  ).current;
  
  useEffect(() => {
    // Stagger animations for each item
    const animations = animValues.map((animValue, index) =>
      Animated.timing(animValue, {
        toValue: 1,
        duration: 300,
        delay: index * 100, // 100ms stagger
        useNativeDriver: true,
      })
    );
    
    Animated.stagger(100, animations).start();
  }, [items.length]);
  
  return (
    <View>
      {items.map((item, index) => (
        <Animated.View
          key={item.id}
          style={{
            opacity: animValues[index],
            transform: [{
              translateY: animValues[index].interpolate({
                inputRange: [0, 1],
                outputRange: [20, 0],
              })
            }]
          }}
        >
          <ItemComponent item={item} />
        </Animated.View>
      ))}
    </View>
  );
}
```

## Interaction Feedback Systems

### 1. Enhanced Touch Feedback

```typescript
export function TouchableCard({ children, onPress }: TouchableCardProps) {
  const scaleAnim = useRef(new Animated.Value(1)).current;
  const opacityAnim = useRef(new Animated.Value(1)).current;
  
  const handlePressIn = useCallback(() => {
    Animated.parallel([
      Animated.timing(scaleAnim, {
        toValue: 0.96,
        duration: 100,
        useNativeDriver: true,
      }),
      Animated.timing(opacityAnim, {
        toValue: 0.8,
        duration: 100,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);
  
  const handlePressOut = useCallback(() => {
    Animated.parallel([
      Animated.spring(scaleAnim, {
        toValue: 1,
        friction: 5,
        useNativeDriver: true,
      }),
      Animated.timing(opacityAnim, {
        toValue: 1,
        duration: 150,
        useNativeDriver: true,
      }),
    ]).start();
  }, []);
  
  return (
    <Pressable
      onPressIn={handlePressIn}
      onPressOut={handlePressOut}
      onPress={onPress}
    >
      <Animated.View
        style={{
          transform: [{ scale: scaleAnim }],
          opacity: opacityAnim,
        }}
      >
        {children}
      </Animated.View>
    </Pressable>
  );
}
```

### 2. Slide-Up Notification System

```typescript
export function SnackBar({ message, type, onDismiss }: SnackBarProps) {
  const slideAnim = useRef(new Animated.Value(100)).current;
  const opacityAnim = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    // Slide in from bottom
    Animated.parallel([
      Animated.timing(slideAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }),
      Animated.timing(opacityAnim, {
        toValue: 1,
        duration: 200,
        useNativeDriver: true,
      }),
    ]).start();
    
    // Auto dismiss after 3 seconds
    const timer = setTimeout(() => {
      dismissSnackbar();
    }, 3000);
    
    return () => clearTimeout(timer);
  }, []);
  
  const dismissSnackbar = useCallback(() => {
    Animated.parallel([
      Animated.timing(slideAnim, {
        toValue: 100,
        duration: 250,
        useNativeDriver: true,
      }),
      Animated.timing(opacityAnim, {
        toValue: 0,
        duration: 150,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onDismiss();
    });
  }, []);
  
  return (
    <Animated.View
      style={[
        styles.snackbar,
        styles[type],
        {
          opacity: opacityAnim,
          transform: [{ translateY: slideAnim }],
        },
      ]}
    >
      <Text style={styles.message}>{message}</Text>
      <TouchableOpacity onPress={dismissSnackbar}>
        <Text style={styles.dismissButton}>✕</Text>
      </TouchableOpacity>
    </Animated.View>
  );
}
```

### 3. Progress Animation System

```typescript
export function AnimatedProgressBar({ progress, duration = 1000 }: ProgressBarProps) {
  const progressAnim = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    Animated.timing(progressAnim, {
      toValue: progress,
      duration,
      useNativeDriver: false, // Layout property
    }).start();
  }, [progress]);
  
  return (
    <View style={styles.progressContainer}>
      <View style={styles.progressBackground}>
        <Animated.View
          style={[
            styles.progressFill,
            {
              width: progressAnim.interpolate({
                inputRange: [0, 100],
                outputRange: ['0%', '100%'],
                extrapolate: 'clamp',
              }),
            },
          ]}
        />
      </View>
    </View>
  );
}
```

## Gesture-Based Interactions

### 1. Swipe Navigation for Flashcards

```typescript
export function SwipeableFlashcard({ onSwipeLeft, onSwipeRight }: SwipeableProps) {
  const pan = useRef(new Animated.ValueXY()).current;
  const panGesture = useRef(
    PanGestureHandler.Activator.onGestureEvent(
      Animated.event(
        [{ nativeEvent: { translationX: pan.x, translationY: pan.y } }],
        { useNativeDriver: false }
      )
    )
  ).current;
  
  const onGestureStateChange = useCallback((event: PanGestureHandlerStateChangeEvent) => {
    if (event.nativeEvent.state === State.END) {
      const { translationX, velocityX } = event.nativeEvent;
      
      // Determine swipe direction and threshold
      if (Math.abs(translationX) > 100 || Math.abs(velocityX) > 500) {
        if (translationX > 0) {
          onSwipeRight?.();
        } else {
          onSwipeLeft?.();
        }
      }
      
      // Reset position
      Animated.spring(pan, {
        toValue: { x: 0, y: 0 },
        useNativeDriver: false,
      }).start();
    }
  }, [onSwipeLeft, onSwipeRight]);
  
  return (
    <PanGestureHandler
      onGestureEvent={panGesture}
      onHandlerStateChange={onGestureStateChange}
    >
      <Animated.View
        style={{
          transform: [
            { translateX: pan.x },
            {
              rotate: pan.x.interpolate({
                inputRange: [-200, 0, 200],
                outputRange: ['-15deg', '0deg', '15deg'],
                extrapolate: 'clamp',
              }),
            },
          ],
        }}
      >
        <FlashcardContent />
      </Animated.View>
    </PanGestureHandler>
  );
}
```

### 2. Pull-to-Refresh Implementation

```typescript
export function PullToRefreshList({ onRefresh, children }: PullToRefreshProps) {
  const [refreshing, setRefreshing] = useState(false);
  const pullAnim = useRef(new Animated.Value(0)).current;
  
  const handleRefresh = useCallback(async () => {
    setRefreshing(true);
    
    // Animate refresh indicator
    Animated.timing(pullAnim, {
      toValue: 1,
      duration: 200,
      useNativeDriver: true,
    }).start();
    
    try {
      await onRefresh();
    } finally {
      // Hide refresh indicator
      Animated.timing(pullAnim, {
        toValue: 0,
        duration: 300,
        useNativeDriver: true,
      }).start(() => {
        setRefreshing(false);
      });
    }
  }, [onRefresh]);
  
  return (
    <ScrollView
      refreshControl={
        <RefreshControl
          refreshing={refreshing}
          onRefresh={handleRefresh}
          tintColor="#FF7B54"
          colors={["#FF7B54"]}
        />
      }
    >
      {children}
    </ScrollView>
  );
}
```

## Loading State Animations

### 1. Skeleton Loading Animation

```typescript
export function SkeletonLoader({ width, height, borderRadius = 4 }: SkeletonProps) {
  const pulseAnim = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    const pulse = Animated.sequence([
      Animated.timing(pulseAnim, {
        toValue: 1,
        duration: 1000,
        useNativeDriver: true,
      }),
      Animated.timing(pulseAnim, {
        toValue: 0,
        duration: 1000,
        useNativeDriver: true,
      }),
    ]);
    
    Animated.loop(pulse).start();
  }, []);
  
  return (
    <Animated.View
      style={[
        {
          width,
          height,
          borderRadius,
          backgroundColor: '#F0F0F0',
          opacity: pulseAnim.interpolate({
            inputRange: [0, 1],
            outputRange: [0.3, 0.7],
          }),
        },
      ]}
    />
  );
}
```

### 2. Shimmer Loading Effect

```typescript
export function ShimmerPlaceholder({ width, height }: ShimmerProps) {
  const shimmerAnim = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    Animated.loop(
      Animated.timing(shimmerAnim, {
        toValue: 1,
        duration: 1500,
        useNativeDriver: true,
      })
    ).start();
  }, []);
  
  return (
    <View style={[styles.shimmerContainer, { width, height }]}>
      <Animated.View
        style={[
          styles.shimmerOverlay,
          {
            transform: [
              {
                translateX: shimmerAnim.interpolate({
                  inputRange: [0, 1],
                  outputRange: [-width, width],
                }),
              },
            ],
          },
        ]}
      />
    </View>
  );
}
```

## Accessibility and Interaction

### 1. Accessible Animations

```typescript
// Respect user's motion preferences
export function useReducedMotion() {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);
  
  useEffect(() => {
    const checkMotionPreference = async () => {
      try {
        // Check system accessibility settings
        const isReduceMotionEnabled = await AccessibilityInfo.isReduceMotionEnabled();
        setPrefersReducedMotion(isReduceMotionEnabled);
      } catch (error) {
        console.error('Error checking motion preference:', error);
      }
    };
    
    checkMotionPreference();
  }, []);
  
  return prefersReducedMotion;
}

// Conditional animation based on accessibility
export function AccessibleAnimation({ children }: { children: React.ReactNode }) {
  const prefersReducedMotion = useReducedMotion();
  const fadeAnim = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    if (prefersReducedMotion) {
      // Skip animation for accessibility
      fadeAnim.setValue(1);
    } else {
      // Normal animation
      Animated.timing(fadeAnim, {
        toValue: 1,
        duration: 300,
        useNativeDriver: true,
      }).start();
    }
  }, [prefersReducedMotion]);
  
  return (
    <Animated.View style={{ opacity: fadeAnim }}>
      {children}
    </Animated.View>
  );
}
```

### 2. Haptic Feedback Integration

```typescript
import { HapticFeedbackTypes, trigger } from 'react-native-haptic-feedback';

export function useHapticFeedback() {
  const triggerSuccess = useCallback(() => {
    trigger(HapticFeedbackTypes.notificationSuccess);
  }, []);
  
  const triggerError = useCallback(() => {
    trigger(HapticFeedbackTypes.notificationError);
  }, []);
  
  const triggerSelection = useCallback(() => {
    trigger(HapticFeedbackTypes.selection);
  }, []);
  
  return { triggerSuccess, triggerError, triggerSelection };
}

// Usage in quiz component
export function QuizAnswer({ isCorrect, onPress }: QuizAnswerProps) {
  const { triggerSuccess, triggerError } = useHapticFeedback();
  
  const handlePress = useCallback(() => {
    if (isCorrect) {
      triggerSuccess();
    } else {
      triggerError();
    }
    onPress();
  }, [isCorrect, onPress]);
  
  return (
    <TouchableOpacity onPress={handlePress}>
      <Text>Answer Option</Text>
    </TouchableOpacity>
  );
}
```

## Performance Optimization

### 1. Animation Performance Monitoring

```typescript
export function useAnimationPerformance() {
  const frameCount = useRef(0);
  const startTime = useRef(0);
  
  const startMonitoring = useCallback(() => {
    frameCount.current = 0;
    startTime.current = Date.now();
    
    const monitor = () => {
      frameCount.current++;
      requestAnimationFrame(monitor);
    };
    
    requestAnimationFrame(monitor);
  }, []);
  
  const getAverageFPS = useCallback(() => {
    const duration = Date.now() - startTime.current;
    return (frameCount.current / duration) * 1000;
  }, []);
  
  return { startMonitoring, getAverageFPS };
}
```

### 2. Memory-Efficient Animation Cleanup

```typescript
export function useAnimationCleanup() {
  const animationsRef = useRef<Animated.CompositeAnimation[]>([]);
  const timersRef = useRef<NodeJS.Timeout[]>([]);
  
  const addAnimation = useCallback((animation: Animated.CompositeAnimation) => {
    animationsRef.current.push(animation);
    return animation;
  }, []);
  
  const addTimer = useCallback((timer: NodeJS.Timeout) => {
    timersRef.current.push(timer);
    return timer;
  }, []);
  
  useEffect(() => {
    return () => {
      // Clean up all animations
      animationsRef.current.forEach(animation => {
        animation.stop();
      });
      
      // Clear all timers
      timersRef.current.forEach(timer => {
        clearTimeout(timer);
      });
      
      // Clear arrays
      animationsRef.current = [];
      timersRef.current = [];
    };
  }, []);
  
  return { addAnimation, addTimer };
}
```

This comprehensive animation and interaction system creates a polished, engaging user experience that enhances the educational content delivery while maintaining excellent performance and accessibility standards.
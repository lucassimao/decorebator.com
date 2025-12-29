# UI Design System & Component Architecture

## Overview

The Decorebator mobile app implements a comprehensive design system built for educational contexts, emphasizing warmth, accessibility, and user engagement. This document details the UI patterns, component architecture, and design decisions that create a cohesive learning experience.

## Color System

### Brand Color Palette

```typescript
export const Colors = {
  // Primary brand colors
  primary: "#FF7B54", // Brand orange - CTAs, emphasis
  success: "#4CAF50", // Green - achievements, correct answers
  error: "#FF6B6B", // Red - errors, incorrect answers
  warning: "#FFC107", // Yellow - cautions, alerts
  premium: "#FFD700", // Gold - premium features, achievements

  // Background system (warm, educational)
  background: {
    primary: "#FDF6E3", // Warm beige base
    light: "#FFF9F0", // Light gradient variant
    peach: "#FFE8D6", // Peachy accent
    orange: "#FFDCC3", // Orange tint
    sage: "#F5F0E6", // Sage accent
  },

  // Text hierarchy
  text: {
    dark: "#2D3436", // Primary text
    medium: "#636E72", // Secondary text
    light: "#B2BEC3", // Tertiary text, hints
  },

  // UI elements
  ui: {
    white: "#FFFFFF", // Pure white backgrounds
    divider: "#F0F0F0", // Subtle dividers
    border: "#E0E0E0", // Standard borders
    disabled: "#DFE6E9", // Disabled states
  },
};
```

### Color Usage Guidelines

- **Primary Orange (#FF7B54)**: CTAs, active states, brand emphasis
- **Warm Backgrounds**: Create comfortable, non-intimidating learning environment
- **Semantic Colors**: Green for success, red for errors, gold for achievements
- **Transparency System**: `rgba()` values for overlays and glass effects

## Typography Scale

### Font Hierarchy

```typescript
const Typography = {
  display: {
    size: 48,
    weight: "bold",
    usage: "Word display on flashcards",
  },
  heading: {
    size: 32,
    weight: "700",
    usage: "Quiz questions, main headings",
  },
  title: {
    size: 20,
    weight: "600",
    usage: "Section headers, card titles",
  },
  body: {
    size: 18,
    weight: "400",
    usage: "Primary content, definitions",
  },
  caption: {
    size: 16,
    weight: "500",
    usage: "Labels, metadata",
  },
  small: {
    size: 14,
    weight: "400",
    usage: "Hints, secondary info",
  },
  micro: {
    size: 12,
    weight: "500",
    usage: "Badges, tiny labels",
  },
};
```

### Font Implementation

- **Primary Font**: System default for optimal readability
- **Icon Font**: Expo Vector Icons (MaterialIcons, Ionicons)
- **Custom Font**: SpaceMono for monospace needs
- **Weight Variations**: 400 (regular), 500 (medium), 600 (semibold), 700 (bold)

## Layout System

### 8px Base Grid

All spacing follows an 8px base grid for visual consistency:

```typescript
const Spacing = {
  xs: 4, // Micro spacing
  sm: 8, // Base unit
  md: 16, // Standard spacing
  lg: 24, // Section spacing
  xl: 32, // Large gaps
  xxl: 40, // Screen margins
};
```

### Layout Principles

1. **Card-Based Design**: Primary content in elevated cards
2. **Consistent Spacing**: 8px increments throughout
3. **Touch-Friendly**: 44px minimum touch targets
4. **Visual Hierarchy**: Clear content organization
5. **Responsive**: Adapts to different screen sizes

## Shadow System

### Enhanced Shadow Configuration

```typescript
const Shadows = {
  // Card shadows
  card: {
    shadowColor: "#FF7B54", // Brand-colored shadows
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.15,
    shadowRadius: 20,
    elevation: 12, // Android
  },

  // Button shadows
  button: {
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 6,
  },

  // Icon container shadows
  icon: {
    shadowColor: "#FF7B54",
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2,
  },
};
```

### Visual Depth Strategy

- **Colored Shadows**: Use brand orange instead of black for warmth
- **Layered Depth**: Different elevation levels for visual hierarchy
- **Cross-Platform**: Consistent shadows on iOS and Android
- **Subtle Enhancement**: Adds polish without overwhelming content

## Component Architecture

### Modular Component System

#### 1. Analytics Components (7 Components)

```typescript
// Centralized theme configuration
export const analyticsTheme = {
  colors: Colors,
  spacing: Spacing,
  borderRadius: 16,
  chartHeight: 200,
};

// Individual components
-AnalyticsHeader - // Title and controls
  StatsGrid - // Metric overview
  WordMasteryChart - // Individual progress
  LearningProgressChart - // Daily trends
  QuizPerformanceChart - // Performance metrics
  BoxDistributionChart - // Leitner box distribution
  TopWordsSection; // High-performing words
```

#### 2. Quiz Components (Modular Interface)

```typescript
// Quiz rendering system
-QuizContent - // Dynamic content rendering (7 types)
  QuizOptions - // Answer selection interface
  QuizHeader - // Progress and navigation
  QuizModeToggle - // Fast mode toggle
  QuizNextButton - // Progression control
  QuizProgressBar - // Visual progress
  QuizLoadingState; // Loading with timeout
```

#### 3. Flashcard Components (4-Component System)

```typescript
// 3D flip animation system
-FlashcardContent - // Card rendering with animations
  FlashcardHeader - // Session controls
  FlashcardNavigation - // Card navigation
  FlashcardProgressBar; // Session progress
```

#### 4. Shared Components

```typescript
// Cross-feature components
-LoadingWithTimeout - // Configurable loading states
  ErrorState - // Type-based error display
  ErrorReportModal - // User feedback system
  OfflineIndicator; // Network status display
```

## Animation System

### 3D Flip Animation (Flashcards)

```typescript
// Complex rotation interpolation
const frontInterpolate = flipAnimation.interpolate({
  inputRange: [0, 180],
  outputRange: ["0deg", "180deg"],
});

const backInterpolate = flipAnimation.interpolate({
  inputRange: [0, 180],
  outputRange: ["180deg", "360deg"],
});

// Animation configuration
Animated.timing(flipAnimation, {
  toValue: isFlipped ? 180 : 0,
  duration: 600,
  useNativeDriver: true, // 60fps performance
});
```

### Performance-Optimized Animations

- **Native Driver**: All transform animations use native driver
- **Backface Visibility**: Proper 3D effect handling
- **Animation Cleanup**: Refs and listeners properly disposed
- **Parallel Animations**: Complex transitions with timing coordination

### Common Animation Patterns

1. **Fade + Scale**: Entry animations for cards
2. **Slide Transitions**: Page and modal transitions
3. **Spring Physics**: Natural feeling interactions
4. **Rotation Effects**: 3D card flips
5. **Progressive Loading**: Staggered content appearance

## Interactive Feedback

### Touch Feedback System

```typescript
// Standard touch feedback
<TouchableOpacity
  activeOpacity={0.7}
  style={styles.button}
>

// Enhanced feedback with scale
<Pressable
  onPressIn={() => scaleAnim.setValue(0.95)}
  onPressOut={() => scaleAnim.setValue(1.0)}
>
```

### Visual Feedback Patterns

- **Button States**: Opacity changes on press
- **Loading States**: Skeleton placeholders
- **Success/Error**: Color-coded feedback
- **Progress Indicators**: Real-time progress updates
- **Haptic Feedback**: iOS/Android haptic responses

## Accessibility Implementation

### Touch Target Guidelines

- **Minimum Size**: 44px × 44px for all interactive elements
- **Spacing**: 8px minimum between touch targets
- **Hit Slop**: Extended touch areas where needed

### Screen Reader Support

```typescript
// Semantic accessibility
<TouchableOpacity
  accessibilityLabel="Play pronunciation audio"
  accessibilityHint="Plays the pronunciation of the current word"
  accessibilityRole="button"
>
```

### Visual Accessibility

- **Color Contrast**: WCAG AA compliant text contrast
- **Text Scaling**: Supports system font scaling
- **Focus Indicators**: Clear focus states for navigation
- **Icon + Text**: Icons paired with descriptive text

## Responsive Design

### Breakpoint System (Enhanced January 2025)

The app now implements a comprehensive responsive design system:

```typescript
export const BREAKPOINTS = {
  SMALL: 359, // <= 359px width (iPhone SE, small Android)
  MEDIUM: 389, // 360-389px width (standard Android)
  LARGE: 390, // >= 390px width (modern iPhones, large Android)
} as const;
```

### Adaptive Design Patterns

#### 1. Responsive Utilities (`utils/responsive.ts`)

- **Dynamic Spacing**: Screen-size-aware padding, margins, and element spacing
- **Typography Scaling**: Font sizes adapt to screen categories
- **Touch Target Optimization**: Ensures minimum 44px touch targets
- **Keyboard Behavior**: Platform-specific keyboard avoidance strategies

#### 2. useResponsive Hook (`hooks/useResponsive.ts`)

Centralized responsive calculations with performance optimization:

```typescript
export const useResponsive = () => {
  const { width: screenWidth, height: screenHeight } = getScreenDimensions();

  return useMemo(
    () => ({
      screenWidth,
      screenHeight,
      category,
      spacing,
      fontSizes,
      imageConfig,
      keyboardBehavior,
      keyboardOffset,
    }),
    [screenWidth, screenHeight],
  );
};
```

**Performance Benefits**:

- Single computation per render cycle
- Memoization prevents unnecessary recalculations
- Centralized logic reduces code duplication

#### 3. Component-Level Responsiveness

```typescript
// Example: Responsive form styling
const styles = React.useMemo(
  () => createStyles(theme, responsive, keyboardVisible),
  [theme, responsive, keyboardVisible],
);
```

### Screen Size Adaptation

```typescript
// Enhanced responsive sizing
const getResponsiveSpacing = (width?: number) => {
  const category = getScreenSizeCategory(width);

  switch (category) {
    case "small":
      return { horizontal: 16, vertical: 12, buttonHeight: 48 };
    case "medium":
      return { horizontal: 20, vertical: 16, buttonHeight: 52 };
    case "large":
      return { horizontal: 24, vertical: 20, buttonHeight: 56 };
  }
};
```

### Orientation Support

- **Portrait Primary**: Optimized for portrait usage with responsive breakpoints
- **Landscape Support**: Graceful degradation in landscape with adaptive layouts
- **Tablet Support**: Enhanced layouts for larger screens with proper spacing
- **Cross-Platform**: Consistent responsive behavior on iOS and Android

## Performance Considerations

### Memory Management

```typescript
// Proper cleanup patterns
useEffect(() => {
  return () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    animationRef.current?.stop();
  };
}, []);
```

### Image Optimization

- **Progressive Loading**: Opacity transitions during load
- **Error Handling**: Retry mechanisms with fallbacks
- **Caching**: Intelligent image caching
- **Timeout Handling**: 10-second load timeouts

### Animation Performance

- **Native Driver**: Transform animations use native thread
- **Bounded Animations**: Prevent memory leaks
- **Cleanup**: Proper disposal of animation values
- **Debouncing**: Prevent rapid-fire interactions

### Component Performance Optimizations (January 2025)

#### 1. Memoization Strategy

```typescript
// Styles memoization
const styles = React.useMemo(
  () => createStyles(theme, responsive, keyboardVisible),
  [theme, responsive, keyboardVisible],
);

// Callback memoization
const onSubmit = React.useCallback(
  (data: FormData) => {
    // Form submission logic
  },
  [dependencies],
);
```

#### 2. Component Extraction

- **ErrorMessage Component**: Extracted reusable component with React.memo
- **Reduced Inline Components**: Eliminated inline component definitions
- **Proper Props Handling**: Type-safe props with proper memoization

#### 3. Performance Benefits

- **Reduced Re-renders**: Memoized components prevent unnecessary re-renders
- **Better Memory Usage**: Eliminated redundant calculations with useResponsive hook
- **Improved Responsiveness**: Native KeyboardAvoidingView replaces complex animations
- **Type Safety**: Proper TypeScript types prevent runtime errors

## Error State Design

### Error Type System

```typescript
enum ErrorType {
  Network = "network",
  Timeout = "timeout",
  Offline = "offline",
  NoData = "noData",
  General = "general",
}
```

### Error UI Patterns

- **Icon + Message**: Clear visual error communication
- **Retry Actions**: User-controlled error recovery
- **Context Awareness**: Different errors for different contexts
- **Graceful Degradation**: Partial functionality when possible

## Loading State Architecture

### Loading Strategy

1. **Immediate Feedback**: Instant loading indicators
2. **Timeout Detection**: 10-second timeout for slow networks
3. **Progressive Disclosure**: Delayed indicators (300ms) prevent flashing
4. **Skeleton Loading**: Content-shaped placeholders
5. **Retry Mechanisms**: User-controlled retry with exponential backoff

### Implementation Pattern

```typescript
const [loading, setLoading] = useState(false);
const [hasTimeout, setHasTimeout] = useState(false);

useEffect(() => {
  if (loading) {
    const timeout = setTimeout(() => {
      setHasTimeout(true);
    }, 10000);

    return () => clearTimeout(timeout);
  }
}, [loading]);
```

This comprehensive design system ensures consistent, accessible, and performant UI across the entire application while maintaining the warm, educational brand aesthetic that makes learning vocabulary engaging and approachable.

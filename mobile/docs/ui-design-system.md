# UI Design System & Component Architecture

## Overview

The Decorebator mobile app implements a comprehensive design system built for educational contexts, emphasizing warmth, accessibility, and user engagement. This document details the UI patterns, component architecture, and design decisions that create a cohesive learning experience.

## Color System

### Brand Color Palette

```typescript
export const Colors = {
  // Primary brand colors
  primary: "#FF7B54",        // Brand orange - CTAs, emphasis
  success: "#4CAF50",        // Green - achievements, correct answers
  error: "#FF6B6B",          // Red - errors, incorrect answers
  warning: "#FFC107",        // Yellow - cautions, alerts
  premium: "#FFD700",        // Gold - premium features, achievements
  
  // Background system (warm, educational)
  background: {
    primary: "#FDF6E3",      // Warm beige base
    light: "#FFF9F0",        // Light gradient variant
    peach: "#FFE8D6",        // Peachy accent
    orange: "#FFDCC3",       // Orange tint
    sage: "#F5F0E6",         // Sage accent
  },
  
  // Text hierarchy
  text: {
    dark: "#2D3436",         // Primary text
    medium: "#636E72",       // Secondary text
    light: "#B2BEC3",        // Tertiary text, hints
  },
  
  // UI elements
  ui: {
    white: "#FFFFFF",        // Pure white backgrounds
    divider: "#F0F0F0",      // Subtle dividers
    border: "#E0E0E0",       // Standard borders
    disabled: "#DFE6E9",     // Disabled states
  }
};
```

### Color Usage Guidelines

- **Primary Orange (#FF7B54)**: CTAs, active states, brand emphasis
- **Warm Backgrounds**: Create comfortable, non-intimidating learning environment
- **Semantic Colors**: Green for success, red for errors, gold for achievements
- **Transparency System**: `rgba()` values for overlays and glass effects

## Typography Scale

### Responsive Font Hierarchy

The typography system now automatically scales based on device type for optimal readability:

```typescript
// Base typography (Phone - 1x scaling)
const Typography = {
  display: {
    size: 40,      // Tablet 7": 44px, Tablet 10": 50px
    weight: "bold",
    usage: "Word display on flashcards"
  },
  headline: {
    size: 32,      // Tablet 7": 35px, Tablet 10": 40px
    weight: "700",
    usage: "Quiz questions, main headings"
  },
  title: {
    size: 24,      // Tablet 7": 26px, Tablet 10": 30px
    weight: "600", 
    usage: "Section headers, card titles"
  },
  body: {
    size: 16,      // Tablet 7": 18px, Tablet 10": 20px
    weight: "400",
    usage: "Primary content, definitions"
  },
  caption: {
    size: 12,      // Tablet 7": 13px, Tablet 10": 15px
    weight: "500",
    usage: "Labels, metadata"
  }
};

// Responsive multipliers applied automatically
const ResponsiveMultipliers = {
  phone: 1.0,     // Base scaling
  tablet7: 1.1,   // 10% larger for 7" tablets
  tablet10: 1.25  // 25% larger for 10" tablets
};
```

### Font Implementation

- **Primary Font**: System default for optimal readability
- **Icon Font**: Expo Vector Icons (MaterialIcons, Ionicons)
- **Custom Font**: SpaceMono for monospace needs
- **Weight Variations**: 400 (regular), 500 (medium), 600 (semibold), 700 (bold)

## Layout System

### Responsive Spacing Grid

All spacing follows a responsive 8px base grid that scales with device type:

```typescript
// Base spacing (Phone - 1x scaling)
const Spacing = {
  xs: 4,     // Tablet 7": 5px, Tablet 10": 6px
  sm: 8,     // Tablet 7": 10px, Tablet 10": 11px  
  md: 16,    // Tablet 7": 19px, Tablet 10": 22px
  lg: 24,    // Tablet 7": 29px, Tablet 10": 34px
  xl: 32,    // Tablet 7": 38px, Tablet 10": 45px
  xxl: 40,   // Tablet 7": 48px, Tablet 10": 56px
};

// Responsive multipliers
const SpacingMultipliers = {
  phone: 1.0,     // Base spacing
  tablet7: 1.2,   // 20% larger for 7" tablets
  tablet10: 1.4   // 40% larger for 10" tablets
};
```

### Responsive Layout Principles

1. **Card-Based Design**: Primary content in elevated cards with adaptive sizing
2. **Responsive Grids**: 1/2/3 column layouts based on device type
3. **Touch-Friendly**: Platform-appropriate touch targets (44pt iOS, 48dp Android)
4. **Content Constraints**: Maximum content widths for optimal readability
5. **Side-by-Side Layouts**: Efficient use of tablet screen real estate
6. **Visual Hierarchy**: Clear content organization that scales appropriately

### Device-Specific Layout Adaptations

```typescript
const LayoutAdaptations = {
  phone: {
    gridColumns: 1,
    maxContentWidth: "100%",
    touchTargets: { minimum: 44, comfortable: 48 }
  },
  tablet7: {
    gridColumns: 2,
    maxContentWidth: 680,
    touchTargets: { minimum: 48, comfortable: 56 }
  },
  tablet10: {
    gridColumns: 3,
    maxContentWidth: 800,
    touchTargets: { minimum: 52, comfortable: 60 }
  }
};
```

## Shadow System

### Enhanced Shadow Configuration

```typescript
const Shadows = {
  // Card shadows
  card: {
    shadowColor: "#FF7B54",      // Brand-colored shadows
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.15,
    shadowRadius: 20,
    elevation: 12,               // Android
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
  }
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
- AnalyticsHeader      // Title and controls
- StatsGrid           // Metric overview  
- WordMasteryChart    // Individual progress
- LearningProgressChart // Daily trends
- QuizPerformanceChart // Performance metrics
- BoxDistributionChart // Leitner box distribution
- TopWordsSection     // High-performing words
```

#### 2. Quiz Components (Modular Interface)

```typescript
// Quiz rendering system
- QuizContent         // Dynamic content rendering (7 types)
- QuizOptions         // Answer selection interface
- QuizHeader          // Progress and navigation
- QuizModeToggle      // Fast mode toggle
- QuizNextButton      // Progression control
- QuizProgressBar     // Visual progress
- QuizLoadingState    // Loading with timeout
```

#### 3. Flashcard Components (4-Component System)

```typescript
// 3D flip animation system
- FlashcardContent    // Card rendering with animations
- FlashcardHeader     // Session controls
- FlashcardNavigation // Card navigation
- FlashcardProgressBar // Session progress
```

#### 4. Shared Components

```typescript
// Cross-feature components
- LoadingWithTimeout  // Configurable loading states
- ErrorState          // Type-based error display
- ErrorReportModal    // User feedback system
- OfflineIndicator    // Network status display
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
  useNativeDriver: true,  // 60fps performance
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

## Responsive Design System

### Device Classification and Breakpoints

```typescript
// Device detection based on smaller screen dimension
export const DeviceBreakpoints = {
  phone: { min: 0, max: 767 },       // < 768px
  tablet7: { min: 768, max: 1023 },  // 768-1023px  
  tablet10: { min: 1024, max: 9999 } // ≥ 1024px
};

// Responsive hooks integration
const { isTablet, type, gridColumns, contentWidth } = useResponsive();
```

### Multi-Screen Layout Strategy

**Phone Layouts (< 768px)**:
- Single column content organization
- Full-width utilization 
- Vertical stacking for all components
- Standard touch targets (44pt minimum)

**Tablet 7" Layouts (768-1023px)**:
- 2-column grid systems for content cards
- Enhanced content areas with larger spacing
- Centered content with max-width constraints (680px)
- Larger touch targets (48pt minimum)

**Tablet 10" Layouts (≥ 1024px)**:
- 3-column grid systems for optimal content density
- Side-by-side layouts (content left, options right)
- Maximum content width constraints (800px)
- Largest touch targets (52pt minimum)

### Implementation Examples

```typescript
// Adaptive grid layouts
<FlatList
  numColumns={isTablet ? gridColumns : 1}
  columnWrapperStyle={isTablet ? styles.gridRow : undefined}
  data={items}
  renderItem={({ item }) => (
    <View style={isTablet ? commonStyles.gridItem : undefined}>
      <ItemComponent item={item} />
    </View>
  )}
/>

// Side-by-side tablet layouts
<View style={
  isTablet && deviceType === "tablet-10" 
    ? styles.tabletLayout 
    : styles.phoneLayout
}>
  <View style={styles.contentSection}>
    <ContentComponent />
  </View>
  <View style={styles.optionsSection}>
    <OptionsComponent />
  </View>
</View>

// Responsive containers with content constraints
<View style={commonStyles.responsiveContainer}>
  {/* Content automatically centers and constrains width */}
</View>
```

### Cross-Platform Orientation Support

- **Portrait Primary**: All layouts optimized for portrait usage
- **Landscape Compatibility**: Maintains device classification in landscape
- **Dynamic Adaptation**: Real-time layout updates on orientation change
- **Content Preservation**: Maintains optimal readability across orientations

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

## Error State Design

### Error Type System

```typescript
enum ErrorType {
  Network = "network",
  Timeout = "timeout", 
  Offline = "offline",
  NoData = "noData",
  General = "general"
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
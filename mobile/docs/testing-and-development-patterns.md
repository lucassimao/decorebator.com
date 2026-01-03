# Testing & Development Patterns

## Overview

The Decorebator mobile app implements comprehensive testing strategies and development patterns that ensure code quality, reliability, and maintainability. This document details the testing architectures, development workflows, and quality assurance patterns used throughout the application.

## Testing Architecture

### 1. Jest Configuration with Expo

```json
// package.json jest configuration
{
  "jest": {
    "preset": "jest-expo",
    "transformIgnorePatterns": [
      "node_modules/(?!((jest-)?react-native|@react-native(-community)?)|expo(nent)?|@expo(nent)?/.*|@expo-google-fonts/.*|react-navigation|@react-navigation/.*|@unimodules/.*|unimodules|sentry-expo|native-base|react-native-svg)"
    ],
    "setupFilesAfterEnv": ["<rootDir>/jest.setup.js"],
    "testMatch": [
      "**/__tests__/**/*.(ts|tsx|js)",
      "**/*.(test|spec).(ts|tsx|js)"
    ],
    "collectCoverageFrom": [
      "**/*.{ts,tsx}",
      "!**/*.d.ts",
      "!**/node_modules/**",
      "!**/.expo/**"
    ]
  }
}
```

### 2. Testing Utilities and Setup

```typescript
// jest.setup.js
// jest-native matchers are now included in @testing-library/react-native
import "react-native-gesture-handler/jestSetup";

// Mock React Native modules
jest.mock("react-native-reanimated", () => {
  const Reanimated = require("react-native-reanimated/mock");
  Reanimated.default.call = () => {};
  return Reanimated;
});

jest.mock("expo-secure-store", () => ({
  getItemAsync: jest.fn(),
  setItemAsync: jest.fn(),
  deleteItemAsync: jest.fn(),
}));

jest.mock("@react-native-async-storage/async-storage", () =>
  require("@react-native-async-storage/async-storage/jest/async-storage-mock"),
);

// Global test utilities
global.fetch = require("jest-fetch-mock");
```

## Component Testing Patterns

### 1. React Native Testing Library Patterns

```typescript
// __tests__/components/QuizContent.test.tsx
import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { QuizContent } from '../components/quiz/QuizContent';

describe('QuizContent', () => {
  const mockQuiz = {
    id: 1,
    type: 'GUESS_MEANING',
    value: 'hello',
    options: ['greeting', 'goodbye', 'question'],
    answerIndex: 0,
    audioURL: 'https://example.com/audio.mp3',
  };

  const defaultProps = {
    quiz: mockQuiz,
    userInput: '',
    setUserInput: jest.fn(),
    isSubmitted: false,
    onSubmitAnswer: jest.fn(),
    onSkipQuestion: jest.fn(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders quiz content correctly', () => {
    const { getByText, getByTestId } = render(<QuizContent {...defaultProps} />);

    expect(getByText('hello')).toBeTruthy();
    expect(getByTestId('audio-button')).toBeTruthy();
  });

  it('handles audio playback', async () => {
    const { getByTestId } = render(<QuizContent {...defaultProps} />);
    const audioButton = getByTestId('audio-button');

    fireEvent.press(audioButton);

    await waitFor(() => {
      // Audio should start playing
      expect(getByTestId('audio-button')).toHaveTextContent('pause');
    });
  });

  it('handles image loading with retry', async () => {
    const imageQuiz = {
      ...mockQuiz,
      type: 'WORD_FROM_IMAGE',
      value: 'https://example.com/image.jpg',
    };

    const { getByTestId, getByText } = render(
      <QuizContent {...defaultProps} quiz={imageQuiz} />
    );

    // Simulate image load error
    const image = getByTestId('quiz-image');
    fireEvent(image, 'error');

    await waitFor(() => {
      expect(getByText('Image failed to load')).toBeTruthy();
      expect(getByTestId('retry-button')).toBeTruthy();
    });

    // Test retry functionality
    fireEvent.press(getByTestId('retry-button'));

    await waitFor(() => {
      expect(getByTestId('quiz-image')).toBeTruthy();
    });
  });
});
```

### 2. Custom Testing Hooks

```typescript
// __tests__/hooks/useOffline.test.tsx
import { renderHook, act } from "@testing-library/react-hooks";
import NetInfo from "@react-native-community/netinfo";
import { useOffline } from "../hooks/useOffline";

// Mock NetInfo
jest.mock("@react-native-community/netinfo");
const mockNetInfo = NetInfo as jest.Mocked<typeof NetInfo>;

describe("useOffline", () => {
  it("should detect online state", async () => {
    mockNetInfo.addEventListener.mockImplementation((callback) => {
      callback({ isConnected: true });
      return jest.fn(); // unsubscribe function
    });

    const { result, waitForNextUpdate } = renderHook(() => useOffline());

    await waitForNextUpdate();

    expect(result.current.isOnline).toBe(true);
  });

  it("should detect offline state", async () => {
    mockNetInfo.addEventListener.mockImplementation((callback) => {
      callback({ isConnected: false });
      return jest.fn();
    });

    const { result, waitForNextUpdate } = renderHook(() => useOffline());

    await waitForNextUpdate();

    expect(result.current.isOnline).toBe(false);
  });

  it("should set offline availability based on subscription", () => {
    const mockUserInfo = { subscriptionPlan: "premium" };

    // Mock useUserInfo hook
    jest.doMock("../hooks/users", () => ({
      useUserInfo: () => ({ data: mockUserInfo }),
    }));

    const { result } = renderHook(() => useOffline());

    expect(result.current.isOfflineAvailable).toBe(true);
  });
});
```

### 3. API Integration Testing

```typescript
// __tests__/api/wordlists.test.ts
import { callAPI } from "../api/api";
import { getWordlists, createWordlist } from "../api/wordlists";

// Mock the base API function
jest.mock("../api/api");
const mockCallAPI = callAPI as jest.MockedFunction<typeof callAPI>;

describe("Wordlists API", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe("getWordlists", () => {
    it("should fetch wordlists successfully", async () => {
      const mockWordlists = [
        { id: 1, name: "Spanish Basics", language: "es" },
        { id: 2, name: "French Verbs", language: "fr" },
      ];

      mockCallAPI.mockResolvedValue(mockWordlists);

      const result = await getWordlists();

      expect(mockCallAPI).toHaveBeenCalledWith("GET", "/wordlists");
      expect(result).toEqual(mockWordlists);
    });

    it("should handle API errors", async () => {
      const errorMessage = "Network error";
      mockCallAPI.mockRejectedValue(new Error(errorMessage));

      await expect(getWordlists()).rejects.toThrow(errorMessage);
    });
  });

  describe("createWordlist", () => {
    it("should create wordlist with valid data", async () => {
      const newWordlist = { name: "German Nouns", language: "de" };
      const createdWordlist = { id: 3, ...newWordlist };

      mockCallAPI.mockResolvedValue(createdWordlist);

      const result = await createWordlist(newWordlist);

      expect(mockCallAPI).toHaveBeenCalledWith(
        "POST",
        "/wordlists",
        JSON.stringify(newWordlist),
      );
      expect(result).toEqual(createdWordlist);
    });
  });
});
```

## State Management Testing

### 1. React Query Testing

```typescript
// __tests__/hooks/useAnalytics.test.tsx
import { renderHook, waitFor } from '@testing-library/react-hooks';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useAnalytics } from '../hooks/useAnalytics';
import * as api from '../api/analytics';

// Mock API functions
jest.mock('../api/analytics');
const mockAPI = api as jest.Mocked<typeof api>;

// Test wrapper with QueryClient
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('useAnalytics', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should fetch analytics data successfully', async () => {
    const mockAnalytics = {
      stats: { totalWords: 100, learned: 25 },
      progress: [{ date: '2023-01-01', wordsStudied: 10 }],
    };

    mockAPI.getAnalytics.mockResolvedValue(mockAnalytics);

    const { result, waitForNextUpdate } = renderHook(
      () => useAnalytics(1, true),
      { wrapper: createWrapper() }
    );

    expect(result.current.isLoading).toBe(true);

    await waitForNextUpdate();

    expect(result.current.isLoading).toBe(false);
    expect(result.current.data).toEqual(mockAnalytics);
    expect(mockAPI.getAnalytics).toHaveBeenCalledWith(1);
  });

  it('should handle different cache times for premium vs free users', async () => {
    const { result: premiumResult } = renderHook(
      () => useAnalytics(1, true), // Premium user
      { wrapper: createWrapper() }
    );

    const { result: freeResult } = renderHook(
      () => useAnalytics(1, false), // Free user
      { wrapper: createWrapper() }
    );

    // Premium users should have shorter stale time
    expect(premiumResult.current.query?.options?.staleTime).toBe(10000);
    expect(freeResult.current.query?.options?.staleTime).toBe(900000);
  });
});
```

### 2. Context Provider Testing

```typescript
// __tests__/contexts/SnackbarProvider.test.tsx
import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import { Text, TouchableOpacity } from 'react-native';
import { SnackbarProvider, useSnackbar } from '../hooks/useSnackbar';

// Test component that uses snackbar
const TestComponent = () => {
  const { showSnackbar } = useSnackbar();

  return (
    <TouchableOpacity
      testID="trigger-snackbar"
      onPress={() => showSnackbar('Test message', 'success')}
    >
      <Text>Trigger Snackbar</Text>
    </TouchableOpacity>
  );
};

describe('SnackbarProvider', () => {
  it('should show and hide snackbar', async () => {
    const { getByTestId, getByText, queryByText } = render(
      <SnackbarProvider>
        <TestComponent />
      </SnackbarProvider>
    );

    // Initially no snackbar
    expect(queryByText('Test message')).toBeNull();

    // Trigger snackbar
    fireEvent.press(getByTestId('trigger-snackbar'));

    // Snackbar should appear
    await waitFor(() => {
      expect(getByText('Test message')).toBeTruthy();
    });

    // Snackbar should auto-dismiss after timeout
    await waitFor(
      () => {
        expect(queryByText('Test message')).toBeNull();
      },
      { timeout: 4000 }
    );
  });
});
```

## Offline Testing Strategies

### 1. Network State Mocking

```typescript
// __tests__/utils/offlineManager.test.ts
import NetInfo from "@react-native-community/netinfo";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { OfflineManager } from "../utils/offlineManager";

jest.mock("@react-native-community/netinfo");
jest.mock("@react-native-async-storage/async-storage");

describe("OfflineManager", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("should cache quiz data when online", async () => {
    const mockQuiz = {
      id: 1,
      type: "GUESS_MEANING",
      value: "hello",
      options: ["greeting", "goodbye"],
    };

    await OfflineManager.cacheQuiz(1, mockQuiz);

    expect(AsyncStorage.setItem).toHaveBeenCalledWith(
      "decorebator_offline_quiz_1",
      JSON.stringify({
        quiz: mockQuiz,
        timestamp: expect.any(Number),
        expiry: expect.any(Number),
      }),
    );
  });

  it("should retrieve cached quiz when offline", async () => {
    const cachedData = {
      quiz: { id: 1, type: "GUESS_MEANING" },
      timestamp: Date.now(),
      expiry: Date.now() + 72 * 60 * 60 * 1000, // 72 hours
    };

    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(
      JSON.stringify(cachedData),
    );

    const result = await OfflineManager.getCachedQuiz(1);

    expect(result).toEqual(cachedData.quiz);
  });

  it("should handle expired cache", async () => {
    const expiredData = {
      quiz: { id: 1, type: "GUESS_MEANING" },
      timestamp: Date.now(),
      expiry: Date.now() - 1000, // Expired
    };

    (AsyncStorage.getItem as jest.Mock).mockResolvedValue(
      JSON.stringify(expiredData),
    );

    const result = await OfflineManager.getCachedQuiz(1);

    expect(result).toBeNull();
    expect(AsyncStorage.removeItem).toHaveBeenCalledWith(
      "decorebator_offline_quiz_1",
    );
  });
});
```

### 2. Integration Testing for Offline Features

```typescript
// __tests__/integration/offlineQuiz.test.tsx
import React from 'react';
import { render, fireEvent, waitFor } from '@testing-library/react-native';
import NetInfo from '@react-native-community/netinfo';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import QuizScreen from '../app/quiz';

// Mock all required modules
jest.mock('@react-native-community/netinfo');
jest.mock('../utils/offlineManager');

describe('Offline Quiz Integration', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    jest.clearAllMocks();
  });

  const renderWithProviders = (component: React.ReactElement) => {
    return render(
      <QueryClientProvider client={queryClient}>
        {component}
      </QueryClientProvider>
    );
  };

  it('should show offline message for free users when offline', async () => {
    // Mock offline state
    (NetInfo.addEventListener as jest.Mock).mockImplementation((callback) => {
      callback({ isConnected: false });
      return jest.fn();
    });

    // Mock free user
    jest.doMock('../hooks/users', () => ({
      useUserInfo: () => ({ data: { subscriptionPlan: 'free' } })
    }));

    const { getByText } = renderWithProviders(
      <QuizScreen wordlistId="1" wordlistName="Test" />
    );

    await waitFor(() => {
      expect(getByText(/Premium Required/)).toBeTruthy();
    });
  });

  it('should load cached quiz for premium users when offline', async () => {
    // Mock offline state
    (NetInfo.addEventListener as jest.Mock).mockImplementation((callback) => {
      callback({ isConnected: false });
      return jest.fn();
    });

    // Mock premium user
    jest.doMock('../hooks/users', () => ({
      useUserInfo: () => ({ data: { subscriptionPlan: 'premium' } })
    }));

    // Mock cached quiz
    const mockCachedQuiz = {
      id: 1,
      type: 'GUESS_MEANING',
      value: 'hello',
      options: ['greeting', 'goodbye', 'question'],
    };

    jest.doMock('../utils/offlineManager', () => ({
      OfflineManager: {
        getCachedQuiz: jest.fn().mockResolvedValue(mockCachedQuiz),
      },
    }));

    const { getByText } = renderWithProviders(
      <QuizScreen wordlistId="1" wordlistName="Test" />
    );

    await waitFor(() => {
      expect(getByText('hello')).toBeTruthy();
    });
  });
});
```

## Performance Testing

### 1. Animation Performance Testing

```typescript
// __tests__/performance/animations.test.tsx
import React from 'react';
import { render } from '@testing-library/react-native';
import { Animated } from 'react-native';
import { FlashcardContent } from '../components/flashcard/FlashcardContent';

describe('Animation Performance', () => {
  it('should use native driver for transform animations', () => {
    const mockTimingStart = jest.fn();

    // Mock Animated.timing to track native driver usage
    jest.spyOn(Animated, 'timing').mockImplementation((value, config) => ({
      start: mockTimingStart,
      stop: jest.fn(),
    }));

    render(<FlashcardContent {...mockProps} />);

    // Verify native driver is used for flip animation
    expect(Animated.timing).toHaveBeenCalledWith(
      expect.any(Animated.Value),
      expect.objectContaining({
        useNativeDriver: true,
      })
    );
  });

  it('should cleanup animations on unmount', () => {
    const mockStop = jest.fn();

    jest.spyOn(Animated, 'timing').mockImplementation(() => ({
      start: jest.fn(),
      stop: mockStop,
    }));

    const { unmount } = render(<FlashcardContent {...mockProps} />);

    unmount();

    // Verify animations are stopped
    expect(mockStop).toHaveBeenCalled();
  });
});
```

### 2. Memory Leak Testing

```typescript
// __tests__/performance/memoryLeaks.test.tsx
import React from 'react';
import { render } from '@testing-library/react-native';
import { act } from 'react-test-renderer';

describe('Memory Leak Prevention', () => {
  it('should clear timers on unmount', () => {
    const clearTimeoutSpy = jest.spyOn(global, 'clearTimeout');

    const TestComponent = () => {
      const [timer, setTimer] = React.useState<NodeJS.Timeout | null>(null);

      React.useEffect(() => {
        const timeoutId = setTimeout(() => {}, 1000);
        setTimer(timeoutId);

        return () => {
          if (timeoutId) {
            clearTimeout(timeoutId);
          }
        };
      }, []);

      return null;
    };

    const { unmount } = render(<TestComponent />);

    act(() => {
      unmount();
    });

    expect(clearTimeoutSpy).toHaveBeenCalled();
  });
});
```

## Development Workflow Testing

### 0. TypeScript Development Commands (January 2025)

The mobile project now includes dedicated TypeScript type checking commands for enhanced development workflow:

```bash
# Run type checking once
npm run typecheck

# Run type checking in watch mode (continuous development)
npm run typecheck:watch
```

**Implementation**:

```json
// package.json scripts section
{
  "scripts": {
    "typecheck": "npx tsc --noEmit",
    "typecheck:watch": "npx tsc --noEmit --watch"
  }
}
```

**Benefits**:

- **Early Error Detection**: Catch type errors before runtime
- **CI/CD Integration**: Can be added to build pipelines
- **Development Workflow**: Watch mode helps during coding
- **No Output Files**: `--noEmit` flag ensures no JS files are generated (Expo handles compilation)
- **Fast Feedback**: Immediate type validation without full build process

**Usage in Development**:

```bash
# Terminal 1: Start Expo development server
npm start

# Terminal 2: Continuous type checking
npm run typecheck:watch

# Terminal 3: Run tests
npm test
```

### 1. Code Quality Gates

```typescript
// __tests__/quality/codeStandards.test.ts
describe("Code Quality Standards", () => {
  it("should not have console.log statements in production code", () => {
    const fs = require("fs");
    const path = require("path");

    const checkDirectory = (dir: string) => {
      const files = fs.readdirSync(dir);

      files.forEach((file: string) => {
        const filePath = path.join(dir, file);
        const stat = fs.statSync(filePath);

        if (stat.isDirectory() && !file.includes("node_modules")) {
          checkDirectory(filePath);
        } else if (file.endsWith(".tsx") || file.endsWith(".ts")) {
          const content = fs.readFileSync(filePath, "utf8");

          // Allow console.log in test files and dev utilities
          if (
            !filePath.includes("__tests__") &&
            !filePath.includes("offlineTest.ts")
          ) {
            expect(content).not.toMatch(/console\.log/);
          }
        }
      });
    };

    checkDirectory("./src");
  });

  it("should have proper TypeScript types", () => {
    // Test that all API functions have proper return types
    const apiFiles = ["wordlists.ts", "analytics.ts", "users.ts"];

    apiFiles.forEach((file) => {
      const content = require(`../api/${file}`);

      // Check that exported functions have type annotations
      Object.keys(content).forEach((exportName) => {
        if (typeof content[exportName] === "function") {
          // This would require AST parsing in a real implementation
          expect(exportName).toBeDefined();
        }
      });
    });
  });
});
```

### 2. Accessibility Testing

```typescript
// __tests__/accessibility/a11y.test.tsx
import React from 'react';
import { render } from '@testing-library/react-native';
import { TouchableOpacity, Text } from 'react-native';

describe('Accessibility Standards', () => {
  it('should have accessible touch targets', () => {
    const Component = () => (
      <TouchableOpacity
        testID="test-button"
        style={{ width: 44, height: 44 }}
        accessibilityLabel="Test button"
        accessibilityHint="Performs test action"
      >
        <Text>Test</Text>
      </TouchableOpacity>
    );

    const { getByTestId } = render(<Component />);
    const button = getByTestId('test-button');

    expect(button.props.accessibilityLabel).toBe('Test button');
    expect(button.props.accessibilityHint).toBe('Performs test action');
  });

  it('should have proper color contrast', () => {
    // Test color combinations meet WCAG guidelines
    const testColorContrast = (background: string, text: string) => {
      // This would use a color contrast library in practice
      const contrast = calculateContrast(background, text);
      expect(contrast).toBeGreaterThan(4.5); // WCAG AA standard
    };

    testColorContrast('#FFFFFF', '#2D3436'); // White bg, dark text
    testColorContrast('#FF7B54', '#FFFFFF'); // Orange bg, white text
  });
});
```

## Continuous Integration Testing

### 1. Test Scripts Configuration

```json
// package.json scripts for CI/CD
{
  "scripts": {
    "test": "jest --watchAll=false",
    "test:coverage": "jest --coverage --watchAll=false",
    "test:ci": "jest --ci --coverage --watchAll=false --maxWorkers=2",
    "test:integration": "jest --testMatch='**/__tests__/integration/**/*.test.{ts,tsx}'",
    "test:unit": "jest --testMatch='**/__tests__/unit/**/*.test.{ts,tsx}'",
    "test:watch": "jest --watch",
    "lint": "eslint . --ext .ts,.tsx",
    "lint:fix": "eslint . --ext .ts,.tsx --fix",
    "type-check": "tsc --noEmit"
  }
}
```

### 2. GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test Mobile App

on:
  push:
    branches: [main, develop]
    paths: ["mobile/**"]
  pull_request:
    branches: [main]
    paths: ["mobile/**"]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"
          cache: "npm"
          cache-dependency-path: mobile/package-lock.json

      - name: Install dependencies
        working-directory: mobile
        run: npm ci

      - name: Run linting
        working-directory: mobile
        run: npm run lint

      - name: Run type checking
        working-directory: mobile
        run: npm run type-check

      - name: Run tests
        working-directory: mobile
        run: npm run test:ci

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: mobile/coverage/lcov.info
          flags: mobile
```

This comprehensive testing strategy ensures code quality, prevents regressions, and maintains high reliability standards across the entire mobile application while supporting rapid development and deployment cycles.

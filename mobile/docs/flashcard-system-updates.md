# Flashcard System Updates (January 2025)

## Overview
The flashcard system has been significantly enhanced with route renaming, pronunciation display improvements, and internationalization updates to provide better semantic clarity and user experience.

## Major Changes

### 1. Route Rename: `/practice` → `/flashcard`

**Motivation**: The original `/practice` route was semantically unclear and didn't accurately represent the flashcard functionality.

**Changes Made**:
- **File Rename**: `app/practice.tsx` → `app/flashcard.tsx`
- **Navigation Update**: Updated `components/dashboard/WordlistItem.tsx` to use new route
- **Route Access**: Now accessible at `/flashcard?wordlistId=X&wordlistName=Y`
- **Backward Compatibility**: Old route no longer exists (breaking change)

**Benefits**:
- Clearer semantic meaning - users immediately understand it's flashcard functionality
- Better URL structure for bookmarking and sharing
- Consistent with component naming (`components/flashcard/`)
- Improved developer experience with self-documenting routes

### 2. Pronunciation Display Enhancement

**Problem**: Pronunciation was displayed on the back of flashcards with definitions, making it less accessible to learners.

**Solution**: Moved pronunciation to the front of flashcards for immediate visibility.

**Implementation Details**:
```tsx
// Front of card (NEW)
<Text style={styles.wordText}>{currentWord?.name}</Text>
{currentWord?.pronunciation && (
  <Text style={styles.phoneticTextFront}>
    /{currentWord?.pronunciation}/
  </Text>
)}

// Removed from back of card
// {currentWord?.pronunciation && (
//   <Text style={styles.phoneticText}>
//     /{currentWord?.pronunciation}/
//   </Text>
// )}
```

**New Styling**:
```tsx
phoneticTextFront: {
  fontSize: 18,
  color: theme.colors.text.secondary,
  fontStyle: "italic",
  marginTop: 8,
  marginBottom: 16,
  textAlign: "center",
},
```

**Benefits**:
- **Immediate Learning**: Users see pronunciation as soon as they see the word
- **Better Flow**: No need to flip card to know how to pronounce the word
- **Improved UX**: Follows language learning best practices
- **Cleaner Back**: Definition content is more focused and less cluttered

### 3. Internationalization Updates

**Updated Translation Key**: Changed from `"practice"` to `"flashcards"` across all languages.

**Languages Updated** (8 total):
- **English**: "Practice" → "Flashcards"
- **German**: "Üben" → "Lernkarten"
- **Spanish**: "Practicar" → "Tarjetas"
- **French**: "Pratiquer" → "Cartes" (was already correct)
- **Italian**: "Pratica" → "Carte"
- **Japanese**: "練習" → "フラッシュカード"
- **Portuguese (BR)**: "Praticar" → "Cartões"
- **Portuguese (PT)**: "Praticar" → "Cartões"

**Component Updates**:
```tsx
// Updated in WordlistItem.tsx
accessibilityLabel={t("wordlistItem.flashcards")}
{t("wordlistItem.flashcards")}
```

**Translation Files Updated**:
- `i18n/locales/en.json`
- `i18n/locales/de.json`
- `i18n/locales/es.json`
- `i18n/locales/fr.json`
- `i18n/locales/it.json`
- `i18n/locales/ja.json`
- `i18n/locales/pt-BR.json`
- `i18n/locales/pt-PT.json`

## Technical Implementation

### File Structure Changes
```diff
mobile/
├── app/
│   ├── quiz.tsx
-  │   ├── practice.tsx          # REMOVED
+  │   ├── flashcard.tsx         # NEW - renamed from practice.tsx
│   └── analytics.tsx
├── components/
│   ├── dashboard/
│   │   └── WordlistItem.tsx     # Updated navigation route
│   └── flashcard/              # Existing flashcard components
└── i18n/locales/               # All 8 language files updated
```

### Route Migration Guide

**For Users**:
- Old bookmarks to `/practice` will no longer work
- New route: `/flashcard?wordlistId=123&wordlistName=Example`

**For Developers**:
```tsx
// OLD
router.push(`/practice?wordlistId=${id}&wordlistName=${name}`);

// NEW
router.push(`/flashcard?wordlistId=${id}&wordlistName=${name}`);
```

### Component Interface

The flashcard system maintains the same component interface and functionality:

```tsx
interface FlashcardContentProps {
  currentWord: Word;
  definitions: Definition[];
  isFlipped: boolean;
  shouldFetchDefinitions: boolean;
  loadingDefinitions: boolean;
  definitionsError: any;
  slideAnimation: Animated.Value;
  scaleAnimation: Animated.Value;
  onFlip: () => void;
  onRefetchDefinitions: () => void;
}
```

## Quality Assurance

### Testing Completed
- ✅ **Linting**: All code passes ESLint checks
- ✅ **TypeScript**: No compilation errors with `npm run typecheck`
- ✅ **Route Navigation**: Confirmed route change works correctly
- ✅ **Translation Loading**: All 8 languages load correct translations
- ✅ **Component Functionality**: Flashcard flip animations and audio work properly
- ✅ **Pronunciation Display**: Front-of-card pronunciation displays correctly

### Breaking Changes
- **Route Change**: `/practice` → `/flashcard` (users need to update bookmarks)
- **Translation Keys**: Components using `t("wordlistItem.practice")` need updates

### Non-Breaking Changes
- All existing flashcard component functionality preserved
- Same query parameters and navigation flow
- Identical user interaction patterns

## Performance Impact

**Positive Impacts**:
- Better semantic caching with clearer route names
- Improved accessibility with pronunciation on front
- Reduced cognitive load with immediate pronunciation visibility

**No Performance Degradation**:
- Same component rendering performance
- Identical network request patterns
- Same memory usage patterns

## Future Considerations

### Enhancement Opportunities
1. **Audio Pronunciation**: Add play button next to pronunciation text
2. **Pronunciation Languages**: Support multiple pronunciation systems per language
3. **Custom Pronunciation**: Allow users to record custom pronunciations
4. **Pronunciation Highlighting**: Visual emphasis when audio plays

### Accessibility Improvements
1. **Screen Reader Support**: Enhanced pronunciation announcements
2. **Voice Navigation**: Voice commands for card navigation
3. **High Contrast**: Better pronunciation visibility in dark mode
4. **Font Scaling**: Responsive pronunciation text sizing

## Documentation References

This update affects the following documentation:
- `mobile-app-architecture.md` - Updated flashcard system section
- `offline-feature.md` - Updated to reflect flashcard route changes
- Component documentation in flashcard component files

## Commit Information

**Commit**: `946e83d - feat: Rename practice route to flashcard and move pronunciation to front`

**Files Changed**:
- `app/practice.tsx` → `app/flashcard.tsx` (renamed)
- `components/dashboard/WordlistItem.tsx` (route update)
- `components/flashcard/FlashcardContent.tsx` (pronunciation display)
- `i18n/locales/*.json` (8 language files updated)
- `package.json` (added typecheck command)

**Branch**: `fix/mobile-ui-issues`
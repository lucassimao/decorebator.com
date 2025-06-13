# Timezone & Date Handling in Analytics Charts

## Overview

This document describes the timezone and date handling improvements implemented across all analytics charts in the mobile app to ensure consistent and accurate date display regardless of user timezone.

## Problem Statement

The backend API sends date information as ISO timestamps in UTC format (e.g., `"2025-06-11T00:00:00Z"`). When JavaScript's native `Date` constructor parses these timestamps, it interprets them as UTC and converts them to the user's local timezone, causing date display issues.

**Example Issue:**
- Backend sends: `"2025-06-11T00:00:00Z"` (representing June 11th)
- User in UTC-3 timezone sees: June 10th (due to timezone conversion)
- Chart displays: "6/10" instead of correct "6/11"

## Solution Implementation

### Technical Approach

Instead of using `new Date(isoString)` which triggers timezone conversion, we now:

1. **Extract date part**: Split ISO timestamp on 'T' to get just the date portion
2. **Manual parsing**: Parse year, month, day components manually
3. **Local date creation**: Create Date object in local timezone using `new Date(year, month-1, day)`

### Code Pattern

```typescript
// ❌ Old approach (causes timezone issues)
const date = new Date("2025-06-11T00:00:00Z");

// ✅ New approach (timezone-safe)
const datePart = "2025-06-11T00:00:00Z".split('T')[0]; // "2025-06-11"
const [year, month, day] = datePart.split('-').map(Number); // [2025, 6, 11]
const date = new Date(year, month - 1, day); // Local timezone, correct date
```

## Components Updated

### 1. HistoricalBoxDistributionChart.tsx

**Location**: `mobile/components/analytics/HistoricalBoxDistributionChart.tsx`

**Changes**:
- Updated date label generation for chart x-axis
- Fixed "Today"/"Yesterday" detection logic
- Ensures correct MM/DD format display

```typescript
// Extract date part from ISO timestamp
const datePart = item.date.split('T')[0];
const [year, month, day] = datePart.split('-').map(Number);
const date = new Date(year, month - 1, day);
```

### 2. PracticeTimeChart.tsx

**Location**: `mobile/components/analytics/PracticeTimeChart.tsx`

**Changes**:
- Fixed `formatDateLabel` function for chart labels
- Updated date sorting logic for chronological ordering
- Added comprehensive error handling with fallbacks

```typescript
const formatDateLabel = (dateString: string): string => {
  try {
    const datePart = dateString.split('T')[0];
    const [year, month, day] = datePart.split('-').map(Number);
    
    if (!year || !month || !day || isNaN(year) || isNaN(month) || isNaN(day)) {
      // Fallback to UTC parsing if manual parsing fails
      const fallbackDate = new Date(dateString);
      return `${fallbackDate.getMonth() + 1}/${fallbackDate.getDate()}`;
    }
    
    const date = new Date(year, month - 1, day);
    return `${date.getMonth() + 1}/${date.getDate()}`;
  } catch (error) {
    console.error("Error parsing date:", dateString, error);
    return dateString;
  }
};
```

### 3. LearningProgressChart.tsx

**Location**: `mobile/components/analytics/LearningProgressChart.tsx`

**Changes**:
- Updated chart label generation for line chart x-axis
- Maintains consistent date display across all analytics

```typescript
learningProgress?.slice(-7).map((p) => {
  const datePart = p.date.split('T')[0];
  const [year, month, day] = datePart.split('-').map(Number);
  const date = new Date(year, month - 1, day);
  return `${date.getMonth() + 1}/${date.getDate()}`;
})
```

## Error Handling Strategy

### Graceful Fallbacks

All components include multiple fallback mechanisms:

1. **Validation checks**: Verify date parts are valid numbers
2. **UTC fallback**: Fall back to original UTC parsing if manual parsing fails
3. **Error logging**: Console errors for debugging while maintaining functionality
4. **Ultimate fallback**: Display original date string if all parsing fails

### Robust Date Sorting

For components that sort dates chronologically:

```typescript
.sort((a, b) => {
  const datePartA = a.date.split('T')[0];
  const datePartB = b.date.split('T')[0];
  const [yearA, monthA, dayA] = datePartA.split('-').map(Number);
  const [yearB, monthB, dayB] = datePartB.split('-').map(Number);
  const dateA = new Date(yearA, monthA - 1, dayA);
  const dateB = new Date(yearB, monthB - 1, dayB);
  
  if (isNaN(dateA.getTime()) || isNaN(dateB.getTime())) {
    console.error("Invalid date during sorting:", a.date, b.date);
    return 0; // Keep original order if dates are invalid
  }
  
  return dateA.getTime() - dateB.getTime();
})
```

## Testing Considerations

### Manual Testing

To verify the fix works correctly:

1. **Different Timezones**: Test app in various timezone settings
2. **Date Boundaries**: Test around midnight in different timezones
3. **Chart Accuracy**: Verify chart labels match expected dates
4. **Cross-Platform**: Test on both iOS and Android

### Test Cases

- User in UTC-3 (Brazil): Dates should display correctly without day offset
- User in UTC+9 (Japan): Dates should not be shifted forward
- Chart labels should show "Today" for current date, "Yesterday" for previous day
- Historical data should maintain chronological order

## Future Considerations

### Backend Standardization

Consider standardizing backend date format to avoid timezone ambiguity:
- Option 1: Send dates as "YYYY-MM-DD" strings without time component
- Option 2: Always include explicit timezone information
- Option 3: Send Unix timestamps with timezone offset

### Centralized Date Utility

For future maintenance, consider creating a centralized date parsing utility:

```typescript
// utils/dateUtils.ts
export const parseBackendDate = (dateString: string): Date => {
  const datePart = dateString.split('T')[0];
  const [year, month, day] = datePart.split('-').map(Number);
  return new Date(year, month - 1, day);
};

export const formatChartLabel = (dateString: string): string => {
  const date = parseBackendDate(dateString);
  return `${date.getMonth() + 1}/${date.getDate()}`;
};
```

## Impact

### User Experience

- **Accurate Analytics**: Users see correct dates in all analytics charts
- **Timezone Independence**: App works correctly regardless of user's timezone
- **Consistent Display**: All charts use the same date formatting approach

### Developer Experience

- **Predictable Behavior**: Date handling is now consistent across components
- **Better Debugging**: Error logging helps identify date parsing issues
- **Maintainable Code**: Clear pattern for handling backend dates

## Related Files

- `mobile/components/analytics/HistoricalBoxDistributionChart.tsx`
- `mobile/components/analytics/PracticeTimeChart.tsx`
- `mobile/components/analytics/LearningProgressChart.tsx`
- `mobile/hooks/useAnalytics.ts` (data fetching)
- `mobile/api/analytics.ts` (API types and calls)

## Performance Impact

The manual date parsing approach has minimal performance impact:
- String operations are lightweight
- Date object creation is the same cost
- No additional API calls or processing
- Error handling adds negligible overhead
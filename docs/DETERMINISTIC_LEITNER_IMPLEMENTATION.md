# Deterministic Leitner System Implementation

## Overview

This implementation replaces the probabilistic approach with a deterministic, priority-based selection algorithm that guarantees definition selection while respecting Leitner spaced repetition intervals.

## The Solution: Priority-Based Selection

Instead of probability calculations and random selection, definitions are selected based on their progress toward their target review interval using a deterministic priority system.

### Priority Buckets

| Priority | Progress Ratio | Description |
|----------|---------------|-------------|
| 1 | ≥ 100% | Overdue - Past target interval |
| 2 | 80-100% | Due Soon - Approaching target |
| 3 | 50-80% | Available - Mid-interval |
| 4 | < 50% | Early - Recently reviewed |

### Leitner Box Intervals

| Box | Target Interval | Description |
|-----|----------------|-------------|
| 1 | Immediate | Always 100% progress (highest priority) |
| 2 | 1 hour | Quick review cycle |
| 3 | 1 day (24h) | Daily review |
| 4 | 3 days (72h) | Every few days |
| 5 | 1 week (168h) | Weekly review |
| 6 | 2 weeks (336h) | Bi-weekly review |
| 7 | 1 month (720h) | Monthly review |

## Implementation Details

### Key Algorithm

1. **Calculate Progress Ratio**: `hours_elapsed / target_interval`
2. **Assign Priority**: Based on progress ratio thresholds
3. **Sort Deterministically**:
   - By priority bucket (1, 2, 3, 4)
   - By oldest reviewed timestamp (NULL first)
   - By definition ID (final tiebreaker)
4. **Select First**: Always returns the top definition

### SQL Query Structure

```sql
-- Calculate progress ratio for each definition
progress_ratio = hours_since_review / target_hours

-- Order by priority buckets
ORDER BY 
    CASE WHEN progress_ratio >= 1.0 THEN 0 ELSE 1 END,  -- Overdue
    CASE WHEN progress_ratio >= 0.8 THEN 0 ELSE 1 END,  -- Due soon
    CASE WHEN progress_ratio >= 0.5 THEN 0 ELSE 1 END,  -- Available
    COALESCE(updated_at, '1970-01-01') ASC,             -- Oldest first
    id ASC                                               -- Deterministic
```

## Examples

### Example 1: Mixed Box Scenario

**User has definitions in multiple boxes:**
- Definition A: Box 2, reviewed 2 hours ago (200% progress) → **Priority 1**
- Definition B: Box 4, reviewed 60 hours ago (83% progress) → **Priority 2**
- Definition C: Box 7, reviewed 360 hours ago (50% progress) → **Priority 3**
- Definition D: Box 7, reviewed 100 hours ago (14% progress) → **Priority 4**

**Result**: Definition A selected (overdue has highest priority)

### Example 2: All Box 7 Scenario

**All definitions in Box 7:**
- Definition A: Reviewed 600 hours ago (83% progress) → **Priority 2**
- Definition B: Reviewed 400 hours ago (56% progress) → **Priority 3**
- Definition C: Reviewed 200 hours ago (28% progress) → **Priority 4**
- Definition D: Reviewed 50 hours ago (7% progress) → **Priority 4**

**Result**: Definition A selected (highest progress ratio)

### Example 3: New Words Scenario

**Mix of new and existing:**
- Definition A: Box 1, NULL timestamp (100% progress) → **Priority 1**
- Definition B: Box 3, reviewed 12 hours ago (50% progress) → **Priority 3**
- Definition C: Box 5, reviewed 200 hours ago (119% progress) → **Priority 1**

**Result**: Between A and C (both Priority 1), A selected (NULL timestamp = oldest)

## Benefits Over Probabilistic Approach

1. **100% Deterministic**: Same data always produces same result
2. **No Randomness**: Eliminates intermittent "no words selected" errors
3. **Guaranteed Selection**: Always returns a definition if any exist
4. **Simpler Logic**: No complex probability calculations
5. **Easier Debugging**: Clear priority buckets and ordering
6. **Better Performance**: No RANDOM() calls or complex math
7. **Predictable Behavior**: Users see consistent progression

## Edge Case Handling

### All Definitions Recently Reviewed
- System picks the one with highest progress ratio
- Ensures continuous practice availability
- No stagnation possible

### Mix of Overdue and Fresh
- Overdue always takes precedence
- Maintains spaced repetition effectiveness

### New Definitions Added
- Box 1 definitions with NULL timestamp get highest priority
- Ensures new content is learned first

## Monitoring

The implementation logs each selection:

```go
common.Logger.Info("deterministic_selection",
    "userID", userID,
    "wordlistID", wordlistID,
    "definitionID", definition.ID,
    "boxID", result.BoxID,
    "hoursSinceReview", hoursSinceReview,
    "progressRatio", progressRatio,
    "wasOverdue", progressRatio >= 1.0)
```

## Migration from Probabilistic

The deterministic approach is a drop-in replacement requiring no schema changes:
- Same table structure
- Same function signatures
- Only the selection query changed
- Backwards compatible with existing data
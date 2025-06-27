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
| 2 | 6 hours | Quick review cycle |
| 3 | 1 day (24h) | Daily review |
| 4 | 3 days (72h) | Every few days |
| 5 | 1 week (168h) | Weekly review |
| 6 | 2 weeks (336h) | Bi-weekly review |
| 7 | 1 month (720h) | Monthly review |

## Implementation Details

### Key Algorithm

1. **Calculate Progress Ratio**: `hours_elapsed / target_interval`
2. **Assign Pure Priority Weight**: Based on simple progress ratio thresholds
3. **Sort by 3-Tier System**:
   - **Primary**: Highest weight wins (1000 > 800 > 500 > proportional)
   - **Secondary**: Oldest reviewed first (NULL timestamps = highest priority)
   - **Tertiary**: Definition ID ascending (deterministic final tiebreaker)
4. **Select First**: Always returns the top definition deterministically

### SQL Query Structure

```sql
-- Pure priority weight calculation
selection_weight = CASE 
    WHEN progress_ratio >= 1.0 THEN 1000                -- Overdue (highest priority)
    WHEN progress_ratio >= 0.8 THEN 800                 -- Due soon
    WHEN progress_ratio >= 0.5 THEN 500                 -- Available
    ELSE FLOOR(progress_ratio * 100)                    -- Early (proportional)
END

-- Clean 3-tier ordering (no complex calculations)
ORDER BY 
    selection_weight DESC,                              -- Primary: Highest weight wins
    COALESCE(updated_at, '1970-01-01') ASC,            -- Secondary: Oldest first
    id ASC                                             -- Tertiary: Deterministic
```

## Examples

### Example 1: Mixed Box Scenario

**User has definitions in multiple boxes:**
- Definition A: Box 2, reviewed 2 hours ago (200% progress) → **Weight = 1000** (overdue)
- Definition B: Box 4, reviewed 60 hours ago (83% progress) → **Weight = 800** (due soon)
- Definition C: Box 7, reviewed 360 hours ago (50% progress) → **Weight = 500** (available)
- Definition D: Box 7, reviewed 100 hours ago (14% progress) → **Weight = 14** (proportional)

**Result**: Definition A selected (highest weight = 1000)

### Example 2: All Box 7 Scenario

**All definitions in Box 7:**
- Definition A: Reviewed 600 hours ago (83% progress) → **Weight = 800** (due soon)
- Definition B: Reviewed 400 hours ago (56% progress) → **Weight = 500** (available)
- Definition C: Reviewed 200 hours ago (28% progress) → **Weight = 28** (proportional)
- Definition D: Reviewed 50 hours ago (7% progress) → **Weight = 7** (proportional)

**Result**: Definition A selected (highest weight = 800)

### Example 3: New Words Scenario

**Mix of new and existing:**
- Definition A: Box 1, NULL timestamp (100% progress) → **Weight = 1000** (new word)
- Definition B: Box 3, reviewed 12 hours ago (50% progress) → **Weight = 500** (available)
- Definition C: Box 5, reviewed 200 hours ago (119% progress) → **Weight = 1000** (overdue)

**Result**: Between A and C (both weight = 1000), A selected (NULL timestamp = oldest in tiebreaker)

## Benefits of Pure Priority Approach

1. **100% Deterministic**: Same data always produces same result
2. **No Artificial Randomness**: Respects spaced repetition science exactly
3. **Guaranteed Selection**: Always returns a definition if any exist
4. **Clean Priority Buckets**: Simple weight thresholds (1000, 800, 500, proportional)
5. **Fast Performance**: Minimal calculations, no complex math or trigonometry
6. **Easy Debugging**: Clear weight values and 3-tier tiebreaking
7. **Transparent Logic**: Users can understand and trust the system behavior
8. **Maintainable Code**: Simple algorithm that's easy to modify and test

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
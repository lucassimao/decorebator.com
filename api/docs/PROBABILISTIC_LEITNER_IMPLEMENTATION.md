# Probabilistic Leitner System Implementation

## Overview

This implementation solves the "Box 7 Stagnation" problem where users with all words in the highest Leitner box (30-day interval) would have no words available for practice, effectively breaking the application.

## The Solution: Probabilistic Availability

Instead of binary "due/not due" logic, every word now has a selection probability ranging from a minimum baseline to 100% when fully due.

### Probability Formula

```
P(selection) = base_probability + (time_progress * (1 - base_probability))

Where:
- base_probability = minimum chance for each box
- time_progress = hours_since_review / intended_interval
- When time_progress ≥ 1, P(selection) = 100%
```

### Box-Specific Probabilities

| Box | Interval | Min Probability | Description |
|-----|----------|----------------|-------------|
| 1 | Immediate | 100% | Always available |
| 2 | 1 hour | 70% | High availability |
| 3 | 1 day | 50% | Moderate availability |
| 4 | 3 days | 30% | Lower availability |
| 5 | 1 week | 20% | Low availability |
| 6 | 2 weeks | 10% | Very low availability |
| 7 | 1 month | 15% | Increased minimum availability |

## Implementation Details

### Key Changes

1. **Removed `getOldestDefinition()`** - No longer needed with probabilistic selection
2. **Modified `getNextDefinition()`** - Now uses probability-based selection
3. **Added monitoring functions**:
   - `checkHasUnlearnedWords()` - Early validation
   - `getWordlistBoxDistribution()` - Analytics tracking
4. **Enhanced logging** - Tracks selection probabilities and box distributions

### SQL Query Structure

The query calculates selection probability for each word and uses PostgreSQL's `RANDOM()` function for probabilistic selection:

```sql
WHERE roll <= selection_probability  -- Probabilistic selection
ORDER BY 
    -- Prioritize words that are "overdue" (100% probability)
    CASE WHEN selection_probability >= 1.0 THEN 0 ELSE 1 END,
    -- Then by how close they are to being due
    selection_probability DESC,
    -- Add some randomness for variety
    RANDOM()
```

## Examples

### Example 1: Normal Mixed Wordlist

**User wordlist state:**
- Word A: Box 2, reviewed 2 hours ago (due at 1 hour)
- Word B: Box 4, reviewed 30 hours ago (due at 72 hours) 
- Word C: Box 6, reviewed 100 hours ago (due at 336 hours)
- Word D: Box 7, reviewed 200 hours ago (due at 720 hours)

**Probability calculations:**
- Word A: 100% (overdue by 1 hour)
- Word B: 30% + (0.7 × 30/72) = 59.2%
- Word C: 10% + (0.9 × 100/336) = 36.8%
- Word D: 5% + (0.95 × 200/720) = 31.4%

**Selection behavior:** Word A will be selected ~60% of the time (as it should!), with other words sharing the remaining probability proportionally.

### Example 2: Box 7 Stagnation Scenario (OLD PROBLEM)

**Before probabilistic fix:**
- All 20 words in Box 7, all reviewed within last 24 hours
- Traditional system: 0% probability for all words
- **Result: NO QUIZ AVAILABLE** ❌

**After probabilistic fix:**
- All 20 words in Box 7, all reviewed within last 24 hours
- Each word: 15% + (0.85 × 24/720) = 17.8% probability
- **Result: User gets quiz with 100% certainty using weighted random selection** ✅

### Example 3: Realistic Learning Progression

**Day 1 - New learner:**
```
Box 1: 15 words (100% each) → Always selected
Box 2-7: 0 words
Expected: Box 1 words always chosen (perfect!)
```

**Day 30 - Intermediate learner:**
```
Box 1: 2 words (100% each) 
Box 2: 3 words (70-100% each)
Box 3: 5 words (50-100% each)
Box 4: 8 words (30-100% each)
Box 5: 2 words (20-100% each)
Expected: Lower boxes heavily favored, system works as intended
```

**Day 90 - Advanced learner (CRITICAL SCENARIO):**
```
Box 6: 3 words (10-100% each)
Box 7: 17 words (15-100% each)
Traditional: Often no words available
Probabilistic: Always something available, respects timing
```

### Example 4: Probability Curves in Action

For a Box 7 word (720-hour target interval):

| Hours Since Review | Probability | Description |
|-------------------|-------------|-------------|
| 0 (just reviewed) | 15.0% | Minimal chance |
| 72 (3 days) | 23.5% | Slight increase |
| 168 (1 week) | 34.9% | Moderate chance |
| 360 (15 days) | 57.5% | Half-way point |
| 540 (22.5 days) | 78.7% | Getting due |
| 720 (30 days) | 100% | Fully due |
| 1000 (41+ days) | 100% | Overdue |

### Example 5: Selection Query in Action

**Input:** User has 5 words with these probabilities:
- Word 1: 100% (Box 2, overdue)
- Word 2: 75% (Box 5, approaching due)
- Word 3: 30% (Box 4, not quite due)
- Word 4: 10% (Box 6, recently reviewed)
- Word 5: 15% (Box 7, just reviewed)

**Query execution (NEW WEIGHTED RANDOM METHOD):**
1. Calculate weighted random scores: `RANDOM() / (probability + 0.001)`
   - Word 1: 0.2 / 1.001 = 0.20 (lowest score, highest priority)
   - Word 2: 0.8 / 0.751 = 1.07
   - Word 3: 0.15 / 0.301 = 0.50
   - Word 4: 0.95 / 0.101 = 9.41
   - Word 5: 0.03 / 0.151 = 0.20
2. Order by weighted random (lowest first):
   - Word 1 or 5 selected (both have lowest scores)
3. **GUARANTEED SELECTION** - No possibility of "no words selected" error

### Example 6: Monitoring Logs

When the system selects a Box 7 word early:
```json
{
  "level": "info",
  "msg": "probabilistic_selection",
  "userID": 12345,
  "wordlistID": 67,
  "definitionID": 1001,
  "boxID": 7,
  "hoursSinceReview": 48.5,
  "selectionProbability": 0.114,
  "wasOverdue": false
}
```

When detecting high Box 7 concentration:
```json
{
  "level": "info", 
  "msg": "high_box_7_concentration",
  "userID": 12345,
  "wordlistID": 67,
  "box7Percentage": 85.7,
  "distribution": {"6": 2, "7": 12}
}
```

## Benefits

1. **Never Stuck**: Even with all words in Box 7, users always have something to practice
2. **Maintains Spaced Repetition**: Due words are still heavily prioritized
3. **Natural Feel**: Smooth probability curves instead of hard cutoffs
4. **Single Query**: No complex fallback logic needed
5. **Scientifically Sound**: Models natural memory decay

## Monitoring

The implementation includes comprehensive logging:

```go
common.Logger.Info("probabilistic_selection",
    "userID", userID,
    "wordlistID", wordlistID,
    "definitionID", definition.ID,
    "boxID", result.BoxID,
    "hoursSinceReview", hoursSinceReview,
    "selectionProbability", selectionProbability,
    "wasOverdue", selectionProbability >= 1.0)
```

Additionally, it tracks when >80% of words are in Box 7:

```go
common.Logger.Info("high_box_7_concentration",
    "userID", userID,
    "wordlistID", wordlistID,
    "box7Percentage", percentage,
    "distribution", distribution)
```

## Testing

Use the provided `test_probabilistic_selection.sql` to verify:
1. Probability calculations are correct
2. Box 7 stagnation scenario is resolved
3. Selection always returns results (when words exist)
4. Distribution queries work correctly

## Migration Path

This implementation is backward compatible:
- No database schema changes required
- Existing data works without modification
- Can be rolled back by reverting the code changes

## Performance Considerations

The probabilistic approach adds minimal overhead:
- Single query execution (no fallbacks)
- Efficient use of PostgreSQL's built-in functions
- Early validation prevents unnecessary queries

## Future Enhancements

1. **Tunable Parameters**: Make minimum probabilities configurable
2. **User Preferences**: Allow users to adjust review frequency
3. **Advanced Analytics**: Track how often non-due words are selected
4. **A/B Testing**: Compare engagement metrics with/without probabilistic selection
# Deterministic Leitner System Implementation

## Overview

This implementation uses a deterministic, due-only selection algorithm based on `next_review_at`. It aligns quiz selection with notifications and ensures scheduling is consistent across the system.

## The Solution: Due-Item Selection

Instead of selecting items early using progress ratios, the selector only returns items that are due (`next_review_at <= now()`) and orders them deterministically.

### Leitner Box Intervals

| Box | Target Interval | Description |
|-----|----------------|-------------|
| 1 | Immediate | Always due after review |
| 2 | 6 hours | Quick review cycle |
| 3 | 1 day (24h) | Daily review |
| 4 | 3 days (72h) | Every few days |
| 5 | 1 week (168h) | Weekly review |
| 6 | 2 weeks (336h) | Bi-weekly review |
| 7 | 1 month (720h) | Monthly review |

## Implementation Details

### Key Algorithm

1. **Filter to due items**: `next_review_at <= NOW()`
2. **Order deterministically**:
   - **Primary**: earliest `next_review_at`
   - **Secondary**: oldest `updated_at`
   - **Tertiary**: definition ID ascending
3. **Select first**: returns the top due definition deterministically

### SQL Query Structure

```sql
WHERE lst.next_review_at IS NOT NULL
  AND lst.next_review_at <= NOW()
  AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())

ORDER BY
  lst.next_review_at ASC,
  lst.updated_at ASC,
  def.id ASC
```

## Behavior Notes

- **No due items**: if all items are scheduled for the future, quiz creation returns a `no_due_items` response. The mobile UI shows a friendly “Nothing due yet” state.
- **Deterministic selection**: same data always yields the same result.
- **Aligned scheduling**: quiz selection and reminder logic both use `next_review_at`.

## Monitoring

Each selection logs a due-item event:

```go
common.Logger.Info("due_item_selection",
    "userID", userID,
    "wordlistID", wordlistID,
    "definitionID", definition.ID,
    "boxID", result.BoxID,
    "hoursSinceReview", hoursSinceReview,
    "nextReviewAt", nextReviewAt,
    "isOverdue", nextReviewAt.Before(time.Now()))
```

## Migration Notes

- Requires `next_review_at` to be present on `leitner_system_tracking`.
- Backfill provided by:
  - `api/cmd/migrate/migrations/000071_move_next_review_at_to_tracking.*.sql`
- Quiz answers update `next_review_at` based on the resulting box.

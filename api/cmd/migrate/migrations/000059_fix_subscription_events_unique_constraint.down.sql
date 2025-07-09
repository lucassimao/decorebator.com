-- Revert the unique constraint fix for subscription_events table

-- Drop the new constraint
ALTER TABLE subscription_events 
DROP CONSTRAINT IF EXISTS subscription_events_external_event_id_key;

-- Restore the old constraint (this may fail if the column no longer exists)
-- This is intentionally left as a comment since reverting this migration
-- would require also reverting migration 000056, which is not recommended
-- ALTER TABLE subscription_events 
-- ADD CONSTRAINT subscription_events_stripe_event_id_key UNIQUE (stripe_event_id);
-- Fix unique constraint for subscription_events table
-- The constraint was created with the old column name but wasn't properly renamed in migration 000056

-- Drop the old constraint that references the old column name
ALTER TABLE subscription_events 
DROP CONSTRAINT IF EXISTS subscription_events_stripe_event_id_key;

-- Add the new constraint with the correct column name
ALTER TABLE subscription_events 
ADD CONSTRAINT subscription_events_external_event_id_key UNIQUE (external_event_id);
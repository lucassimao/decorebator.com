-- Rollback subscription_events table provider-agnostic changes

-- Remove provider column
ALTER TABLE subscription_events 
DROP COLUMN provider;

-- Rename external_event_id back to stripe_event_id
ALTER TABLE subscription_events 
RENAME COLUMN external_event_id TO stripe_event_id;

-- Update the index back to original name
DROP INDEX IF EXISTS idx_subscription_events_external_event_id;
CREATE INDEX idx_subscription_events_stripe_event_id ON subscription_events(stripe_event_id);

-- Drop the provider enum type if no other tables use it
DROP TYPE IF EXISTS subscription_provider;
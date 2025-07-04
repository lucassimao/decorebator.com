-- Update subscription_events table to be provider-agnostic

-- First, create the provider enum if it doesn't exist
DO $$ BEGIN
    CREATE TYPE subscription_provider AS ENUM ('stripe', 'revenuecat');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add new provider column with default value
ALTER TABLE subscription_events 
ADD COLUMN provider subscription_provider DEFAULT 'stripe';

-- Rename stripe_event_id to external_event_id for provider agnosticism
ALTER TABLE subscription_events 
RENAME COLUMN stripe_event_id TO external_event_id;

-- Update the index to use the new column name
DROP INDEX IF EXISTS idx_subscription_events_stripe_event_id;
CREATE INDEX idx_subscription_events_external_event_id ON subscription_events(external_event_id);

-- Make provider column NOT NULL after setting defaults
ALTER TABLE subscription_events 
ALTER COLUMN provider SET NOT NULL;

-- Update existing Stripe records to have the correct provider value (already defaulted to 'stripe')
-- No UPDATE needed since default is 'stripe'
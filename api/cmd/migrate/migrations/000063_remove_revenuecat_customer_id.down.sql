-- Add back revenuecat_customer_id column and index
-- This restores the column in case rollback is needed

-- Add the column back
ALTER TABLE users ADD COLUMN revenuecat_customer_id VARCHAR(255) UNIQUE DEFAULT NULL;

-- Recreate the index
CREATE INDEX idx_users_revenuecat_customer_id ON users(revenuecat_customer_id);
-- Remove obsolete revenuecat_customer_id column and index
-- The RevenueCat integration now uses app_user_id directly as user ID

-- Drop the index first
DROP INDEX IF EXISTS idx_users_revenuecat_customer_id;

-- Drop the column
ALTER TABLE users DROP COLUMN IF EXISTS revenuecat_customer_id;
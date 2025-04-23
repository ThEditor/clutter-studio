DROP INDEX IF EXISTS idx_verification_codes_user_id;

DROP TABLE IF EXISTS VerificationCodes;

ALTER TABLE Users
DROP COLUMN IF EXISTS email_verified;

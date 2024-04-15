DROP TABLE IF EXISTS "verify_emails" CASCADE;

ALTER TABLE "user" DROP COLUMN "is_email_verified";
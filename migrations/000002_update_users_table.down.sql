-- Drop indexes first
DROP INDEX IF EXISTS idx_documents_type;
DROP INDEX IF EXISTS idx_documents_loan_id;
DROP INDEX IF EXISTS idx_loans_status;
DROP INDEX IF EXISTS idx_loans_user_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS loans;

-- Remove columns from users table
ALTER TABLE users
    DROP COLUMN IF EXISTS monthly_income,
    DROP COLUMN IF EXISTS ktp_number,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS phone_number,
    DROP COLUMN IF EXISTS status;

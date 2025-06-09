CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    address TEXT NOT NULL,
    ktp_number VARCHAR(16) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    monthly_income DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TYPE loan_status AS ENUM (
    'pending', 'in_review', 'approved', 'rejected', 
    'disbursed', 'paid_off', 'defaulted', 'restructured'
);

CREATE TYPE document_status AS ENUM (
    'required', 'uploaded', 'verified', 'rejected'
);

CREATE TYPE document_type AS ENUM (
    'ktp', 'payslip', 'bank_statement', 'employee_letter'
);

CREATE TABLE loans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    amount DECIMAL(15,2) NOT NULL,
    tenure_months INTEGER NOT NULL,
    purpose TEXT NOT NULL,
    status loan_status NOT NULL DEFAULT 'pending',
    monthly_payment DECIMAL(15,2),
    interest_rate DECIMAL(5,2) NOT NULL,
    disbursed_amount DECIMAL(15,2),
    disbursed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT chk_amount CHECK (amount >= 1000000),
    CONSTRAINT chk_tenure CHECK (tenure_months BETWEEN 6 AND 60)
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    loan_id UUID NOT NULL REFERENCES loans(id),
    type document_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    status document_status NOT NULL DEFAULT 'required',
    url TEXT,
    uploaded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

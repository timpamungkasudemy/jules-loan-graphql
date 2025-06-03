CREATE TYPE IF NOT EXISTS collateral_categories AS ENUM ('CAR', 'MOTORCYCLE');

CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY,
    full_name TEXT NOT NULL,
    date_of_birth DATE NOT NULL,
    id_number VARCHAR(25) NOT NULL UNIQUE,
    email TEXT,
    phone VARCHAR(30) NOT NULL,
    address_street VARCHAR(200) NOT NULL,
    address_city VARCHAR(100) NOT NULL,
    address_zipcode VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS loans (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id),
    tenure INTEGER NOT NULL,
    amount FLOAT NOT NULL,
    loan_status TEXT NOT NULL,
    collateral_category collateral_categories NOT NULL,
    collateral_brand TEXT NOT NULL,
    collateral_variant TEXT NOT NULL,
    collateral_manufacturing_year INTEGER NOT NULL,
    collateral_is_document_complete BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ
);

-- This script runs when PostgreSQL container starts for the first time
-- The database 'product_views' is already created by POSTGRES_DB environment variable

-- Connect to the product_views database
\c product_views;

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    view_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster top N queries
CREATE INDEX IF NOT EXISTS idx_products_view_count ON products(view_count DESC);

-- Create function to update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for updated_at
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create goose migration tracking table
CREATE TABLE IF NOT EXISTS goose_db_version (
    id SERIAL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    tstamp TIMESTAMP DEFAULT NOW()
);

-- Mark initial migration as applied
INSERT INTO goose_db_version (version_id, is_applied) 
VALUES (1, true) 
ON CONFLICT DO NOTHING;

-- Create products table if not exists
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    view_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for efficient top N queries
CREATE INDEX IF NOT EXISTS idx_products_view_count ON products(view_count DESC);

-- Clear existing data
TRUNCATE TABLE products;

-- Insert sample products with varying view counts
INSERT INTO products (id, name, description, view_count) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'iPhone 15 Pro', 'Latest Apple smartphone with titanium design', 95000),
    ('550e8400-e29b-41d4-a716-446655440002', 'Samsung Galaxy S24', 'Flagship Android phone with AI features', 87000),
    ('550e8400-e29b-41d4-a716-446655440003', 'MacBook Pro M3', 'Professional laptop with M3 chip', 82000),
    ('550e8400-e29b-41d4-a716-446655440004', 'iPad Pro', 'Powerful tablet for professionals', 75000),
    ('550e8400-e29b-41d4-a716-446655440005', 'Sony WH-1000XM5', 'Premium noise-cancelling headphones', 68000),
    ('550e8400-e29b-41d4-a716-446655440006', 'AirPods Pro', 'Wireless earbuds with active noise cancellation', 65000),
    ('550e8400-e29b-41d4-a716-446655440007', 'Dell XPS 15', 'High-performance Windows laptop', 58000),
    ('550e8400-e29b-41d4-a716-446655440008', 'Nintendo Switch OLED', 'Hybrid gaming console', 55000),
    ('550e8400-e29b-41d4-a716-446655440009', 'PlayStation 5', 'Next-gen gaming console', 52000),
    ('550e8400-e29b-41d4-a716-446655440010', 'Xbox Series X', 'Microsoft gaming console', 48000),
    ('550e8400-e29b-41d4-a716-446655440011', 'Apple Watch Ultra', 'Rugged smartwatch for extreme sports', 45000),
    ('550e8400-e29b-41d4-a716-446655440012', 'Samsung QLED TV', '65-inch 4K Smart TV', 42000),
    ('550e8400-e29b-41d4-a716-446655440013', 'Bose QuietComfort', 'Comfortable noise-cancelling headphones', 38000),
    ('550e8400-e29b-41d4-a716-446655440014', 'GoPro Hero 12', 'Action camera for adventures', 35000),
    ('550e8400-e29b-41d4-a716-446655440015', 'Kindle Oasis', 'Premium e-reader with warm light', 32000),
    ('550e8400-e29b-41d4-a716-446655440016', 'Dyson V15', 'Cordless vacuum cleaner', 28000),
    ('550e8400-e29b-41d4-a716-446655440017', 'Instant Pot Pro', 'Multi-use pressure cooker', 25000),
    ('550e8400-e29b-41d4-a716-446655440018', 'Fitbit Charge 6', 'Fitness tracker with heart rate monitor', 22000),
    ('550e8400-e29b-41d4-a716-446655440019', 'Ring Video Doorbell', 'Smart doorbell with camera', 18000),
    ('550e8400-e29b-41d4-a716-446655440020', 'Echo Dot', 'Smart speaker with Alexa', 15000);

-- Generate additional random products to simulate large dataset
INSERT INTO products (name, description, view_count)
SELECT 
    'Product ' || generate_series,
    'Description for product ' || generate_series,
    floor(random() * 10000)::bigint
FROM generate_series(21, 1000);

-- Update statistics
ANALYZE products;

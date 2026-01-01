-- Rollback Migration: Remove PostgreSQL Full-Text Search for Products
-- Date: 2026-01-01
-- Phase: 2 - Product Catalog & Cart System
-- REQ-6: PostgreSQL Full-Text Search (Rollback)

-- This rollback script safely removes all full-text search components
-- in reverse order of creation to maintain referential integrity.

-- ========================================
-- Step 1: Drop GIN index
-- ========================================
-- Remove the GIN index first (dependent on column)
DROP INDEX IF EXISTS products_search_vector_idx;

-- ========================================
-- Step 2: Drop trigger
-- ========================================
-- Remove the trigger (dependent on trigger function)
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;

-- ========================================
-- Step 3: Drop trigger function
-- ========================================
-- Remove the trigger function (no longer needed)
DROP FUNCTION IF EXISTS products_search_vector_update();

-- ========================================
-- Step 4: Remove search_vector column
-- ========================================
-- Finally, remove the search_vector column from products table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'search_vector'
    ) THEN
        ALTER TABLE products DROP COLUMN search_vector;
    END IF;
END $$;

-- ========================================
-- Rollback Complete
-- ========================================
-- All full-text search components have been removed.
-- The products table has been restored to its previous state.
--
-- NOTE: This rollback does NOT affect product data (name, description, etc.)
-- It only removes the search optimization infrastructure.
--
-- After rollback, you can still search products using LIKE queries:
--   SELECT * FROM products WHERE name ILIKE '%search term%' OR description ILIKE '%search term%'
--
-- However, LIKE queries are significantly slower than full-text search
-- on large datasets (50k+ products).

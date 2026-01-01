#!/bin/bash

# Test script for PostgreSQL Full-Text Search migration
# This script applies the migration, tests it, and optionally rolls it back

set -e  # Exit on error

# Configuration (override with environment variables)
PGHOST=${PGHOST:-localhost}
PGPORT=${PGPORT:-5432}
PGUSER=${PGUSER:-admin}
PGDATABASE=${PGDATABASE:-go-test}
PGPASSWORD=${PGPASSWORD:-adminadmin}

export PGHOST PGPORT PGUSER PGDATABASE PGPASSWORD

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if psql is available
if ! command -v psql &> /dev/null; then
    print_error "psql command not found. Please install PostgreSQL client."
    exit 1
fi

# Check database connection
print_header "Testing Database Connection"
if psql -c "SELECT version();" &> /dev/null; then
    print_success "Connected to database: $PGDATABASE@$PGHOST:$PGPORT"
else
    print_error "Failed to connect to database. Check your credentials."
    exit 1
fi

# Function to check if migration is applied
check_migration_status() {
    local column_exists=$(psql -tAc "SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'search_vector');")
    local index_exists=$(psql -tAc "SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'products' AND indexname = 'products_search_vector_idx');")
    local trigger_exists=$(psql -tAc "SELECT EXISTS (SELECT 1 FROM information_schema.triggers WHERE trigger_name = 'products_search_vector_trigger');")

    if [ "$column_exists" = "t" ] && [ "$index_exists" = "t" ] && [ "$trigger_exists" = "t" ]; then
        return 0  # Migration is applied
    else
        return 1  # Migration is not applied
    fi
}

# Main menu
print_header "PostgreSQL Full-Text Search Migration Test"
echo "Database: $PGDATABASE@$PGHOST:$PGPORT"
echo "User: $PGUSER"
echo ""

if check_migration_status; then
    print_warning "Migration is currently APPLIED"
    echo ""
    echo "Options:"
    echo "  1) Run tests only"
    echo "  2) Rollback migration"
    echo "  3) Re-apply migration (rollback + apply)"
    echo "  4) Exit"
else
    print_warning "Migration is currently NOT APPLIED"
    echo ""
    echo "Options:"
    echo "  1) Apply migration"
    echo "  2) Exit"
fi

read -p "Choose an option: " choice

case $choice in
    1)
        if check_migration_status; then
            # Migration already applied, run tests
            print_header "Running Tests"

            # Test 1: Check column exists
            print_header "Test 1: Check search_vector column"
            psql -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'search_vector';"
            print_success "Column exists"

            # Test 2: Check index exists
            print_header "Test 2: Check GIN index"
            psql -c "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'products' AND indexname = 'products_search_vector_idx';"
            print_success "Index exists"

            # Test 3: Check trigger exists
            print_header "Test 3: Check trigger"
            psql -c "SELECT trigger_name, event_manipulation FROM information_schema.triggers WHERE trigger_name = 'products_search_vector_trigger';"
            print_success "Trigger exists"

            # Test 4: Test search functionality
            print_header "Test 4: Test full-text search"
            psql -c "SELECT id, name FROM products WHERE search_vector @@ websearch_to_tsquery('english', 'product') LIMIT 5;"
            print_success "Search works"

            # Test 5: Test trigger auto-update
            print_header "Test 5: Test trigger auto-update"
            psql -c "INSERT INTO products (id, name, description, slug, base_price, status) VALUES (gen_random_uuid(), 'Test Cotton Shirt', 'A beautiful cotton shirt for testing', 'test-cotton-shirt', 100000, 'draft') RETURNING id, name, search_vector IS NOT NULL AS search_vector_populated;"
            print_success "Trigger auto-populates search_vector on INSERT"

            # Cleanup test data
            psql -c "DELETE FROM products WHERE slug = 'test-cotton-shirt';" &> /dev/null

            print_header "All Tests Passed!"
        else
            # Apply migration first
            print_header "Applying Migration"
            psql -f /Volumes/External/Work/personal/go-yippi/migrations/20260101_add_product_search.sql
            print_success "Migration applied successfully"

            # Run tests
            $0  # Restart script to run tests
        fi
        ;;
    2)
        if check_migration_status; then
            # Rollback
            print_header "Rolling Back Migration"
            psql -f /Volumes/External/Work/personal/go-yippi/migrations/20260101_add_product_search_down.sql
            print_success "Migration rolled back successfully"

            # Verify rollback
            if check_migration_status; then
                print_error "Rollback verification failed - components still exist"
                exit 1
            else
                print_success "Rollback verified - all components removed"
            fi
        else
            # Exit
            print_header "Exiting"
            exit 0
        fi
        ;;
    3)
        # Re-apply (rollback + apply)
        print_header "Re-applying Migration (Rollback + Apply)"

        print_header "Step 1: Rollback"
        psql -f /Volumes/External/Work/personal/go-yippi/migrations/20260101_add_product_search_down.sql
        print_success "Rollback complete"

        print_header "Step 2: Apply"
        psql -f /Volumes/External/Work/personal/go-yippi/migrations/20260101_add_product_search.sql
        print_success "Migration re-applied successfully"

        # Run tests
        $0  # Restart script to run tests
        ;;
    4|*)
        print_header "Exiting"
        exit 0
        ;;
esac

print_header "Done!"

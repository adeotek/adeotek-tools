#!/bin/bash

# Integration test script for sql-toolbox

set -e

echo "Running integration tests for sql-toolbox..."

# Clean up any existing test artifacts
rm -f test-integration.db test-integration2.db

# Test 1: Basic SQLite migration
echo "Test 1: Basic SQLite migration"
./build/sql-toolbox migration --target-path ./test-scripts --provider sqlite --database test-integration.db --verbose

# Verify database was created and has expected tables
echo "Verifying database structure..."
TABLES=$(sqlite3 test-integration.db "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name;")
EXPECTED_TABLES="__migrations_history
posts
users"

if [ "$TABLES" = "$EXPECTED_TABLES" ]; then
    echo "✓ Database tables created correctly"
else
    echo "✗ Database tables not created correctly"
    echo "Expected: $EXPECTED_TABLES"
    echo "Got: $TABLES"
    exit 1
fi

# Test 2: Re-run migration (should skip all scripts)
echo "Test 2: Re-run migration (should skip all scripts)"
OUTPUT=$(./build/sql-toolbox migration --target-path ./test-scripts --provider sqlite --database test-integration.db --verbose 2>&1)

if echo "$OUTPUT" | grep -q "Skipped: 5"; then
    echo "✓ Scripts correctly skipped on re-run"
else
    echo "✗ Scripts not skipped correctly"
    echo "Output: $OUTPUT"
    exit 1
fi

# Test 3: Dry run mode
echo "Test 3: Dry run mode"
OUTPUT=$(./build/sql-toolbox migration --target-path ./test-scripts --provider sqlite --database test-integration2.db --dry-run --verbose 2>&1)

if echo "$OUTPUT" | grep -q "Dry run: would execute script"; then
    echo "✓ Dry run mode working correctly"
else
    echo "✗ Dry run mode not working correctly"
    echo "Output: $OUTPUT"
    exit 1
fi

# Verify dry run didn't create database
if [ -f "test-integration2.db" ]; then
    echo "✗ Dry run created database file (should not have)"
    exit 1
else
    echo "✓ Dry run correctly didn't create database"
fi

# Test 4: Version flag
echo "Test 4: Version flag"
OUTPUT=$(./build/sql-toolbox --version 2>&1)

if echo "$OUTPUT" | grep -q "sql-toolbox version"; then
    echo "✓ Version flag working correctly"
else
    echo "✗ Version flag not working correctly"
    echo "Output: $OUTPUT"
    exit 1
fi

# Test 5: Environment variables
echo "Test 5: Environment variables"
export SQLTB_PROVIDER=sqlite
export SQLTB_DATABASE=test-integration-env.db
export SQLTB_VERBOSE=true

OUTPUT=$(./build/sql-toolbox migration -t ./test-scripts 2>&1)

if echo "$OUTPUT" | grep -q "Success: 5"; then
    echo "✓ Environment variables working correctly"
else
    echo "✗ Environment variables not working correctly"
    echo "Output: $OUTPUT"
    exit 1
fi

unset SQLTB_PROVIDER
unset SQLTB_DATABASE
unset SQLTB_VERBOSE

# Clean up
rm -f test-integration.db test-integration-env.db

echo "All integration tests passed! ✓"

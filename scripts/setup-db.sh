#!/bin/bash
# Setup Postgres database for go-foundation development.
# Idempotent — safe to re-run.
#
# Usage: ./scripts/setup-db.sh

set -euo pipefail

DB_NAME="go_foundation"
DB_USER="go_foundation"
DB_PASS="devpass"

echo "Setting up database '$DB_NAME' with user '$DB_USER'..."

# Check if already exists
if sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" 2>/dev/null | grep -q 1; then
    echo "  Database $DB_NAME already exists, skipping CREATE"
else
    sudo -u postgres psql -c "CREATE DATABASE $DB_NAME;"
    echo "  Database $DB_NAME created"
fi

if sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" 2>/dev/null | grep -q 1; then
    echo "  Role $DB_USER already exists, skipping CREATE"
else
    sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASS';"
    echo "  Role $DB_USER created"
fi

sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"
sudo -u postgres psql -d $DB_NAME -c "GRANT ALL ON SCHEMA public TO $DB_USER;"

echo ""
echo "Done. Connection string:"
echo "  export DATABASE_URL=\"postgres://$DB_USER:$DB_PASS@localhost:5432/$DB_NAME?sslmode=disable\""
echo ""
echo "Verify with:"
echo "  go run ./cmd/pgping"

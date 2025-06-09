$env:PGPASSWORD = "root"

# Create test database
psql -U postgres -h localhost -p 5432 -c "DROP DATABASE IF EXISTS xyz_multifinance_test;"
psql -U postgres -h localhost -p 5432 -c "CREATE DATABASE xyz_multifinance_test;"

# Run migrations on test database
$env:DATABASE_URL = "postgresql://postgres:root@localhost:5432/xyz_multifinance_test?sslmode=disable"
migrate -path "../migrations" -database $env:DATABASE_URL up

Write-Host "Test database setup completed successfully"

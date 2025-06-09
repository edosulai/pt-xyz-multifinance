#!/bin/sh
set -e

echo "Waiting for database to be ready..."
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
        echo "Database is ready!"
        break
    fi
    echo "Waiting for database... (Attempt: $attempt/$max_attempts)"
    attempt=$((attempt + 1))
    sleep 2
done

if [ $attempt -gt $max_attempts ]; then
    echo "Failed to connect to database after $max_attempts attempts!"
    exit 1
fi

echo "Running database migrations..."
migrate -path /app/migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE" up

echo "Starting the application..."
exec /app/main

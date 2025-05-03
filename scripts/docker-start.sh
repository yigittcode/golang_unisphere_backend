#!/bin/sh

# Script to start the application in a Docker environment
# Useful for handling initialization operations before the main app starts

set -e

echo "Starting UniSphere API..."

# Wait for PostgreSQL to be ready (retry for 60 seconds)
echo "Waiting for PostgreSQL..."
for i in $(seq 1 60); do
  if nc -z postgres 5432; then
    echo "PostgreSQL is up!"
    break
  fi
  
  echo "Attempt $i: PostgreSQL not ready yet, waiting..."
  sleep 1
  
  if [ $i -eq 60 ]; then
    echo "PostgreSQL did not become available in time. Proceeding anyway..."
  fi
done

# Run the application
echo "Starting the application..."
exec ./unisphere
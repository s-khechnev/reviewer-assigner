#!/bin/sh
set -e

echo "Running database migrations..."
make migrate_up

echo "Starting application..."
exec ./main

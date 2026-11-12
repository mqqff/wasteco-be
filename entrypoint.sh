#!/bin/sh
set -e

if [ "$1" != "" ] && [ "$1" != "server" ]; then
    exec "$@"
fi

echo "============================================"
echo " Starting WasteCo Backend Services"
echo "============================================"

DB_TARGET_HOST="${DB_HOST:-mysql}"
DB_TARGET_PORT="${DB_PORT:-3306}"

echo "Waiting for MySQL database at ${DB_TARGET_HOST}:${DB_TARGET_PORT}..."
while ! nc -z -w 2 "${DB_TARGET_HOST}" "${DB_TARGET_PORT}"; do
    sleep 1
done
echo "MySQL database is reachable and ready!"

if [ "${AUTO_MIGRATE:-true}" = "true" ]; then
    echo "Running database schema migrations..."
    /app/migrate
fi

if [ "${AUTO_SEED:-true}" = "true" ]; then
    echo "Running database seeder..."
    /app/seed
fi

echo "Starting HTTP Server on ${APP_HOST:-0.0.0.0}:${APP_PORT:-8000}..."
exec /app/wasteco

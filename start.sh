#!/bin/sh


set -e

echo "run db migration"
source /app/.app.env
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up
echo "statr the app"
#execute any command passed as arguments to the script
exec "$@"
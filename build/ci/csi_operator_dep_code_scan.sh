#!/bin/bash

OWASPDC_DIRECTORY=$1
DATA_DIRECTORY="$OWASPDC_DIRECTORY/data"
REPORT_DIRECTORY="$OWASPDC_DIRECTORY/reports"
CACHE_DIRECTORY="$OWASPDC_DIRECTORY/data/cache"
EXCLUDE=$2

if [ ! -d "$DATA_DIRECTORY" ]; then
    echo "Initially creating persistent directory: $DATA_DIRECTORY"
    mkdir -p "$DATA_DIRECTORY"
fi
if [ ! -d "$REPORT_DIRECTORY" ]; then
    echo "Initially creating persistent directory: $REPORT_DIRECTORY"
    mkdir -p "$REPORT_DIRECTORY"
fi
if [ ! -d "$CACHE_DIRECTORY" ]; then
    echo "Initially creating persistent directory: $CACHE_DIRECTORY"
    mkdir -p "$CACHE_DIRECTORY"
fi

chmod -R 777 "$OWASPDC_DIRECTORY"

# Make sure we are using the latest version
docker pull owasp/dependency-check

docker run --rm \
    --volume go.mod:/src \
    --volume "$DATA_DIRECTORY":/usr/share/dependency-check/data \
    --volume "$REPORT_DIRECTORY":/report \
    owasp/dependency-check \
    --scan /src \
    --exclude "${EXCLUDE}" \
    --format "ALL" \
    --project "OWASP Dependency Check Operator Dependencies" \
    --out /report \
    --log /report/logs


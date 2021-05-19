#!/bin/bash

set -e
INPUT_ENV=$1

for i in 3 2 1 0
  do
    nc -z "$POSTGRES_HOST" 5432 && break || echo "Waiting for PostgreSQL to get ready..." && sleep $i
  done

if [ "$INPUT_ENV" == "test" ]
  then
    echo -e '[psql]\nhost = "'"${POSTGRES_HOST}"'"\nname = "test"\nusername = "test"\npassword = "test"\nport = 5432' >> config/test.toml
    make migrate ENV="$INPUT_ENV"
    make -k test coverage.xml ENV="$INPUT_ENV"
elif [ "$INPUT_ENV" == "composed" ]
  then
    cp /home/stageai/sunshine/docker/composed.toml /home/stageai/sunshine/config/dev.toml
    cd /home/stageai/sunshine
    make migrate ENV=dev
    make build
    SUNSHINE_ENV=dev sunshine openapi
    cd /home/stageai/sunshine/cmd/sunshine && SUNSHINE_ENV=dev watcher serve
else
  make migrate ENV="$INPUT_ENV"
  SUNSHINE_ENV=$INPUT_ENV sunshine openapi
  SUNSHINE_ENV=$INPUT_ENV sunshine serve
fi

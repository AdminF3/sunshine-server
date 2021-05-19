#!/bin/bash

INPUT_ENV=$1
POSTGRES_HOST=$2
DOCKER_IMAGE=$3
CONTAINER_NAME="sunshine"
CONTAINER_NETWORK="sunshine"

if [ "$INPUT_ENV" != "production" ]
  then
    CONTAINER_NAME=$INPUT_ENV-$CONTAINER_NAME
    CONTAINER_NETWORK=$CONTAINER_NETWORK-$INPUT_ENV
fi

docker container rm "$CONTAINER_NAME" -f
docker container run \
  --detach \
  --env POSTGRES_HOST="$POSTGRES_HOST" \
  --volume "$PROJECTS_PATH"/sunshine-uploads:/home/stageai/sunshine/uploads \
  --volume "$PROJECTS_PATH"/sunshine/docker/"$INPUT_ENV".toml:/home/stageai/sunshine/config/"$INPUT_ENV".toml \
  --network="$CONTAINER_NETWORK" \
  --name "$CONTAINER_NAME" \
  "$DOCKER_IMAGE" "$INPUT_ENV"


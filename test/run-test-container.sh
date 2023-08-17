#!/bin/sh

SCRIPT_DIR=$(dirname $(realpath -s $0))

cd $SCRIPT_DIR && cd ..

export DOCKERKIT=1

docker build --build-arg="RELEASE_VERSION=0.0.0-dev" --tag ngstaticserver-test . 
docker run --publish 8080:8080 ngstaticserver-test
#!/bin/sh

export DOCKERKIT=1

docker build --tag ngstaticserver . 
docker run --publish 8080:8080 ngstaticserver
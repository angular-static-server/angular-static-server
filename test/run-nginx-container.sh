#!/bin/sh

SCRIPT_DIR=$(dirname $(realpath -s $0))

cd $SCRIPT_DIR && cd ..

curl https://raw.githubusercontent.com/nginxinc/docker-nginx/master/modules/Dockerfile.alpine -o "$SCRIPT_DIR/nginx/Dockerfile"
cat "$SCRIPT_DIR/nginx/Dockerfile.template" >> "$SCRIPT_DIR/nginx/Dockerfile"

export DOCKERKIT=1

# --build-arg NGINX_FROM_IMAGE="nginxinc/nginx-unprivileged:stable-alpine" 
docker build --tag ngstaticserver-nginx --build-arg ENABLED_MODULES="brotli" --file "$SCRIPT_DIR/nginx/Dockerfile" .
docker run --publish 8080:8080 ngstaticserver-nginx


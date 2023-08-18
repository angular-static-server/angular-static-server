#!/bin/sh

# This script expects the nginx container to have been built with the run-nginx-container.sh scripts.
# It also expects k6 (https://k6.io/docs/get-started/installation/) and
# bombadier (https://pkg.go.dev/github.com/codesenberg/bombardier) to be installed.

SCRIPT_DIR=$(dirname $(realpath -s $0))

cd $SCRIPT_DIR && cd ../..

export DOCKERKIT=1

docker build --tag ngstaticserver-test . 

cd $SCRIPT_DIR

# angular-static-server

containerid=$(docker run --detach --publish 8080:8080 ngstaticserver-test)

k6 run -e TYPE=ngss benchmark.js

bombardier -p r http://localhost:8080/de-CH/ > bombardier-ngss-index.txt
bombardier -p r http://localhost:8080/de-CH/3rdpartylicenses.txt > bombardier-ngss-license.txt

docker stop $containerid


# nginx

containerid=$(docker run --detach --publish 8080:8080 ngstaticserver-nginx)

k6 run -e TYPE=nginx benchmark.js

bombardier -p r http://localhost:8080/de-CH/ > bombardier-nginx-index.txt
bombardier -p r http://localhost:8080/de-CH/3rdpartylicenses.txt > bombardier-nginx-license.txt

docker stop $containerid


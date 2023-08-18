#!/bin/sh

SCRIPT_DIR=$(dirname $(realpath -s $0))

cd $SCRIPT_DIR && cd ..

export DOCKERKIT=1

docker build --target server --tag ngstaticserver .

TMP_FILE=$(mktemp -q /tmp/bar.XXXXXX || exit 1)
trap 'rm -f -- "$TMP_FILE"' EXIT

docker save ngstaticserver > "$TMP_FILE"

filesize=$(ls -lh $TMP_FILE | awk '{print  $5}')
echo "Container image size: $filesize"

rm -f -- "$TMP_FILE"
trap - EXIT
exit
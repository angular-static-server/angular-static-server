#!/bin/sh

SCRIPT_DIR=$(dirname $(realpath -s $0))
NG_DIR="$SCRIPT_DIR/angular"

cd $NG_DIR && npm run build
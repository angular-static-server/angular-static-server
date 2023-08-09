#!/bin/sh

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
NG_DIR="$SCRIPT_DIR/angular"

cd $NG_DIR && yarn build
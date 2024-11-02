#!/bin/bash

set -e -x -o pipefail -u

export PATH=./node_modules/.bin/:$PATH

# check nodejs and npm prerequisites
nodejs --version
npm --version

# install build and deploy tools (func-js)
if [ ! -d "node_modules" ]; then
    npm install
fi

# compile contracts
mkdir -pv compiled
func-js contract/qqoken.fc --boc-base64 compiled/qqoken.boc64
func-js contract/qqollection.fc --boc-base64 compiled/qqollection.boc64

# deploy contracts
AUTH_ADDR=${AUTH_ADDR:=}
OWNER_ADDR=${OWNER_ADDR:=}
QQOKEN_VALUE=${QQOKEN_VALUE:=100}
QQOLLECTION_ID=${QQOLLECTION_ID:=1000}
QQOKEN_ID=${QQOKEN_ID:=99}

nodejs deploy-qq.js \
    --auth "${AUTH_ADDR}" \
    --owner "${OWNER_ADDR}" \
    --value "${QQOKEN_VALUE}" \
    --qqolection-id "${QQOLLECTION_ID}" \
    --qqoken-id "${QQOKEN_ID}"

#!/bin/bash
set -x

JSMIN_BIN=`which terser`
CSSMIN_BIN=`which csso`
O5_BIN=`which o5`
CONVERT_BIN=`which convert`

WEBAPP_DIR=${WEBAPP_DIR:-"webapp"}
SOURCE_DIR="${WEBAPP_DIR}/src"
TARGET_DIR="${WEBAPP_DIR}"
BUILD_DIR=`mktemp -d`

if [ -n "${JSMIN_BIN}" ] && [ -n "${CSSMIN_BIN}" ] && [ -n "${O5_BIN}" ]
then
    # encode font file for embedding
    base64 -w0 "${SOURCE_DIR}/4iCv6KVjbNBYlgoCxCvjsGyN.woff2" > "${BUILD_DIR}/4iCv6KVjbNBYlgoCxCvjsGyN.woff2.base64"
    # src/qqoin.png → apple-touch-icon.png, favicon.ico
    [ "${SOURCE_DIR}/qqoin.png" -nt "${TARGET_DIR}/apple-touch-icon.png" ] && \
    convert "${SOURCE_DIR}/qqoin.png" -resize 144x144 "${TARGET_DIR}/apple-touch-icon.png" && \
    convert "${SOURCE_DIR}/qqoin.png" -resize 96x96 "${TARGET_DIR}/favicon.ico"
    # minify assets and compile index.html
    "${JSMIN_BIN}" "${SOURCE_DIR}/qqoin.js" >"${BUILD_DIR}/qqoin.min.js"
    "${CSSMIN_BIN}" "${SOURCE_DIR}/qqoin.css" >"${BUILD_DIR}/qqoin.min.css"
    "${O5_BIN}" -dd ${ENVFILE:-/dev/null} -i "${SOURCE_DIR}/index.html.template" -w "${BUILD_DIR}" -start "/***" -end "***/" \
    | tee "${TARGET_DIR}/index.html"
    ls -ltr "${TARGET_DIR}"
    # cleanup
    [ -n "${BUILD_DIR}" -a -d "${BUILD_DIR}" ] && rm -rfv "${BUILD_DIR}"
else
    echo "build tools missing"
    echo "JSMIN_BIN=${JSMIN_BIN}"
    echo "CSSMIN_BIN=${CSSMIN_BIN}"
    echo "O5_BIN=${O5_BIN}"
    exit 1
fi

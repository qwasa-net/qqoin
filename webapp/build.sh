#!/bin/bash
set -x

JSMIN_BIN=`which terser`
CSSMIN_BIN=`which csso`
O5_BIN=`which o5`
CONVERT_BIN=`which convert`

if [ -n "${JSMIN_BIN}" ] && [ -n "${CSSMIN_BIN}" ] && [ -n "${O5_BIN}" ]
then
    #
    BUILDDIR=`mktemp -d`
    base64 -w0 src/4iCv6KVjbNBYlgoCxCvjsGyN.woff2 > "${BUILDDIR}/4iCv6KVjbNBYlgoCxCvjsGyN.woff2.base64"
    convert src/qqoin.png -resize 144x144 apple-touch-icon.png
    convert src/qqoin.png -resize 96x96 favicon.ico
    "${JSMIN_BIN}" src/qqoin.js >"${BUILDDIR}/qqoin.min.js"
    "${CSSMIN_BIN}" src/qqoin.css >"${BUILDDIR}/qqoin.min.css"
    "${O5_BIN}" -dd ${ENVFILE:-/dev/null} -i src/index.html.template -w "${BUILDDIR}" -start "/***" -end "***/" | tee index.html
    ls -ltr
else
    echo "build tools missing"
    echo "JSMIN_BIN=${JSMIN_BIN}"
    echo "CSSMIN_BIN=${CSSMIN_BIN}"
    echo "O5_BIN=${O5_BIN}"
    exit 1
fi

#!/bin/bash
set -e -u -o pipefail

echo "*** QQoin TG Web App build script *** ***"
echo ""

JSMIN_BIN=`which terser`
CSSMIN_BIN=`which cleancss`
O5_BIN=`which o5`
CONVERT_BIN=`which convert`

echo "*** build tools:"
echo "JSMIN_BIN=${JSMIN_BIN}  # terser - JavaScript parser and mangler/compressor and beautifier toolkit"
echo "CSSMIN_BIN=${CSSMIN_BIN}  # cleancss - CSS minifier"
echo "O5_BIN=${O5_BIN}  # o5 - micro macro processor (github.com/qwasa-net/o5)"
echo "CONVERT_BIN=${CONVERT_BIN}  # ImageMagick convert"
echo ""

if [ -n "${JSMIN_BIN}" ] && [ -n "${CSSMIN_BIN}" ] && [ -n "${O5_BIN}" ] && [ -n "${CONVERT_BIN}" ];
then
    true  # all tools found
else
    echo "*** build tools missing ***"
    exit 1
fi

WEBAPP_DIR=${WEBAPP_DIR:-"webapp"}
SOURCE_DIR="${WEBAPP_DIR}/src"
TARGET_DIR="${WEBAPP_DIR}"
BUILD_DIR=`mktemp -d`

echo "*** directories:"
echo "SOURCE_DIR=${SOURCE_DIR}"
echo "TARGET_DIR=${TARGET_DIR}"
echo ""

if [ -d "${SOURCE_DIR}" ] && [ -d "${TARGET_DIR}" ] && [ -n "${BUILD_DIR}" ];
then
    true  # all directories found
else
    echo "*** source or target directories missing ***"
    exit 1
fi

echo "*** building:"
set -x

# encode font file for embedding
base64 -w0 "${SOURCE_DIR}/4iCv6KVjbNBYlgoCxCvjsGyN.woff2" > "${BUILD_DIR}/4iCv6KVjbNBYlgoCxCvjsGyN.woff2.base64"

# src/qqoin.png → apple-touch-icon.png, favicon.ico (if newer)
[ "${SOURCE_DIR}/qqoin.png" -nt "${TARGET_DIR}/apple-touch-icon.png" ] && \
convert "${SOURCE_DIR}/qqoin.png" -resize 144x144 "${TARGET_DIR}/apple-touch-icon.png" && \
convert "${SOURCE_DIR}/qqoin.png" -resize 256x256 "${TARGET_DIR}/qqoken.png" && \
convert "${SOURCE_DIR}/qqoin.png" -resize 300x300 "${TARGET_DIR}/qqoken.webp" && \
convert "${SOURCE_DIR}/qqoin.png" -resize 96x96 "${TARGET_DIR}/favicon.ico"

# minify assets and compile index.html
"${JSMIN_BIN}" "${SOURCE_DIR}/qqoin.js" >"${BUILD_DIR}/qqoin.min.js"
"${CSSMIN_BIN}" "${SOURCE_DIR}/qqoin.css" >"${BUILD_DIR}/qqoin.min.css"
"${O5_BIN}" -dd ${ENVFILE:-/dev/null} \
    -i "${SOURCE_DIR}/index.html.template" \
    -w "${BUILD_DIR}" \
    -start "/***" -end "***/" \
    > "${TARGET_DIR}/index.html"  # | tee "${TARGET_DIR}/index.html"


ls -lGgtr "${TARGET_DIR}"

set +x

# cleanup
echo "*** cleanup:"
[ -n "${BUILD_DIR}" -a -d "${BUILD_DIR}" ] && rm -rfv "${BUILD_DIR}"

#!/bin/sh
set -e

export LOG_LEVEL="${LOG_LEVEL:-ERROR}"
export AUTOTLS_DOMAINS="${AUTOTLS_DOMAINS:-}"
export AUTOTLS_EMAIL="${AUTOTLS_EMAIL:-}"
export SERVER_BIN="${SERVER_BIN:-/app/server}"
export CONFIG_PATH="${CONFIG_PATH:-/app/proxies.yaml}"

if [ "$1" = 'start-server' ]; then

  if [ ! -f "${CONFIG_PATH}" ]; then
cat <<EOF > $CONFIG_PATH
server:
  bind_address: ":http"

EOF
  fi

  args=""
  if [ "LOG_LEVEL" != "" ]; then
    args="${args} -level=${LOG_LEVEL}"
  fi

  if [ "$AUTOTLS_DOMAINS" != "" ]; then
    args="${args} -domains=${AUTOTLS_DOMAINS}"
  fi

  if [ "$AUTOTLS_EMAIL" != "" ]; then
    args="${args} -email=${AUTOTLS_EMAIL}"
  fi

  echo "Starting server: ${SERVER_BIN} ${args}"

  # run
  exec "$SERVER_BIN" $args

else
  echo ""
  # run
  exec "$@"
fi

#!/bin/sh
set -eu

: "${FRONTEND_GATEWAY_URL:=http://localhost:8080/graphql}"

envsubst '${FRONTEND_GATEWAY_URL}' < /etc/nginx/templates/config.js.tmpl > /usr/share/nginx/html/__sporthub_config__.js

echo "Serving frontend via nginx (gateway ${FRONTEND_GATEWAY_URL})"
exec nginx -g 'daemon off;'

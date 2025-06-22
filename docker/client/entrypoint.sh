#!/bin/sh

sed -i "s/\${WEB_PORT}/${WEB_PORT}/g" /etc/nginx/conf.d/default.conf
sed -i "s/\${API_PORT}/${API_PORT}/g" /etc/nginx/conf.d/default.conf
sed -i "s/\${API_URL}/${API_URL}/g" /etc/nginx/conf.d/default.conf

exec /docker-entrypoint.sh "$@"
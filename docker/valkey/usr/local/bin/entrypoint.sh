#!/bin/sh

VALKEY_PASSWORD=${VALKEY_PASSWORD:-"valkey"}

sed -i "s/\${VALKEY_PASSWORD}/${VALKEY_PASSWORD}/g" /usr/local/etc/valkey/valkey.conf

valkey-server /usr/local/etc/valkey/valkey.conf
ARG ALPINE_VERSION=3.20
ARG VALKEY_VERSION=8.0.1-alpine${ALPINE_VERSION}

FROM valkey/valkey:${VALKEY_VERSION}

COPY docker/valkey/usr/local/etc/valkey/valkey.conf /usr/local/etc/valkey/valkey.conf

CMD ["valkey-server", "/usr/local/etc/valkey/valkey.conf"]
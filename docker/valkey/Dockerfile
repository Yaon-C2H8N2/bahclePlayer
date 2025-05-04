ARG ALPINE_VERSION=3.20
ARG VALKEY_VERSION=8.0.1-alpine${ALPINE_VERSION}

FROM valkey/valkey:${VALKEY_VERSION}

ARG VALKEY_PASSWORD=valkey

COPY docker/valkey/usr/local/etc/valkey/valkey.conf /usr/local/etc/valkey/valkey.conf

RUN sed -i "s/\${VALKEY_PASSWORD}/${VALKEY_PASSWORD}/g" /usr/local/etc/valkey/valkey.conf

CMD ["valkey-server", "/usr/local/etc/valkey/valkey.conf"]
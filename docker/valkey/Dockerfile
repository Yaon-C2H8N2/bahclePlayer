ARG ALPINE_VERSION=3.20
ARG VALKEY_VERSION=8.0.1-alpine${ALPINE_VERSION}

FROM valkey/valkey:${VALKEY_VERSION}

COPY docker/valkey/usr/local/bin/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

COPY docker/valkey/usr/local/etc/valkey/valkey.conf /usr/local/etc/valkey/valkey.conf

CMD ["/usr/local/bin/entrypoint.sh"]
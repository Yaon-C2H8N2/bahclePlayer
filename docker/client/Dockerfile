ARG NODE_VERSION=23.1.0
ARG ALPINE_VERSION=3.20
ARG NGINX_VERISON=1.26.2

FROM node:${NODE_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /app

COPY client/package.json client/package-lock.json ./
RUN npm install

COPY client/ .
RUN npm run build

FROM nginx:${NGINX_VERISON}-alpine${ALPINE_VERSION}-slim

ARG API_PORT=8081
ARG WEB_PORT=80
ARG API_URL=api

ENV API_PORT=${API_PORT}
ENV WEB_PORT=${WEB_PORT}
ENV API_URL=${API_URL}

COPY docker/client/etc/nginx/conf.d /etc/nginx/conf.d

COPY --from=builder /app/dist /var/www/html

COPY docker/client/entrypoint.sh /entrypoint.sh

EXPOSE ${WEB_PORT}

ENTRYPOINT ["/bin/sh", "/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
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

ARG API_PORT=8080
ARG WEB_PORT=80

COPY docker/client/etc/nginx/conf.d /etc/nginx/conf.d

RUN sed -i "s/\${WEB_PORT}/${WEB_PORT}/g" /etc/nginx/conf.d/default.conf && \
    sed -i "s/\${API_PORT}/${API_PORT}/g" /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /var/www/html

EXPOSE ${WEB_PORT}
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

RUN rm /etc/nginx/conf.d/default.conf
RUN touch /etc/nginx/conf.d/default.conf
RUN echo "server { listen ${WEB_PORT}; location / { root /var/www/html; index index.html; try_files \$uri /index.html; } location /api { rewrite /api/(.*) /\$1 break; proxy_pass http://api:${API_PORT}; } }" > /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /var/www/html

EXPOSE 80
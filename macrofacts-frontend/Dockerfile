# ---- Build stage ----
FROM node:20-alpine AS build
LABEL authors="kevin fernandes"

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# ---- Runtime stage ----
FROM nginx:alpine

# SPA fallback for React Router (safe even if you don't use it yet)
RUN printf 'server { \
  listen 80; \
  root /usr/share/nginx/html; \
  location / { try_files $uri /index.html; } \
}' > /etc/nginx/conf.d/default.conf

COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80

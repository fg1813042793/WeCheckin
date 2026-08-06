# Build the uni-app H5 bundle, then serve it with Nginx.
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_BASE_URL=
ARG VITE_DINGTALK_CORP_ID=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_DINGTALK_CORP_ID=${VITE_DINGTALK_CORP_ID}

RUN npm run build:h5

FROM nginx:1.27-alpine

COPY docker/nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist/build/h5 /usr/share/nginx/html

EXPOSE 80

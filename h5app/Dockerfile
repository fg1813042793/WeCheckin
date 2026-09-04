FROM node:20-alpine AS builder

WORKDIR /app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"

RUN corepack enable

ARG VITE_API_BASE_URL=
ARG VITE_DEV_PROXY_TARGET=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_DEV_PROXY_TARGET=${VITE_DEV_PROXY_TARGET}

COPY package.json ./
COPY pnpm-lock.yaml ./
#RUN pnpm install --frozen-lockfile
RUN pnpm install

COPY . .
RUN pnpm build:h5

FROM nginx:1.27-alpine

COPY deploy/nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist/build/h5 /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1

CMD ["nginx", "-g", "daemon off;"]

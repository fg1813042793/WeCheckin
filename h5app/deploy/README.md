# H5 部署

用于部署 uni-app H5 构建产物，默认构建输出目录为 `dist/build/h5`。

## 后端接口

部署采用同域代理方式：前端请求 `/api/v2/...`，Nginx 转发到后端服务。

后端地址在 `deploy/docker-compose.yml` 中配置：

```yaml
environment:
  BACKEND_URL: http://192.168.50.6:8083
```

`VITE_API_BASE_URL` 和 `VITE_DEV_PROXY_TARGET` 在 Docker 构建参数中保持为空，用来覆盖本地 `.env` 中的值，避免生产 H5 直接请求或暴露本地后端地址。

## Docker Compose

在 `h5app` 目录执行：

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

默认访问端口为 `8086`，可在 `deploy/docker-compose.yml` 中调整端口映射。

## Docker

在 `h5app` 目录执行：

```bash
docker build -f deploy/Dockerfile --build-arg VITE_API_BASE_URL= --build-arg VITE_DEV_PROXY_TARGET= -t wecheckin-h5app:latest .
docker run -d --name wecheckin-h5app -p 8086:80 -e BACKEND_URL=http://192.168.50.6:8083 wecheckin-h5app:latest
```

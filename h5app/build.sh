#VERSION=1.2.5 ./build.sh
VERSION="${VERSION:-1.0.0}"
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="${VERSION}" \
  -t registry.cn-hangzhou.aliyuncs.com/ms_other_app/wecheckin-dingding-h5:"${VERSION}" \
  --push .
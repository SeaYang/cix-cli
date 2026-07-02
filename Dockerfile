# syntax=docker/dockerfile:1

# ─── 构建阶段 ─────────────────────────────────────────────────────────────────
FROM golang:1.26.4-alpine AS builder

ARG TARGETOS=linux
ARG TARGETARCH
# 国内网络默认走 goproxy.cn；海外环境可用 --build-arg GOPROXY=https://proxy.golang.org,direct 覆盖
ARG GOPROXY=https://goproxy.cn,direct
# 可选：构建时注入版本信息
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

WORKDIR /workspace

# 优先复制 go.mod / go.sum，利用 Docker 层缓存
COPY go.mod go.sum ./
RUN GOPROXY=${GOPROXY} go mod download

# 复制源码并编译（CGO_ENABLED=0 保证纯静态二进制，跨平台无依赖）
COPY main.go ./
COPY cmd/ cmd/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w \
      -X github.com/seayang/cix-cli/cmd/version.Version=${VERSION} \
      -X github.com/seayang/cix-cli/cmd/version.BuildTime=${BUILD_TIME} \
      -X github.com/seayang/cix-cli/cmd/version.GitCommit=${GIT_COMMIT}" \
    -o cix .

# ─── 运行阶段 ─────────────────────────────────────────────────────────────────
FROM alpine:3.22

# CI 容器中常用的基础工具
#   bash ca-certificates tzdata : shell / HTTPS / 时区
#   openssh sshpass             : SSH 访问（含密码登录）
#   git git-lfs                 : 代码拉取与大文件
#   rsync zip                   : 文件同步与归档
#   jq yq curl                  : JSON/YAML 处理与 HTTP 请求
#   coreutils                   : GNU 版基础命令（替代 BusyBox 默认实现）
#   util-linux                  : blkid/lsblk/mount/fdisk 等系统与磁盘工具
RUN apk add --no-cache \
      bash ca-certificates tzdata \
      openssh sshpass \
      git git-lfs \
      rsync zip jq yq curl \
      coreutils util-linux

WORKDIR /workspace

COPY --from=builder /workspace/cix /usr/local/bin/cix

ENTRYPOINT ["cix"]
CMD ["--help"]

# cix

基于 [cobra](https://github.com/spf13/cobra) 的 CI 流水线命令行工具，运行于容器环境，内置轻量 HTTP 客户端等常用工具。容器镜像预装 bash / git / git-lfs / rsync / jq / yq / curl 等 CI 常用命令，开箱即用。

## 特性

- `cix version` — 输出版本与构建信息
- `cix http get|post` — 轻量 HTTP 客户端，支持自定义 header、超时、verbose
- 纯静态二进制（`CGO_ENABLED=0`），跨平台无依赖
- 容器镜像内置 CI 常用工具，开箱即用

## 环境要求

- Go >= 1.26.4
- macOS / Linux / Windows

## 安装

```bash
go install github.com/seayang/cix-cli@latest
```

或从 Releases 下载预编译二进制，也可使用下方 [Docker](#docker) 镜像。源码构建：

```bash
go build -o cix .
./cix --help
```

## 命令

### version

输出版本与构建信息（版本号 / 构建时间 / Git Commit / Go 版本 / 平台）。

```bash
./cix version
```

```
cix
  Version    : 0.1.0
  Build Time : 2026-06-22T03:12:15Z
  Git Commit : b46c3e6
  Go Version : go1.26.4
  OS/Arch    : darwin/arm64
```

### http

轻量 HTTP 客户端，包含 `get` / `post` 两个子命令：

```bash
# GET
./cix http get https://httpbin.org/get
./cix http get https://api.example.com --header "Authorization:Bearer token" --timeout 10 -v

# POST
./cix http post https://httpbin.org/post --data '{"name":"cix"}'
```

| 参数 | 说明 |
|------|------|
| `--timeout` | 请求超时秒数（默认 30） |
| `--header` | 请求头，`key:value`，可多次指定 |
| `--data` | POST 请求体（仅 `http post`） |
| `-v, --verbose` | 打印请求详情（全局选项） |

## Docker

镜像分两阶段：`golang:1.26.4-alpine` 编译，`alpine:3.22` 运行，并预装 CI 常用工具。

### 构建与运行

```bash
# 构建并加载到本地 Docker（指定平台）
docker buildx build --platform linux/arm64 -t cix:latest --load .

# 运行（镜像 ENTRYPOINT 为 cix）
docker run --rm cix:latest version
docker run --rm -v "$PWD":/work -w /work cix:latest http get https://httpbin.org/get
docker run -it --entrypoint /bin/sh cix:latest   # 进入容器，使用内置工具
```

### 推送到 Registry（多平台）

构建 `linux/amd64` + `linux/arm64` 多平台镜像并推送：

```bash
export DOCKER_USER=YOUR_DOCKERHUB_USERNAME
export VERSION=0.1.0
export GIT_COMMIT="$(git rev-parse --short HEAD)"
export BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" \
  -t "$DOCKER_USER/cix:$VERSION" \
  --push .
```

> `VERSION` / `BUILD_TIME` / `GIT_COMMIT` 通过 ldflags 注入 `cmd/version` 包；不传则走默认值 `dev` / `unknown`。多平台构建需 buildx，构建器配置、跨平台构建、推送细节见 [AGENTS.md](AGENTS.md#docker)。

### 镜像内置工具

```
bash  ca-certificates  tzdata  openssh  sshpass  git  git-lfs
rsync  zip  jq  yq  curl  coreutils  util-linux
```

- `coreutils` / `util-linux` — GNU 版基础命令，以及 blkid / lsblk / mount / fdisk 等系统与磁盘工具
- 其余覆盖 shell / HTTPS / 时区 / SSH / 代码拉取 / 文件归档 / JSON-YAML 处理 / HTTP 请求

---

## 项目结构

```
cix-cli/
├── main.go              # 入口：cmd.Execute()
├── cmd/
│   ├── root.go          # 根命令 cix，注册子命令（package cmd）
│   ├── version/         # cix version（NewCmd）
│   └── http/            # cix http get|post（NewCmd）
├── Dockerfile           # 多阶段构建
└── AGENTS.md            # 开发与构建速查
```

## 开发

构建细节、跨平台构建、Docker 多平台推送、代码组织约定见 [AGENTS.md](AGENTS.md)。

新增命令：在 `cmd/<name>/` 子包实现 `NewCmd()`，并在 `cmd/root.go` 的 `init()` 中注册。

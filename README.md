# cix

基于 [cobra](https://github.com/spf13/cobra) 的 CI 流水线命令行工具，运行于容器环境，内置轻量 HTTP 客户端等常用工具，镜像中预装 bash / git / git-lfs / rsync / jq / yq / curl 等 CI 常用命令。

## 特性

- `cix version` — 输出版本 / 构建信息
- `cix http get|post` — 轻量 HTTP 客户端，支持自定义 header、超时、verbose
- 容器镜像内置 CI 常用工具，开箱即用
- 纯静态二进制（`CGO_ENABLED=0`），跨平台无依赖

## 环境要求

- Go >= 1.26.4
- macOS / Linux / Windows

## 快速开始

```bash
go build -o cix .
./cix --help
./cix version
./cix http get https://httpbin.org/get
./cix http post https://httpbin.org/post --data '{"name":"cix"}'
```

## 安装

```bash
go install github.com/seayang/cix-cli@latest
```

或从 Releases 下载预编译二进制 / 使用下方 [Docker](#docker) 镜像。

---

## 用法示例

### version

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

### http get

```bash
./cix http get https://httpbin.org/get
./cix http get https://api.example.com --header "Authorization:Bearer token" --timeout 10 -v
```

### http post

```bash
./cix http post https://httpbin.org/post --data '{"key":"value"}'
```

| 参数 | 说明 |
|------|------|
| `--timeout` | 请求超时秒数（默认 30） |
| `--header` | 请求头，`key:value`，可多次指定 |
| `-v, --verbose` | 打印请求详情（全局选项） |

---

## 开发指南

### 1. Go 模块代理配置（国内必做）

直连 `proxy.golang.org` 在国内会超时，需切换到国内镜像。

**临时生效（单次命令）**

```bash
GOPROXY=https://goproxy.cn,direct go mod tidy
GOPROXY=https://goproxy.cn,direct go build ./...
```

**永久生效（推荐，写入 shell 配置）**

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GONOSUMCHECK="*"
```

验证：

```bash
go env GOPROXY
# 输出：https://goproxy.cn,direct
```

### 2. 下载依赖

```bash
# 整理并下载所有依赖
GOPROXY=https://goproxy.cn,direct go mod tidy
```

成功后会生成/更新 `go.sum`，所有依赖缓存在 `$GOPATH/pkg/mod`。

### 3. 构建项目

**开发构建（快速验证）**

```bash
go build -o cix .
```

**注入版本信息构建（发布用）**

```bash
go build -trimpath -ldflags "-s -w \
  -X github.com/seayang/cix-cli/cmd/version.Version=1.0.0 \
  -X github.com/seayang/cix-cli/cmd/version.GitCommit=$(git rev-parse --short HEAD) \
  -X 'github.com/seayang/cix-cli/cmd/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o cix .
```

**跨平台构建**

```bash
# macOS ARM64
GOOS=darwin  GOARCH=arm64 go build -o cix-darwin-arm64 .

# Linux AMD64
GOOS=linux   GOARCH=amd64 go build -o cix-linux-amd64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o cix-windows-amd64.exe .
```

### 4. 运行测试

**不构建，直接运行（开发期）**

```bash
go run main.go --help
go run main.go version
go run main.go http get https://httpbin.org/get
```

**功能验证（构建后）**

```bash
# 帮助信息
./cix --help
./cix http --help

# version 命令
./cix version

# http 命令（需要网络）
./cix http get https://httpbin.org/get
./cix http post https://httpbin.org/post --data '{"name":"cix"}'
```

**单元测试**

```bash
# 运行所有测试
go test ./...

# 运行指定包测试，并显示详细输出
go test -v ./cmd/...

# 带竞态检测
go test -race ./...
```

### 5. 常用 go 命令速查

| 命令 | 说明 |
|------|------|
| `go mod tidy` | 整理依赖，移除未使用的包 |
| `go mod download` | 预下载所有依赖到本地缓存 |
| `go mod verify` | 校验依赖完整性 |
| `go build ./...` | 编译所有包（不生成二进制） |
| `go vet ./...` | 静态分析检查 |
| `go clean -modcache` | 清除模块缓存（排查依赖问题时使用） |

---

## Docker

镜像分两阶段：`golang:1.26.4-alpine` 编译，`alpine:3.22` 运行，并预装 CI 常用工具。

### 前提说明

| 问题 | 结论 |
|------|------|
| Mac ARM64 能否构建 linux/amd64 镜像？ | **可以**，Docker Buildx + QEMU 模拟，`RUN apk add` 等命令在目标平台容器内执行，完全支持 |
| `apk add` 是否支持跨平台构建？ | **支持**，命令在目标平台的模拟环境中运行，不受宿主机架构影响 |
| Windows amd64 能否用同一镜像？ | 镜像为 Linux 容器，需通过 Docker Desktop (WSL2) 运行，不是原生 Windows 二进制 |

### 准备 buildx 多平台构建器（首次配置）

```bash
# 确认 buildx 版本
docker buildx version

# 创建并启用支持多平台的构建器（需要安装 QEMU，Docker Desktop 已内置）
docker buildx create --name multiarch --driver docker-container --use
docker buildx inspect --bootstrap
# 确认输出中包含：linux/amd64, linux/arm64
```

### 本地构建（跨平台验证）

构建 `linux/amd64` 镜像并加载到本地 Docker（`--load` 不支持多平台同时加载）：

```bash
docker buildx build \
  --platform linux/amd64 \
  -t cix:latest \
  --load \
  .

# 验证镜像
docker run --rm cix:latest --help
docker run --rm cix:latest version
docker run --rm cix:latest http get https://httpbin.org/get
```

### 本机 Mac（arm64）原生构建

> 上面构建的 `linux/amd64` 镜像在本机 arm64 上需通过 QEMU 模拟运行，速度慢且易出问题。
> 下面的命令直接构建 **`linux/arm64`** 镜像，在 Docker Desktop for Mac (arm64) 上**原生运行**，无需模拟。

```bash
# 构建 linux/arm64 镜像并加载到本地 Docker（适配本机 Mac arm64，原生执行）
docker buildx build \
  --platform linux/arm64 \
  -t cix:latest \
  -t cix:local-arm64 \
  --load \
  .

# 验证镜像架构（应输出：arm64）
docker inspect --format '{{.Architecture}}' cix:latest

# 在本机运行验证
docker run --rm cix:latest --help
docker run --rm cix:latest version
docker run --rm cix:latest http get https://httpbin.org/get

# 挂载本机目录执行（例如读取本地文件）
docker run --rm -v "$PWD":/work -w /work cix:latest version

# 进入容器，使用内置工具
docker run -it --entrypoint /bin/sh cix:latest
```

> **说明**
> - Docker Desktop for Mac 运行的是 Linux 容器，因此本机执行的平台是 `linux/arm64`（而非 `darwin/arm64`）。
> - `--load` 只能加载单平台镜像，所以这里指定 `--platform linux/arm64` 一项即可，不可与 amd64 同时 `--load`。
> - 构建器仍使用上面创建的 `multiarch`（docker-container 驱动）即可；若想用默认构建器，可执行 `docker buildx use default`。

### 推送到 Docker Hub

**第一步：登录 Docker Hub**

```bash
docker login
# 输入 Docker Hub 用户名和密码
```

**第二步：设置版本信息（构建 / 推送共用）**

```bash
export DOCKER_USER=YOUR_DOCKERHUB_USERNAME   # ← 替换为你的 Docker Hub 用户名
export VERSION=0.1.0
export GIT_COMMIT="$(git rev-parse --short HEAD)"
export BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

> 这些变量在当前 shell 会话中持续生效。`BUILD_TIME` 在构建与推送两步**共用同一份**，第四步才能命中第三步的构建缓存；若两步各自重新 `$(date ...)`，时间不同会让 buildx 缓存失效而重新编译。

**第三步：本地构建多平台镜像（验证构建，不推送）**

> 多平台镜像无法用 `--load` 加载到本地 Docker（`--load` 仅支持单平台），因此这里只做构建验证：编译并打包 `linux/amd64` 与 `linux/arm64`，结果保留在 buildx 构建器缓存中，供下一步直接复用。

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" \
  -t "$DOCKER_USER/cix:$VERSION" \
  .
```

如需将多平台镜像存成本地文件（离线分发 / 传输），可用 `--output` 导出为 OCI 镜像包：

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" \
  --output type=oci,dest=cix-multiarch.tar \
  .
```

> 推送该本地包需借助 `skopeo` 或 `crane`（Docker CLI 无法直接 push 多平台 tar）：
> `skopeo copy oci-archive:cix-multiarch.tar docker://$DOCKER_USER/cix:latest`

**第四步：推送多平台镜像**

```bash
# 复用上一步构建缓存，打包并以 manifest list 形式推送到 Registry
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" \
  -t "$DOCKER_USER/cix:$VERSION" \
  --push \
  .
```

> `--push` 会把多平台镜像以 manifest list 形式推送到 Registry，无需 `--load`；同一构建器会命中上一步缓存，这一步通常只做上传。
>
> Dockerfile 已声明 `ARG VERSION / BUILD_TIME / GIT_COMMIT`，上述 `--build-arg` 经 ldflags 注入到 `cmd/version` 包；不传则走默认值 `dev` / `unknown`。

**第五步：验证远程镜像**

```bash
# 查看镜像的多平台 manifest（确认含 linux/amd64、linux/arm64）
docker buildx imagetools inspect "$DOCKER_USER/cix:$VERSION"
```

**第六步：在目标机器上拉取使用**

```bash
docker pull "$DOCKER_USER/cix:latest"
docker run --rm "$DOCKER_USER/cix:latest" --help
docker run --rm "$DOCKER_USER/cix:latest" version   # 检查注入的版本信息
```

### 运行

```bash
docker run --rm cix:latest version
docker run --rm -v "$PWD":/work -w /work cix:latest http get https://httpbin.org/get
docker run -it --entrypoint /bin/sh cix:latest     # 进入容器，使用内置工具
```

### 运行镜像内置工具

```
bash  ca-certificates  tzdata  openssh  sshpass  git  git-lfs  rsync  zip  jq  yq  curl
```

### 常见问题

| 问题 | 解决方案 |
|------|----------|
| `go mod download` 超时 | Dockerfile 中已配置 `GOPROXY=https://goproxy.cn,direct`，可用 `--build-arg GOPROXY=` 覆盖 |
| `exec format error` | 镜像平台与运行平台不符，重新用正确 `--platform` 构建 |
| `buildx create` 失败 | 确保 Docker Desktop 已启动，尝试 `docker buildx use default` |
| 推送被拒绝 | 先执行 `docker login`，确认镜像名前缀与账号一致 |
| buildx 无法直连 Docker Hub | 改用默认 `docker` 驱动（`--builder default` / `docker build`），或为构建器配置 Registry 镜像 |

---

## 项目结构

```
cix-cli/
├── main.go              # 入口：cmd.Execute()
├── cmd/
│   ├── root.go          # 根命令 cix，注册子命令（package cmd）
│   ├── version/         # cix version（子包，NewCmd）
│   └── http/            # cix http get|post（子包，NewCmd）
├── Dockerfile           # 多阶段构建
├── AGENTS.md            # 开发速查（精简）
└── README.md
```

## 开发

代码组织约定与构建细节见 [AGENTS.md](AGENTS.md)。
新增命令：在 `cmd/<name>/` 子包实现 `NewCmd()`，并在 `cmd/root.go` 的 `init()` 中注册。

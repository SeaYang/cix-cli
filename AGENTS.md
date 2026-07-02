# cix-cli — Agent / 开发速查

基于 Go + cobra 的 CI 流水线命令行工具，运行于容器环境。
面向使用者的介绍、安装与用法见 [README.md](README.md)；本文件记录在仓库中工作所需的开发与构建细节（代码约定、跨平台构建、Docker 多平台构建与推送等）。

## 环境

- Go 1.26.4（`go.mod` 已锁定，请勿回退）
- macOS / Linux；CI 中以 linux 容器运行
- 国内拉依赖慢时：`go env -w GOPROXY=https://goproxy.cn,direct`（Dockerfile 默认即用该代理，可用 `--build-arg GOPROXY=` 覆盖）

## 构建与测试

```bash
go build -o cix .     # 开发构建
go vet ./...          # 静态检查
go test ./...         # 单元测试
```

发布构建（注入版本信息）：

```bash
go build -trimpath -ldflags "-s -w \
  -X github.com/seayang/cix-cli/cmd/version.Version=1.0.0 \
  -X github.com/seayang/cix-cli/cmd/version.GitCommit=$(git rev-parse --short HEAD) \
  -X 'github.com/seayang/cix-cli/cmd/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" -o cix .
```

### 跨平台构建

`CGO_ENABLED=0` 纯静态，Go 原生交叉编译，无需目标平台工具链：

```bash
GOOS=darwin  GOARCH=arm64 go build -o cix-darwin-arm64 .
GOOS=linux   GOARCH=amd64 go build -o cix-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o cix-windows-amd64.exe .
```

镜像的跨平台构建见下文 [Docker](#docker)。

## 代码组织约定（重要）

- 入口 `main.go` → `cmd.Execute()`；根命令 `cix` 定义在 `cmd/root.go`。
- **每个自定义命令独立一个子包，放在 `cmd/<命令名>/`**（如 `cmd/version`、`cmd/http`），对外暴露 `NewCmd() *cobra.Command`，由 `cmd/root.go` 注册。**不要直接把命令实现写进根 `cmd` 包（root.go）。**
- 新增命令步骤：建 `cmd/<name>/`，实现 `NewCmd()`，在 `cmd/root.go` 的 `init()` 里 `rootCmd.AddCommand(<name>.NewCmd())`。
- 全局选项（如 `--verbose`）挂在 root 的 PersistentFlags；子命令通过 `cmd.Flags().GetBool(...)` 读取继承下来的值。
- 版本信息变量在 `cmd/version`（`Version` / `BuildTime` / `GitCommit`），构建时通过 ldflags 覆盖。

## 项目结构

```
main.go              # 入口
cmd/root.go          # 根命令 cix + 注册子命令（package cmd）
cmd/version/         # cix version
cmd/http/            # cix http get|post
Dockerfile           # 多阶段构建，运行镜像内置 CI 常用工具
```

## Docker

### 本地构建与运行

```bash
# 构建并加载到本地 Docker（指定平台）
docker buildx build --platform linux/arm64 -t cix:latest --load .
docker run --rm cix:latest version
```

> Mac ARM64 上构建 `linux/amd64` 需 QEMU 模拟（慢且易出问题）；构建 `linux/arm64` 可在 Docker Desktop for Mac 上原生运行。`--load` 仅支持单平台，不能与多平台同时使用。

### 准备多平台构建器（首次）

```bash
docker buildx version                                       # 确认 buildx 可用
docker buildx create --name multiarch --driver docker-container --use
docker buildx inspect --bootstrap                           # 输出应含 linux/amd64, linux/arm64
```

### 多平台构建并推送

多平台镜像无法 `--load`，需直接 `--push`。建议先单独构建（验证 + 命中缓存）再推送：

```bash
export DOCKER_USER=YOUR_DOCKERHUB_USERNAME
export VERSION=0.1.0
export GIT_COMMIT="$(git rev-parse --short HEAD)"
export BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# 1) 仅构建（验证，结果留在 buildx 缓存）
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" --build-arg GIT_COMMIT="$GIT_COMMIT" --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" -t "$DOCKER_USER/cix:$VERSION" .

# 2) 推送（复用上一步缓存，以 manifest list 形式上传）
docker buildx build --platform linux/amd64,linux/arm64 \
  --build-arg VERSION="$VERSION" --build-arg GIT_COMMIT="$GIT_COMMIT" --build-arg BUILD_TIME="$BUILD_TIME" \
  -t "$DOCKER_USER/cix:latest" -t "$DOCKER_USER/cix:$VERSION" --push .

# 3) 校验远程 manifest（确认含 linux/amd64、linux/arm64）
docker buildx imagetools inspect "$DOCKER_USER/cix:$VERSION"
```

> - `BUILD_TIME` 等变量在构建与推送两步**共用同一份**，才能命中缓存；若各自重新 `$(date ...)`，时间不同会让 buildx 缓存失效而重编译。
> - Dockerfile 已声明 `ARG GOPROXY / VERSION / BUILD_TIME / GIT_COMMIT`，均可通过 `--build-arg` 覆盖；版本相关参数经 ldflags 注入 `cmd/version` 包，不传则走默认值 `dev` / `unknown`。

### 常见问题

| 问题 | 解决方案 |
|------|----------|
| `go mod download` / 构建拉依赖超时 | Dockerfile 默认 `GOPROXY=https://goproxy.cn,direct`，可用 `--build-arg GOPROXY=` 覆盖 |
| `exec format error` | 镜像平台与运行平台不符，用正确 `--platform` 重新构建 |
| `buildx create` 失败 | 确保 Docker Desktop 已启动，或 `docker buildx use default` |
| 推送被拒绝 | 先 `docker login`，确认镜像名前缀与账号一致 |
| buildx 无法直连 Docker Hub | 改用默认 `docker` 驱动（`--builder default` / `docker build`），或为构建器配置 Registry 镜像 |

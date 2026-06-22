# cix-cli — Agent / 开发速查

基于 Go + cobra 的 CI 流水线命令行工具，运行于容器环境。
安装、完整用法、Docker、跨平台构建等细节请见 [README.md](README.md)；本文件只记录在仓库中工作所需的关键信息。

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

```bash
docker buildx build --platform linux/arm64 -t cix:latest --load .
docker run --rm cix:latest version
```

> 若使用 `buildx` 的 `docker-container` 驱动构建器且无法直连 Docker Hub，请改用默认 `docker` 驱动构建器（`--builder default` / `docker build`），或为该构建器配置 Registry 镜像。

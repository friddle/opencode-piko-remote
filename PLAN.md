# opencode-piko-remote 项目计划

## 项目概述

通过 piko 隧道远程暴露 opencode web 界面，实现远程 AI 编程助手浏览器访问。

## 架构

```
[用户浏览器] -> [Piko Server (nginx+piko)] -> [Piko Agent] -> [opencode web]
```

### 客户端（部署在开发机上）

```
opencode-piko 二进制 = 内嵌 opencode + piko agent
```

启动流程：
1. 释放内嵌的 opencode 二进制到临时目录
2. 启动 `opencode web --port <random> --hostname 127.0.0.1`，设置 `OPENCODE_SERVER_PASSWORD` 环境变量
3. 启动 piko agent，连接远程 piko server，将流量导向本地 opencode web 端口

### 服务端（Docker 部署，公网可达）

与原 gotty-piko 一致：piko server + nginx 反向代理，按 endpoint name 路由。无需改动。

## 技术要点

| 项目 | 方案 |
|------|------|
| opencode 启动模式 | `opencode web --port <port> --hostname 127.0.0.1` |
| 认证 | `OPENCODE_SERVER_PASSWORD` 环境变量（opencode 自带） |
| 打包 | `go:embed` 嵌入 opencode 二进制，运行时释放到 tmp |
| piko 集成 | 复用 `github.com/andydunstall/piko` Go SDK |
| 端口 | 随机可用端口，避免冲突 |
| 并发 | `oklog/run` 管理生命周期 |

## 命令行

```bash
opencode-piko [project] \
  --name=my-dev \
  --remote=piko.example.com:8088 \
  --pass=secret123 \
  --server-port=8022 \
  --auto-exit=true \
  --model=provider/model \
  --agent=my-agent \
  --cors=https://example.com
```

### 参数说明

#### Piko 相关（本项目特有）

| 参数 | 必需 | 默认值 | 说明 |
|------|------|--------|------|
| `--name` | 是 | - | 端点标识，访问路径 `http://server:8088/<name>/` |
| `--remote` | 是 | - | piko 服务器地址 (host:port) |
| `--user` | 否 | opencode | 认证账号（传给 OPENCODE_SERVER_USERNAME） |
| `--pass` | 否 | - | 认证密码（传给 OPENCODE_SERVER_PASSWORD） |
| `--server-port` | 否 | 8022 | piko 上游端口 |
| `--auto-exit` | 否 | true | 24小时自动退出 |

#### opencode web 透传参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `[project]` | 当前目录 | opencode 工作目录（positional） |

#### 内部自动处理（不暴露）

- `--port`: 随机可用端口（程序自动分配）
- `--hostname`: 固定 `127.0.0.1`
- `--cors`: 自动从 `--remote` 推导（`http://<remote>`），确保通过 piko 访问时不被 CORS 拦截

## 任务清单

### Phase 1: 项目清理

- [x] 清理 git 历史，设置新 remote
- [ ] 删除 gotty 相关代码
- [ ] 重命名模块 `gotty-piko-client` -> `opencode-piko-remote`
- [ ] 重命名二进制 `gottyp` -> `opencode-piko`

### Phase 2: 嵌入 opencode 二进制

- [ ] Makefile 添加 `download-opencode` 目标，按平台下载 opencode 到 `embed/` 目录
- [ ] 使用 `go:embed` 嵌入
- [ ] 运行时释放到 `os.TempDir()/opencode-piko/opencode`，赋执行权限

### Phase 3: 核心逻辑

- [ ] `opencode.go`: 管理 opencode 子进程（启动、等待端口就绪、停止）
- [ ] `service.go`: 改造启动流程（去掉 gotty，换成 opencode web）
- [ ] `config.go`: 更新配置结构（去掉 terminal/tmux，加 workdir/cors）
- [ ] `main.go`: 更新 cobra 命令定义

### Phase 4: 构建

- [ ] 更新 Makefile（download + embed + build）
- [ ] 支持 linux/amd64, linux/arm64, darwin/amd64, darwin/arm64
- [ ] 更新 .gitignore（忽略 embed/ 下的二进制）

### Phase 5: 服务端

- [ ] 更新镜像名称为 `opencode-piko-server`
- [ ] docker-compose.yaml 更新名称
- [ ] 其余逻辑不变

## 目录结构

```
opencode-piko-remote/
├── client/
│   ├── main.go
│   ├── src/
│   │   ├── config.go
│   │   ├── service.go
│   │   ├── opencode.go      # NEW: opencode 进程管理
│   │   └── embed.go         # NEW: 嵌入二进制释放
│   ├── embed/
│   │   └── .gitkeep         # opencode 二进制构建时下载到这里
│   ├── Makefile
│   ├── go.mod
│   └── go.sum
├── server/
│   ├── build/
│   │   ├── Dockerfile
│   │   ├── nginx.conf
│   │   ├── piko.conf
│   │   └── start.sh
│   ├── docker-compose.yaml
│   └── Makefile
└── PLAN.md
```

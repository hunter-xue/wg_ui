# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述
TUI 界面的 WireGuard 服务器安装、服务器管理、客户端管理应用。功能详情见 `docs/requirements.md`。

## 常用命令

```bash
go build -o build/wg_ui .        # 本地构建（产物在 build/）
go test ./...                     # 运行所有测试
go test ./internal/db/ -run TestX # 运行单个测试
make linux                        # 交叉编译 linux/amd64，产物 build/wg_ui_linux
make deploy HOST=1.2.3.4          # 交叉编译并 SCP 到指定服务器
make clean                        # 删除 build/ 目录
```

## 技术栈
- Go + BubbleTea (TUI) + modernc.org/sqlite (纯 Go SQLite，无 CGO) + golang.org/x/crypto/bcrypt（密码哈希）
- 目标运行环境：Debian 11+
- 数据库 schema：`docs/database.md`

## 架构

### 包结构
- `internal/db` — 数据层：Store 结构体封装 `*sql.DB`，提供 Server/Client/User CRUD。模型结构体在 `models.go`
- `internal/wg` — WireGuard 领域逻辑：密钥生成 (`keygen.go`)、配置文件生成 (`config.go`，纯函数)、安装 (`install.go`)、服务管理 (`service.go`)、地址校验 (`validate.go`)
- `internal/tui` — BubbleTea TUI：`root.go` 是根 Model 做屏幕路由，子屏幕在 `menu/`、`server/`、`client/`、`status/`、`setup/`、`settings/` 子包
- `internal/shell` — `exec.CommandContext` 薄封装，所有外部命令通过此包执行
- `main.go` — 入口，组装依赖（创建 app 结构体将 root model 与各子 model 用闭包连接）；首次运行检测

### TUI 屏幕常量（`internal/tui/app.go`）
`ScreenSetupPassword`（首次运行）→ `ScreenMenu` → `ScreenInstall` / `ScreenServerView` → `ScreenServerForm` / `ScreenClientList` → `ScreenClientCreate`（新建）或 `ScreenClientForm`（编辑）→ `ScreenClientDetail` / `ScreenStatus` / `ScreenSettings` → `ScreenChangePassword`

新建和编辑客户端使用不同的 Screen 常量，`OnSwitchScreen` 据此决定是否传入 existing client。

### 关键约定
- 配置变更流程：修改 DB → `GenerateServerConfig()` → 写入 `/etc/wireguard/wg0.conf` → `wg syncconf`（非破坏性）或 `systemctl restart`（接口级变更）
- 客户端配置文件生成后存入 `wg_clients.description` 字段
- 服务器 Endpoint（ip:port）存入 `wg_server.endpoint` 字段，为必填项
- 密钥对通过 `wg genkey`/`wg pubkey` 生成，仅存数据库，不落盘文件；客户端列表按 `r` 键可重新生成密钥对
- 每台主机只允许一个 WireGuard 服务器
- 客户端只能禁用不能删除
- 客户端 Address 校验：不能与服务器 IP 重复、必须同网段（/31、/32 扩展为 /24 判断）、不能与其他客户端重复（`internal/wg/validate.go`）
- 首次运行检测：`main()` 调用 `store.GetAdminUser()`，为 nil 时设初始屏幕为 `ScreenSetupPassword`
- 密码以 bcrypt 哈希存储于 `sys_users` 表；`BackupAndWriteServerConfig()` 在覆盖前自动备份

## 开发测试环境
- 开发在 macOS，`wg` 命令不可用，TUI 和 DB 操作可本地调试
- WireGuard 相关操作需交叉编译后部署到 Debian 测试（`make deploy`）
- 测试服务器信息：`docs/test_server.md`

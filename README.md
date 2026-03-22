# wg_ui

一个基于 TUI 的 WireGuard 管理工具，支持在 Debian 服务器上安装 WireGuard、管理服务器配置和客户端。

## 功能

- **安装**：一键安装 WireGuard 并配置为 systemd 服务，分步骤显示进度和日志
- **服务器管理**：创建、编辑、删除 WireGuard 服务器，自动生成配置文件
- **客户端管理**：创建、编辑、启用/禁用、重新生成密钥对，自动生成客户端配置文件
- **状态查看**：查看服务运行状态和客户端统计

## 系统要求

- Debian 11 及以上版本
- 需要 root 权限运行

## 安装

从 [Releases](../../releases) 下载对应版本，或从源码编译：

```bash
# 在 macOS/Linux 开发机上编译 Linux 版本
make linux

# 上传到服务器（构建产物在 build/ 目录）
scp build/wg_ui_linux root@your-server:/root/wg_ui
```

## 运行

```bash
chmod +x /root/wg_ui
/root/wg_ui
```

数据库文件 `wg.db` 会在当前目录自动创建。

---

## 使用教程

### 第一步：安装 WireGuard

启动程序后，选择 **Install WireGuard**，按 Enter 确认。

程序会先检测是否已安装：
- **已安装**：显示当前服务状态，按 `esc` 返回
- **未安装**：依次执行安装步骤，每步完成显示 `✓`，失败显示 `✗` 及错误日志

```
Install WireGuard

✓ Updating package list...
✓ Installing WireGuard...
✓ Enabling IP forwarding...
✓ Configuring systemd service...

WireGuard installed successfully!

esc/q: back to menu
```

安装完成后按 `esc` 返回主菜单。

---

### 第二步：配置服务器

选择 **Server Management**，按 `c` 新建服务器，填写以下信息：

| 字段 | 示例 | 说明 |
|------|------|------|
| Name | `my-server` | 服务器名称 |
| Address | `100.100.0.1/32` | 服务器 VPN 接口地址 |
| Listen Port | `51820` | WireGuard 监听端口 |
| MTU | `1420` | 默认即可 |
| DNS | 留空 | 可选 |
| PostUp | （已预填）| iptables 转发规则，通常保持默认 |
| PostDown | （已预填）| iptables 清理规则，通常保持默认 |
| Endpoint | `1.2.3.4` | **服务器公网 IP**，用于生成客户端配置 |
| Comments | 留空 | 备注，可选 |

Tab 键切换字段，在最后一个字段按 Enter 保存。

保存后程序自动：
1. 生成密钥对
2. 写入 `/etc/wireguard/wg0.conf`
3. 重启 WireGuard 服务

> 一台主机只能配置一个 WireGuard 服务器。

---

### 第三步：添加客户端

选择 **Client Management**，按 `c` 新建客户端：

| 字段 | 示例 | 说明 |
|------|------|------|
| Name | `phone` | 客户端名称 |
| Address | `100.100.0.2/32` | 客户端 VPN 地址（须与服务器同网段，不能与服务器或其他客户端重复） |
| Allowed IPs | `100.100.0.0/24` | 客户端可访问的网段 |
| MTU | `1420` | 默认即可 |
| DNS | 留空 | 可选 |
| Keepalive | `25` | 心跳保活间隔（秒），默认即可 |
| Comments | 留空 | 备注，可选 |

保存后程序自动：
1. 生成客户端密钥对
2. 生成客户端配置文件并存入数据库
3. 更新服务器配置，添加对应 `[Peer]` 节
4. 执行 `wg syncconf` 使配置生效（不中断现有连接）

---

### 第四步：获取客户端配置

在客户端列表中，选中客户端按 **Enter** 查看完整配置文件：

```
[Interface]
PrivateKey = <自动生成>
Address = 100.100.0.2/32
MTU = 1420

[Peer]
PublicKey = <服务器公钥>
AllowedIPs = 100.100.0.0/24
Endpoint = 1.2.3.4:51820
PersistentKeepalive = 25
```

将此内容复制到客户端设备的 WireGuard 应用中即可连接。

---

### 日常操作

#### 禁用/启用客户端

在客户端列表中选中客户端，按 **Space** 切换状态。禁用后该客户端立即从服务器配置中移除，无法连接；启用后立即恢复。

#### 重新生成密钥对

当客户端密钥泄露时，在客户端列表选中该客户端，按 **r** 重新生成密钥对。生成后需重新获取并导入新的客户端配置文件。

#### 查看运行状态

主菜单选择 **Status**，可以查看：
- WireGuard 服务运行状态
- 服务器地址和端口
- 客户端总数、已启用数量、已禁用数量

#### 按键速查

| 界面 | 按键 | 功能 |
|------|------|------|
| 所有界面 | `esc` | 返回上一级 |
| 所有界面 | `q` | 退出程序 |
| 列表/菜单 | `↑` `↓` 或 `k` `j` | 移动光标 |
| 客户端列表 | `c` | 新建客户端 |
| 客户端列表 | `e` | 编辑选中客户端 |
| 客户端列表 | `Enter` | 查看客户端配置 |
| 客户端列表 | `Space` | 启用/禁用 |
| 客户端列表 | `r` | 重新生成密钥对 |
| 服务器界面 | `c` | 新建服务器 |
| 服务器界面 | `e` | 编辑服务器 |
| 服务器界面 | `d` | 删除服务器 |
| 服务器界面 | `s` | 查看服务状态 |
| 表单 | `Tab` | 下一个字段 |
| 表单 | `Shift+Tab` | 上一个字段 |
| 表单 | `Enter`（最后字段） | 保存 |

---

## 从源码构建

```bash
# 依赖 Go 1.21+
go build -o build/wg_ui .   # 本地构建（产物在 build/ 目录）
make linux                    # 交叉编译 Linux amd64
make deploy                   # 编译并 SCP 到测试服务器
go test ./...                 # 运行测试
make clean                    # 清理 build/ 目录
```

## 技术栈

- [Go](https://golang.org/)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI 框架
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) — 纯 Go SQLite（无 CGO）

# 功能描述

## 领域知识
可以通过以下链接，学习WireGuard的配置方式：
https://www.man7.org/linux/man-pages/man8/wg-quick.8.html
https://www.man7.org/linux/man-pages/man8/wg.8.html

## 密钥对生成
- 密钥对通过以下方式生成（通过 shell 调用）：
  ```bash
  wg genkey | tee privatekey | wg pubkey > publickey
  ```
- 生成的密钥对不以文件形式保留在磁盘上，公钥和私钥保存在对应的数据库表中。

---

## 主菜单

应用启动后显示主菜单，包含以下5个选项：

| 选项 | 功能 |
|------|------|
| Install WireGuard | 安装 WireGuard |
| Server Management | 服务器配置管理 |
| Client Management | 客户端管理（需要先配置服务器） |
| Status | 查看运行状态 |
| Quit | 退出程序 |

按键：↑/↓ 导航，Enter 选择，q 退出。

---

## WireGuard 安装

用户选择 "Install WireGuard" 后，程序先检查是否已安装（`which wg`）：

### 已安装
显示提示"WireGuard is already installed"，并显示当前服务状态（`systemctl status wg-quick@wg0` 输出）。按 esc/q 返回主菜单。

### 未安装
按顺序执行以下4个步骤，每步显示 spinner 动画，完成后打 ✓，失败打 ✗ 并显示错误日志：

1. **Updating package list** — `apt-get update -y`
2. **Installing WireGuard** — `apt-get install -y wireguard`
3. **Enabling IP forwarding** — `sysctl -w net.ipv4.ip_forward=1`
4. **Configuring systemd service** — `systemctl enable wg-quick@wg0`

安装完成（成功或失败）后停留在安装界面显示完整日志，按 esc/q 返回主菜单。

---

## WireGuard 服务器管理

### 服务器详情界面
显示当前服务器所有字段，或"No server configured"提示。

按键：
- `c` 新建服务器（无服务器时可用）
- `e` 编辑服务器（有服务器时可用）
- `d` 删除服务器（有服务器时可用）
- `s` 查看服务状态
- `esc` 返回主菜单，`q` 退出

### 服务器表单字段

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| Name | 是 | — | 服务器名称 |
| Address | 是 | — | VPN 接口地址，如 `100.100.0.1/32` |
| Listen Port | 是 | — | WireGuard 监听端口 |
| MTU | 否 | `1420` | 最大传输单元 |
| DNS | 否 | 空 | DNS 服务器（可选） |
| PostUp | 否 | `iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE` | 启动后执行 |
| PostDown | 否 | `iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE` | 停止后执行 |
| Endpoint | 是 | — | 客户端连接地址，格式 `ip:port`（如 `1.2.3.4:51820`）。因服务器可能位于 NAT/防火墙后，对外端口不一定等于 Listen Port，需单独填写 |
| Comments | 否 | 空 | 备注 |

表单按键：Tab/↓ 下一字段，Shift+Tab/↑ 上一字段，在最后一个字段按 Enter 保存，esc 取消。

### 保存后的操作
- 新建：自动生成密钥对，写入数据库
- 新建/编辑/删除：重新生成 `/etc/wireguard/wg0.conf`，执行 `systemctl restart wg-quick@wg0`
- 一台主机只允许存在一个服务器

---

## WireGuard 客户端管理

进入客户端管理需要先配置服务器，否则无法进入。

### 客户端列表界面
表格显示所有客户端：Name | Address | Status（enabled/disabled），按 Name 排序。禁用的客户端以删除线显示。

按键：
- `c` 新建客户端
- `e` 编辑选中客户端
- Enter 查看选中客户端的完整配置文件
- Space 切换选中客户端的启用/禁用状态
- `r` 为选中客户端重新生成密钥对
- ↑/↓ 移动光标
- `esc` 返回主菜单，`q` 退出

### 客户端表单字段

| 字段 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| Name | 是 | — | 客户端名称 |
| Address | 是 | — | 客户端 VPN 地址，如 `100.100.0.x/32` |
| Allowed IPs | 是 | — | 客户端可访问的网段，如 `100.100.0.0/24, 10.100.0.0/16` |
| MTU | 否 | `1420` | 最大传输单元 |
| DNS | 否 | 空 | DNS 服务器（可选） |
| Keepalive | 否 | `25` | PersistentKeepalive（秒） |
| Comments | 否 | 空 | 备注 |

表单按键：Tab/↓ 下一字段，Shift+Tab/↑ 上一字段，在最后一个字段按 Enter 保存，esc 取消。

### 客户端地址校验
保存时对 Address 字段执行以下校验，不通过则显示错误提示：

1. **不能与服务器地址重复** — 客户端 IP 不得与服务器接口 IP 相同
2. **必须与服务器同网段** — 服务器地址为 `/31`、`/32` 时自动扩展为 `/24` 判断；其他前缀长度按实际网络判断
3. **不能与已有客户端地址重复** — 每个客户端的 IP 在同一服务器下必须唯一；编辑时跳过自身

### 保存后的操作
- 新建：自动生成密钥对
- 生成客户端配置文件（见格式示例），存入 `wg_clients.description` 字段
- 重新生成服务器配置文件（仅包含 enabled 的客户端作为 `[Peer]`）
- 执行 `wg syncconf` 同步配置（不中断现有连接）

### 密钥对重新生成
在客户端列表选中客户端，按 `r` 重新生成密钥对：
1. 生成新的私钥和公钥，覆盖原有密钥
2. 用新密钥重新生成客户端配置文件，更新 `description` 字段
3. 更新服务器配置中对应 `[Peer]` 的 PublicKey
4. 执行 `wg syncconf` 立即生效

重新生成密钥后，客户端设备需要重新导入新的配置文件才能连接。

### 启用/禁用
- 客户端只能禁用，不能删除
- 禁用后立即重新生成服务器配置并执行 `wg syncconf`，对应 `[Peer]` 节被移除
- 启用后同理恢复 `[Peer]` 节

### 查看客户端配置
选中客户端按 Enter，展示存储在 `description` 字段中的完整配置文件（可滚动查看，用于复制到客户端软件）。

---

## 状态界面

从主菜单选择 "Status" 进入，显示以下信息：

1. **WireGuard 服务状态** — `systemctl status wg-quick@wg0` 输出
2. **服务器信息**（已配置时）— Name、Address、Listen Port
3. **客户端统计** — 总数 / Enabled 数量（绿色） / Disabled 数量

只读界面，不做任何修改。按 esc/q 返回主菜单。

---

## 配置文件格式

### 服务器配置文件：`/etc/wireguard/wg0.conf`
```
[Interface]
Address = 100.100.0.1/32
ListenPort = 10001
PrivateKey = <server_private_key>
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE
MTU = 1420

[Peer]
PublicKey = <client_public_key>
AllowedIPs = 100.100.0.11/32
```

每个 enabled 的客户端对应一个 `[Peer]` 节，disabled 的客户端不出现在配置文件中。

### 客户端配置文件（存储于 `wg_clients.description`）
```
[Interface]
PrivateKey = <client_private_key>
Address = 100.100.0.11/24
MTU = 1420

[Peer]
PublicKey = <server_public_key>
AllowedIPs = 100.100.0.0/24, 10.100.0.0/16
Endpoint = <endpoint>
PersistentKeepalive = 25
```

Endpoint 取自服务器表单的 Endpoint 输入项（存储在 `wg_server.description` 字段），格式为 `ip:port`，为必填项。

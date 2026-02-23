# 概述

ptool 是一个自用的 PT (Private Tracker) 网站和 BitTorrent 客户端辅助工具。

## 主要特性

- **纯 CLI 程序**: 使用 Go 开发的单文件可执行程序，无外部依赖
- **无状态设计**: 程序不保存状态，不后台运行，通过 cron 等方式定时执行
- **使用简单**: 5 分钟配置即可开始全自动刷流
- **刷流功能**: 全自动刷流，无需配置选种/删种规则
- **自动辅种**: 支持 IYUU 和 Reseed 接口自动辅种
- **客户端控制**: 完整的 BT 客户端管理功能
- **浏览器模仿**: 自动绕过大多数站点的 CF 盾

## 支持的 BT 客户端

| 客户端 | 版本要求 | 备注 |
|--------|----------|------|
| qBittorrent | v4.1+ | 推荐使用，需要启用 Web UI |
| Transmission | <= v3.0 | 未充分测试 |

## 支持的 PT 站点

- 绝大部分使用 NexusPHP 的站点
- M-Team (馒头)
- 测试过的站点: U2、冬樱、红叶、聆音、铂金家等

## 下载与安装

- [开发版本](https://ci.appveyor.com/project/sagan/ptool/build/artifacts)
- [稳定版本](https://github.com/sagan/ptool/releases)

下载后将可执行文件放到 PATH 路径下即可使用。

## 快速开始（刷流）

### 1. 创建配置文件

```bash
ptool config create
```

配置文件默认位置:
- Linux: `~/.config/ptool/ptool.toml`
- Windows: `%USERPROFILE%\.config\ptool\ptool.toml`

### 2. 配置 BT 客户端和站点

编辑 `ptool.toml`:

```toml
[[clients]]
name = "local"
type = "qbittorrent"
url = "http://localhost:8080/"
username = "admin"
password = "adminadmin"

[[sites]]
type = "keepfrds"
cookie = "your_cookie_here"
```

### 3. 测试配置

```bash
# 测试站点连接
ptool status keepfrds -t

# 测试客户端连接
ptool status local
```

### 4. 执行刷流

```bash
ptool brush local keepfrds
```

### 5. 设置定时任务

使用 Linux cron job 或 Windows 计划任务定时执行刷流命令。

## 全局参数

- `--config string`: 手动指定配置文件路径
- `-v, -vv, -vvv`: Verbose 模式，输出更多日志

## 获取帮助

```bash
# 查看所有命令
ptool --help

# 查看指定命令帮助
ptool <command> -h
```

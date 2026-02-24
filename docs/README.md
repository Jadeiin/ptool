# ptool 文档

ptool 是一个用于 PT (Private Tracker) 网站和 BitTorrent 客户端的辅助命令行工具。

## 文档索引

- [概述](overview.md) - 项目介绍、主要特性和快速开始
- [配置](config.md) - 配置文件详解（ptool.toml）
- [命令概览](commands.md) - 所有命令的速查表
- [刷流功能](brush.md) - 自动刷流（brush）命令详解
- [辅种功能](xseed.md) - 自动辅种（iyuu/reseed/xseedadd）命令详解
- [客户端控制](client.md) - BT 客户端控制命令集
- [种子工具](torrent.md) - 种子文件相关工具命令
- [站点功能](site.md) - PT 站点相关功能
- [CookieCloud](cookiecloud.md) - CookieCloud 同步功能
- [交互式终端](shell.md) - 交互式 shell 和其他功能

## 快速开始

```bash
# 创建配置文件
ptool config create

# 编辑配置文件后，测试站点连接
ptool status <site> -t

# 执行刷流
ptool brush <client> <site>

# 查看所有命令
ptool --help
```

## 支持的环境

- **操作系统**: Windows / Linux / macOS
- **架构**: x64 / arm64
- **BT 客户端**: qBittorrent v4.1+ / Transmission (<= v3.0)
- **PT 站点**: 绝大部分 NexusPHP 架构站点、M-Team(馒头)等

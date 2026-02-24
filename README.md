# ptool

自用的 PT ([Private trackers][]) 网站和 [BitTorrent][] 客户端辅助工具。提供全自动刷流(brush)、自动辅种(使用 IYUU 或 Reseed 等接口)、BT 客户端控制等功能。

## 主要特性

- 使用 Go 开发的纯 CLI 程序。单文件可执行程序，没有外部依赖。支持 Windows / Linux、x64 / arm64 等多种环境、架构。
- 无状态(stateless)：程序自身不保存任何状态、不在后台持续运行。刷流等任务需要使用 cron job 等方式定时运行本程序。
- 使用简单。只需 5 分钟时间，配置 BitTorrent 客户端地址、PT 网站地址和 cookie 即可开始全自动刷流。
- 目前支持的 BitTorrent 客户端： qBittorrent v4.1+ / Transmission (<= v3.0)。
  - 推荐使用 qBittorrent。Transmission 客户端未充分测试。
- 目前支持的 PT 站点：绝大部分使用 nexusphp 的网站；M-Team(馒头)。
  - 测试过支持的站点：U2、冬樱、红叶、聆音、铂金家、若干不可说的站点等。
  - 未列出的大部分 np 站点应该也支持。除了个别魔改 np 很厉害的站点可能有问题。
  - 支持通过 [CookieCloud][] 自动同步站点 cookie 或导入站点。
- 刷流功能(brush)：
  - 不依赖 RSS。直接抓取站点页面上最新的种子。
  - 无需配置选种规则。自动跳过非免费的和有 HR 的种子；自动筛选适合刷流的种子。
  - 无需配置删种规则。自动删除已无刷流价值的种子；自动删除免费时间到期并且尚未下载完成的种子；硬盘空间不足时也会自动删种。
- 其它 PT 站点功能：搜索种子、批量下载种子、发布种子。
- BitTorrent 客户端控制功能：提供完整的管理、控制 BitTorrent 客户端的功能。
- Torrent 文件 / BitTorrent 协议相关的各种辅助功能。部分命令为[松鼠党][]特别优化，支持与 [rclone][] 整合。
- 自动模仿浏览器访问 PT 站点，能够绕过大多数站点的 CF 盾 (impersonate 特性)。

## 文档

详细文档请查看 [docs](./docs) 目录：

- [概述](./docs/overview.md) - 项目介绍、主要特性和快速开始
- [配置](./docs/config.md) - 配置文件详解（ptool.toml）
- [命令概览](./docs/commands.md) - 所有命令的速查表
- [刷流功能](./docs/brush.md) - 自动刷流（brush）命令详解
- [辅种功能](./docs/xseed.md) - 自动辅种（iyuu/reseed/xseedadd）命令详解
- [客户端控制](./docs/client.md) - BT 客户端控制命令集
- [种子工具](./docs/torrent.md) - 种子文件相关工具命令
- [站点功能](./docs/site.md) - PT 站点相关功能
- [CookieCloud](./docs/cookiecloud.md) - CookieCloud 同步功能
- [交互式终端](./docs/shell.md) - 交互式 shell 和其他功能

## 快速开始

### 下载

- [开发版本](https://ci.appveyor.com/project/sagan/ptool/build/artifacts) (根据默认分支最新代码自动构建)
- [稳定版本](https://github.com/sagan/ptool/releases)

### 创建配置

```bash
ptool config create
```

编辑配置文件（`~/.config/ptool/ptool.toml` 或 `%USERPROFILE%\.config\ptool\ptool.toml`）：

```toml
[[clients]]
name = "local"
type = "qbittorrent"
url = "http://localhost:8080/"
username = "admin"
password = "adminadmin"

[[sites]]
type = "keepfrds"
cookie = "cookie_here"
```

### 执行刷流

```bash
ptool brush local keepfrds
```

### 设置定时任务

Linux (cron):
```bash
*/10 * * * * /path/to/ptool brush local keepfrds
```

Windows (计划任务): 使用 `taskschd.msc` 创建定时任务。

## 获取帮助

```bash
# 查看所有命令
ptool --help

# 查看指定命令帮助
ptool <command> -h
```

## 开源协议

[AGPL-3.0](./LICENSE)

[Private trackers]: https://wiki.installgentoo.com/wiki/Private_trackers
[BitTorrent]: https://en.wikipedia.org/wiki/BitTorrent
[CookieCloud]: https://github.com/easychen/CookieCloud
[IYUU]: https://github.com/ledccn/iyuuplus-dev
[IYUU 接口]: https://doc.iyuu.cn/
[IYUU 网站]: https://iyuu.cn/
[Reseed]: https://github.com/tongyifan/Reseed-backend
[Reseed 官网]: https://reseed.tongyifan.me/
[rclone]: https://github.com/rclone/rclone
[rclone lsjson]: https://rclone.org/commands/rclone_lsjson/
[松鼠党]: https://www.reddit.com/r/DataHoarder/

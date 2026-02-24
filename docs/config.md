# 配置文件

ptool 支持 TOML 或 YAML 格式的配置文件（推荐 TOML）。

## 配置文件位置

1. **推荐位置**:
   - Linux: `~/.config/ptool/ptool.toml`
   - Windows: `%USERPROFILE%\.config\ptool\ptool.toml`

2. **临时测试**: 当前工作目录下的 `ptool.toml`

## 基本结构

```toml
# 全局配置
brushEnableStats = true
iyuuToken = "your_iyuu_token"

# BT 客户端配置
[[clients]]
name = "local"
type = "qbittorrent"
url = "http://localhost:8080/"
username = "admin"
password = "adminadmin"

# PT 站点配置
[[sites]]
type = "keepfrds"
cookie = "cookie_here"

# 站点分组
[[groups]]
name = "acg"
sites = ["u2", "kamept"]

# 命令别名
[[aliases]]
name = "st"
cmd = "status local -t"
```

## BT 客户端配置 ([[clients]])

### 基本配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 客户端名称，用于命令中引用 |
| type | string | 是 | 客户端类型: `qbittorrent` 或 `transmission` |
| url | string | 是 | Web UI 地址 |
| username | string | 是 | Web UI 用户名 |
| password | string | 是 | Web UI 密码 |
| disabled | bool | 否 | 是否禁用 |

### qBittorrent 额外配置

```toml
[[clients]]
name = "local"
type = "qbittorrent"
url = "http://localhost:8080/"
username = "admin"
password = "adminadmin"
# 可选配置
proxy = "http://127.0.0.1:8080"  # 代理服务器
```

### Transmission 额外配置

```toml
[[clients]]
name = "tr"
type = "transmission"
url = "http://localhost:9091/transmission/rpc"
username = "admin"
password = "adminadmin"
localTorrentsPath = "/var/lib/transmission/torrents"  # 本地种子文件路径
```

## PT 站点配置 ([[sites]])

### 方式 1: 使用内置站点类型（推荐）

```toml
[[sites]]
name = "keepfrds"  # 可选，默认使用 type
type = "keepfrds"  # 站点 type 或 alias
cookie = "cookie_here"
```

站点 type 通常为网站域名的主体部分，例如:
- `https://pt.btschool.club/` → `btschool`
- `https://kp.m-team.cc/` → `mteam` 或 `m-team`

运行 `ptool sites` 查看所有内置支持的站点。

### 方式 2: 使用通用架构类型

对于未内置支持的站点:

```toml
[[sites]]
name = "customsite"
type = "nexusphp"  # nexusphp|gazellepw|unit3d|tnode|discuz|mtorrent
url = "https://pt.example.com/"
cookie = "cookie_here"
```

### 站点配置项

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 站点名称 |
| type | string | 站点类型 |
| url | string | 站点首页 URL（方式 2 必填） |
| cookie | string | 站点 Cookie |
| disabled | bool | 是否禁用 |
| proxy | string | 代理服务器 |
| timeout | int | 请求超时时间（秒） |
| flowControlInterval | int | 流量控制间隔（秒） |
| brushTorrentMinSizeLimit | string | 刷流最小种子大小 |
| brushTorrentMaxSizeLimit | string | 刷流最大种子大小 |
| dynamicSeedingSize | string | 动态保种使用空间 |
| dynamicSeedingTorrentMaxSize | string | 动态保种单个种子大小上限 |

### M-Team (馒头) 特殊配置

新版 M-Team 不使用 Cookie 鉴权:

```toml
[[sites]]
type = "mteam"
# 使用 API Key 鉴权（推荐）
apiKey = "your_api_key"
# 或使用用户名密码
# username = "user"
# password = "pass"
```

## 站点分组 ([[groups]])

定义站点分组后，命令中可以使用分组名代替多个站点:

```toml
[[groups]]
name = "acg"
sites = ["u2", "kamept"]
```

使用示例:
```bash
# 在 acg 分组的所有站点中搜索
ptool search acg clannad
```

预置分组 `_all` 表示所有站点。

## 命令别名 ([[aliases]])

```toml
[[aliases]]
name = "st"
cmd = "status local -t"

[[aliases]]
name = "br"
cmd = "brush"
minArgs = 2
defaultArgs = "local keepfrds"
```

- `minArgs`: 执行别名时必须传入的参数数量
- `defaultArgs`: 可选参数的默认值

## CookieCloud 配置 ([[cookieclouds]])

```toml
[[cookieclouds]]
name = "default"
server = "https://cookiecloud.example.com"
uuid = "your_uuid"
password = "your_password"
sites = ["keepfrds"]  # 可选，限制只同步指定站点
```

## 全局配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| brushEnableStats | bool | false | 启用刷流统计 |
| iyuuToken | string | - | IYUU 令牌 |
| reseedUsername | string | - | Reseed 用户名 |
| reseedPassword | string | - | Reseed 密码 |
| siteImpersonate | string | chrome | 模仿的浏览器 |
| shellMaxSuggestions | int | 5 | 交互式终端最大建议数 |
| shellMaxHistory | int | 500 | 交互式终端历史记录数 |

## 完整示例

参考项目中的 `config/ptool.example.toml` 文件。

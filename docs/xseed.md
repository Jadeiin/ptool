# 辅种功能

ptool 支持三种辅种方式:
1. **[IYUU][] 接口**: 使用 [IYUU][] 服务器自动辅种
2. **[Reseed][] 接口**: 使用 [Reseed][] 服务扫描本地文件辅种
3. **手动辅种**: 使用 xseedadd 命令手动添加辅种

## IYUU 自动辅种

iyuu 命令通过 [IYUU 接口][] 提供自动辅种(cross seed)功能。本功能直接访问 IYUU 的服务器，本机上不需要安装 / 运行 IYUU 客户端。

### 配置 IYUU Token

1. 在 [IYUU 网站][] 微信扫码申请 IYUU 令牌（token）
2. 在 ptool.toml 中配置:

```toml
iyuuToken = "IYUU0011223344..."
```

### 绑定 IYUU Token

```bash
ptool iyuu bind --site zhuque --uid 123456 --passkey 0123456789abcdef
```

参数:
- `--site`: 验证站点名（`ptool iyuu sites -b` 查看支持的合作站点）
- `--uid`: PT 站点用户 UID
- `--passkey`: PT 站点 Passkey

### 查看 IYUU 状态

```bash
# 查看 Token 激活状态
ptool iyuu status

# 查看支持的所有站点
ptool iyuu sites

# 查看支持的合作站点（用于绑定）
ptool iyuu sites -b
```

### 执行辅种

```bash
# 为单个客户端辅种
ptool iyuu xseed local

# 为多个客户端辅种
ptool iyuu xseed local vps
```

辅种特性:
- 自动比较文件列表（路径、大小）
- 完全一致才会添加辅种
- 添加的辅种种子打上 `_xseed` 标签
- 默认跳过 hash 校验立即做种

### 常用参数

| 参数 | 说明 |
|------|------|
| `--sites string` | 只辅种指定站点（逗号分隔） |
| `--exclude-sites string` | 排除指定站点 |
| `--max-torrents int` | 最多添加种子数 |
| `--dry-run` | 模拟运行，不实际添加 |

## Reseed 自动辅种

reseed 命令使用 [Reseed][] 提供的接口自动辅种。

### 配置

首先在 [Reseed 官网][] 注册（需要使用指定 PT 站点验证），然后在 ptool.toml 中配置:

```toml
reseedUsername = "username"
reseedPassword = "password"
```

### 使用方式 1（推荐）

适用于客户端与 ptool 运行环境文件系统不同（如 Docker）:

```bash
# 1. 扫描本地文件，查询可辅种种子
ptool reseed match --download "D:\Downloads"

# 2. 添加下载的种子到客户端
ptool xseedadd local "C:\Users\<username>\.config\ptool\reseed\*.torrent"
```

种子默认下载到 ptool.toml 所在目录的 `reseed` 子文件夹。

### 使用方式 2

适用于客户端与 ptool 同机运行:

```bash
# 1. 扫描并下载种子（保存 save_path 到种子 comment 字段）
ptool reseed match --download --use-comment-meta "D:\Downloads"

# 2. (可选) 校验种子与硬盘文件
ptool verifytorrent --check --rename-fail --use-comment-meta "reseed\*.torrent"

# 3. 添加种子到客户端（使用 comment 中的 save_path）
ptool add local --use-comment-meta --skip-check "reseed\*.torrent"
```

### 其他 Reseed 命令

```bash
# 查看 Reseed 账号状态
ptool reseed status

# 查看支持的站点
ptool reseed sites
```

## xseedadd 手动辅种

手动将指定种子作为辅种添加到客户端:

```bash
ptool xseedadd <client> <torrentFileNameOrIdOrUrl>...
```

示例:

```bash
# 添加本地种子文件作为辅种
ptool xseedadd local ./some.torrent

# 添加站点种子作为辅种
ptool xseedadd local mteam.488424
```

工作原理:
1. 在客户端寻找与提供种子元信息完全一致的目标种子
2. 找到匹配目标后，将提供种子作为辅种添加
3. 未找到匹配目标则不会添加
4. 添加的辅种种子打上 `_xseed` 标签

## 防止重复辅种

在 BT 客户端里给不需要辅种的种子打上 `noxseed` 标签，这些种子不会被纳入辅种范围。

## 客户端里已有种子查找可辅种资源

使用方式 1 或 2 的流程，可以扫描客户端里已有种子的内容文件夹，查找可以辅种的站点种子。

[IYUU]: https://github.com/ledccn/iyuuplus-dev
[IYUU 接口]: https://doc.iyuu.cn/
[IYUU 网站]: https://iyuu.cn/
[Reseed]: https://github.com/tongyifan/Reseed-backend
[Reseed 官网]: https://reseed.tongyifan.me/

# 种子文件工具

ptool 提供多种种子文件(.torrent)相关的辅助工具。

## parsetorrent 命令

显示种子文件元信息:

```bash
ptool parsetorrent <torrentFileNameOrIdOrUrl>...
```

示例:

```bash
# 显示本地种子信息
ptool parsetorrent file.torrent

# 显示站点种子信息
ptool parsetorrent mteam.488424

# 批量显示
ptool parsetorrent *.torrent
```

参数:

| 参数 | 说明 |
|------|------|
| `--show-info-hash-only` | 只显示 info hash |
| `--all` | 显示更多详情（包括 pieces hash） |
| `--sum` | 显示汇总统计信息 |
| `--json` | JSON 格式输出 |
| `--format string` | 自定义格式 |

## verifytorrent 命令

校验种子文件与硬盘内容是否一致:

```bash
ptool verifytorrent <torrentFileNameOrIdOrUrl>...
```

必选参数（4选1）:

| 参数 | 说明 |
|------|------|
| `--save-path <path>` | 种子内容保存路径（可校验多个种子） |
| `--content-path <path>` | 单个种子的内容路径 |
| `--use-comment-meta` | 使用种子 comment 字段的 save_path |
| `--rclone-lsjson-file <file>` | [rclone][] 的 `rclone lsjson --recursive <path>` 命令输出 |
| `--rclone-save-path <path>` | rclone 远程路径 |

其他参数:

| 参数 | 说明 |
|------|------|
| `--check` | 完整 hash 校验 |
| `--check-quick` | 快速 hash 校验（只校验首尾 piece） |
| `--rename-fail` | 校验失败的种子重命名为 `.torrent.fail` |

示例:

```bash
# 文件元信息对比（文件名、大小）
ptool verifytorrent file.torrent --save-path D:\Downloads

# 完整 hash 校验
ptool verifytorrent file.torrent --save-path D:\Downloads --check

# 校验单个种子
ptool verifytorrent MyTorrent.torrent --content-path D:\Downloads\MyTorrent --check

# 配合 [rclone][] 校验云存储
ptool verifytorrent *.torrent --rclone-save-path remote:Downloads --check
```

## maketorrent 命令

制作种子文件:

```bash
ptool maketorrent <content-path>
```

示例:

```bash
# 基本用法
ptool maketorrent ./MyVideos

# 指定输出文件名
ptool maketorrent ./MyVideos --out MyVideos.torrent

# 公开种子（添加公共 tracker）
ptool maketorrent ./MyVideos --public

# PT 私有种子的 tracker
ptool maketorrent ./MyVideos --private --tracker "https://tracker.example.com/announce"
```

参数:

| 参数 | 说明 |
|------|------|
| `--out string` | 输出文件名 |
| `--public` | 添加公共 tracker |
| `--private` | 标记为私有种子 |
| `--tracker string` | 手动指定 tracker 地址 |
| `--piece-length int` | 指定 piece 大小（KB） |

说明:
- 内容文件夹中的临时/隐藏文件（`.*`, `*.tmp`, `Thumbs.db` 等）默认会被忽略

## edittorrent 命令

编辑种子文件:

```bash
ptool edittorrent [flags] <torrent-files>...
```

示例:

```bash
# 批量修改 tracker
ptool edittorrent --update-tracker "https://new.tracker/" *.torrent

# 修改 info.source（会改变 info-hash）
ptool edittorrent --update-info-source "MySite" *.torrent

# 替换 tracker host
ptool edittorrent --replace-tracker-host old.com new.com *.torrent

# 查看帮助了解更多
ptool edittorrent -h
```

## dltorrent 命令

下载站点的种子文件:

```bash
ptool dltorrent <torrentIdOrUrl>...
```

示例:

```bash
# 下载站点种子
ptool dltorrent mteam.488424

# 批量下载
ptool dltorrent mteam.488424 mteam.488425

# 指定保存目录
ptool dltorrent mteam.488424 --download-dir ./torrents
```

## findalone 命令

查找下载目录里的未做种文件:

```bash
ptool findalone <client> <save-path>...
```

示例:

```bash
# 检查多个下载目录
ptool findalone local D:\Downloads E:\Downloads

# Docker 环境（需要路径映射）
ptool findalone local --map-save-path "/root/Downloads:/Downloads" /root/Downloads

# 显示所有文件及种子关联数
ptool findalone local D:\Downloads --all

# 删除未做种文件
ptool findalone local D:\Downloads --delete-alone

# 移动未做种文件
ptool findalone local D:\Downloads --move-alone-to D:\AloneFiles
```

## partialdownload 命令

拆包下载（用于 VPS 等硬盘空间有限的场景）:

该命令的设计目的不是用于刷流。而是用于使用 VPS 等硬盘空间有限的云服务器(分多次)下载体积非常大的单个种子，然后配合 [rclone][] 将下载的文件直接上传到云存储。

参考 [rclone lsjson][] 命令的文档。

```bash
ptool partialdownload <client> <infoHash> [flags]
```

使用步骤:

```bash
# 1. 先将种子以暂停状态添加到客户端

# 2. 查看分片信息（按 1TiB 切分）
ptool partialdownload local <infohash> --chunk-size 1TiB -a

# 3. 设置只下载第 N 块
ptool partialdownload local <infohash> --chunk-size 1TiB --chunk-index 0

# 4. 下载完成后上传文件到云存储
# 5. 重复步骤 3-4 直到下载完成
```

其他用法:

```bash
# 跳过特定文件
ptool partialdownload local <infohash> --exclude "*.txt"

# 只下载特定文件
ptool partialdownload local <infohash> --include "*.mkv"
```

## hardlink 命令

硬链接辅助工具。

### cp 子命令

创建目录硬链接（类似 `cp -rl`）:

```bash
ptool hardlink cp SOURCE DEST
```

参数:

| 参数 | 说明 |
|------|------|
| `--hardlink-min-size string` | 小于此值的文件直接复制而非硬链接 |

### torrent 子命令

根据种子文件定向硬链:

```bash
ptool hardlink torrent MyTorrent.torrent --content-path ./Contents --link-save-path ./Downloads
```

功能说明:
- 在 content path 查找种子里的内容文件
- 支持文件名、路径完全不同的匹配
- 在 link-save-path 生成符合种子结构的内容（可直接辅种）

参考:
- [TorrentHardLinkHelper 教程](https://tieba.baidu.com/p/5572480043)

## movesavepath 命令

修改种子内容文件保存路径:

```bash
ptool movesavepath --client <client> <old-save-path> <new-save-path>
```

示例:

```bash
# 基本用法
ptool movesavepath --client local /root/Downloads /var/Downloads

# Docker 环境（路径映射）
ptool movesavepath --client local /root/Downloads/Uncategoried /root/Downloads/Others --map-save-path "/root/Downloads:/Downloads"
```

说明:
- 将整个内容文件夹整体移动
- 先导出并删除种子、移动文件、再重新添加种子
- 解决 qBittorrent "Set location" 无法移动非种子文件的问题

## transfertorrent 命令

转移种子做种客户端:

```bash
ptool transfertorrent <src-client> --dst-client <dst-client> [<infoHash>...]
```

示例:

```bash
# 转移所有做种种子
ptool transfertorrent local --dst-client vps _seeding

# 转移后删除源种子
ptool delete local --tag _transferred --preserve
```

说明:
- 成功转移的种子打上 `_transferred` 标签
- 源客户端和目标客户端需要在同一机器
- 不同 Docker 容器需要使用 `--map-save-path` 指定路径映射
- Transmission 源客户端支持有限

[rclone]: https://github.com/rclone/rclone
[rclone lsjson]: https://rclone.org/commands/rclone_lsjson/

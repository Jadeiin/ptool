# BT 客户端控制

ptool 提供完整的 BT 客户端管理功能。

## status 命令

显示 BT 客户端或 PT 站点状态:

```bash
ptool status <clientOrSite>...
```

示例:

```bash
# 查看客户端状态
ptool status local

# 查看站点状态
ptool status keepfrds

# 查看多个
ptool status local keepfrds

# 显示种子列表（客户端：当前活动种子；站点：最新种子）
ptool status local -t

# 显示完整种子列表
ptool status local -t -f
```

## clientctl 命令

读取或修改 BT 客户端配置:

```bash
ptool clientctl <client> [<option>[=value] ...]
```

### 支持的配置项

| 配置项 | 说明 |
|--------|------|
| `global_download_speed_limit` | 全局下载速度上限 |
| `global_upload_speed_limit` | 全局上传速度上限 |
| `global_download_speed` | (只读) 当前下载速度 |
| `global_upload_speed` | (只读) 当前上传速度 |
| `free_disk_space` | (只读) 默认下载目录剩余空间 |
| `save_path` | 默认下载目录 |
| `qb_*` | qBittorrent 所有配置项 |
| `tr_*` | Transmission 所有配置项 |

示例:

```bash
# 获取所有配置
ptool clientctl local

# 设置上传速度限制为 10MiB/s
ptool clientctl local global_upload_speed_limit=10M
```

## show 命令

显示客户端种子信息:

```bash
ptool show <client> [flags] [<infoHash>...]
```

示例:

```bash
# 显示所有种子
ptool show local _all

# 显示特定分类的种子
ptool show local --category brush _all

# JSON 格式输出
ptool show local _all --json

# 自定义输出格式
ptool show local _all --format "{{.InfoHash}} - {{.SavePath}}"

# 显示单个种子详情
ptool show local <infohash>

# 只输出 infohash 列表（用于管道）
ptool show local --category rss --completed-before 5d --show-info-hash-only
```

## add 命令

添加种子到客户端:

```bash
ptool add <client> <torrentFileNameOrIdOrUrl>...
```

参数支持:
- 本地种子文件（支持 `*` 通配符）: `*.torrent`
- 站点种子 ID: `mteam.488424`
- 站点种子 URL: `https://kp.m-team.cc/details.php?id=488424`
- 磁力链接: `magnet:?xt=urn:btih:...`

示例:

```bash
# 添加本地所有种子
ptool add local *.torrent

# 添加站点种子
ptool add local mteam.488424
ptool add local "https://kp.m-team.cc/download.php?id=488424"

# 从 stdin 读取
ptool show local --category rss --show-info-hash-only | ptool add other -
```

常用参数:

| 参数 | 说明 |
|------|------|
| `--category string` | 设置分类 |
| `--tag string` | 添加标签 |
| `--save-path string` | 设置保存路径 |
| `--skip-check` | 跳过 hash 校验 |
| `--paused` | 添加后暂停 |
| `--ratio-limit float` | 分享比例限制 |
| `--seeding-time-limit int` | 做种时间限制（秒） |

## pause/resume 命令

暂停/恢复种子:

```bash
ptool pause <client> [flags] [<infoHash>...]
ptool resume <client> [flags] [<infoHash>...]
```

示例:

```bash
# 暂停所有种子
ptool pause local _all

# 暂停指定分类的下载中种子
ptool pause local --category abc _downloading

# 恢复所有种子
ptool resume local _all
```

## delete 命令

删除种子:

```bash
ptool delete <client> [<infoHash>...]
```

示例:

```bash
# 删除指定种子（会提示确认）
ptool delete local 31a615d5984cb63c6f999f72bb3961dce49c194a

# 强制删除（不提示）
ptool delete local --force <infohash>

# 删除并保留文件
ptool delete local --preserve <infohash>

# 结合 show 命令删除符合条件的种子
ptool show local --category rss --completed-before 5d --show-info-hash-only | ptool delete local --force -
```

## reannounce/recheck 命令

```bash
# 强制立即汇报
ptool reannounce <client> [<infoHash>...]

# 强制重新校验
ptool recheck <client> [<infoHash>...]
```

## 分类管理

```bash
# 获取所有分类
ptool getcategories <client>

# 创建分类（可指定保存路径）
ptool createcategory <client> <category> --save-path "/root/downloads"

# 删除分类
ptool deletecategories <client> <category>...

# 修改种子分类
ptool setcategory <client> <category> <infoHashes>...
```

## 标签管理

```bash
# 获取所有标签
ptool gettags <client>

# 创建标签
ptool createtags <client> <tags>...

# 删除标签
ptool deletetags <client> <tags>...

# 添加标签到种子
ptool addtags <client> <tags> <infoHashes>...

# 从种子移除标签
ptool removetags <client> <tags> <infoHashes>...

# 重命名标签
ptool renametag <client> <old-tag> <new-tag>

# 检查标签是否存在（存在返回 0）
ptool checktag <client> <tag>
```

## Tracker 管理

```bash
# 修改 Tracker（旧 tracker 存在才修改）
ptool edittracker <client> _all --old-tracker "https://old.tracker/" --new-tracker "https://new.tracker/"

# 只替换 host 部分
ptool edittracker <client> _all --old-tracker old-tracker.com --new-tracker new-tracker.com --replace-host

# 添加 Tracker
ptool addtrackers <client> <infoHashes...> --tracker "https://..."

# 移除 Tracker
ptool removetrackers <client> <infoHashes...> --tracker "https://..."

# 标记 Tracker 异常种子
ptool markinvalidtracker <client> _all
```

markinvalidtracker 会打上以下标签:
- `_invalid_tracker_not_exist`: 种子未注册或已删除
- `_invalid_tracker_invalid_auth`: Passkey 不正确
- `_invalid_tracker_violate_rule`: 违反做种规则

## 其他命令

```bash
# 修改种子保存路径
ptool setsavepath <client> <savePath> [<infoHash>...]

# 综合修改种子（分类/路径/标签等）
ptool modifytorrent <client> --set-category <cat> --set-save-path <path> --add-tags <tags>

# 导出种子
ptool export <client> <infoHash>...

# 设置只下载元数据
ptool skipchecking <client> <infoHashes>...
```

### export 命令特殊参数

`--use-comment-meta`: 将种子的分类、标签、保存路径等元信息保存到导出种子的 comment 字段，配合 `ptool add` 的同名参数可恢复这些信息。

### modifytorrent 参数

| 参数 | 说明 |
|------|------|
| `--set-category string` | 设置分类 |
| `--set-save-path string` | 设置保存路径 |
| `--add-tags string` | 添加标签 |
| `--remove-tags string` | 移除标签 |
| `--ratio-limit float` | 分享比例限制 |
| `--seeding-time-limit int` | 做种时间限制（秒） |

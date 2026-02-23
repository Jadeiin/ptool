# 站点功能

ptool 提供多种 PT 站点相关的功能。

## search 命令

搜索 PT 站点种子:

```bash
ptool search <sites> <keyword>
```

示例:

```bash
# 搜索单个站点
ptool search keepfrds clannad

# 搜索多个站点
ptool search keepfrds,mteam clannad

# 搜索所有站点
ptool search _all clannad

# 搜索站点分组
ptool search acg clannad
```

## batchdl 命令

批量下载站点种子（别名: `ebookgod`）:

```bash
ptool batchdl <site> [flags]
```

基本用法:

```bash
# 显示找到的种子列表
ptool batchdl kamept

# 下载种子到当前目录
ptool batchdl kamept --download

# 直接添加到 BT 客户端
ptool batchdl kamept --add-client local
```

常用参数:

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--max-torrents int` | 最多下载数量 | -1（无限制） |
| `--sort string` | 排序方式: size/time/name/seeders/leechers/snatched/none | size |
| `--order string` | 排序顺序: asc/desc | asc |
| `--min-torrent-size string` | 最小种子大小 | -1 |
| `--max-torrent-size string` | 最大种子大小 | -1 |
| `--max-total-size string` | 总体积上限 | -1 |
| `--free` | 只下载免费种子 | false |
| `--no-hr` | 跳过 HR 种子 | false |
| `--no-paid` | 跳过分种子 | false |
| `--base-url string` | 自定义列表页 URL | - |
| `--start-page string` | 起始页码 | 0 |
| `--one-page` | 只抓取 1 页 | false |
| `--add-category-auto` | 自动设置分类为站点名 | false |

实际场景示例:

```bash
# 获取首页最新特定分类的免费种子
ptool batchdl kamept --tag "外语音声,同人志" --sort none --start-page 0 --free --one-page --add-client local --add-category-auto
```

## publish 命令

自动发布种子到 PT 站点:

```bash
ptool publish --site <site> --client <client> --save-path <path> [flags]
```

示例:

```bash
ptool publish --site kamept --client local --check-existing --save-path /downloads/i
```

执行流程:

1. 检测文件夹里的 `metadata.nfo` 文件
2. 读取元信息生成上传表单
3. 检测站点是否已存在相同内容
4. 制作种子（保存为 `.torrent`）
5. 上传封面图到图床（如果有 `cover.*`）
6. 发布种子到站点
7. 下载发布后种子（保存为 `.<sitename>.torrent`）
8. 添加到 BT 客户端

### metadata.nfo 格式

```yaml
---
title: 作品标题
author: 作者
narrator: 演播
tags: 标签1, 标签2
number: 番号
source: 来源
---

作品描述...
```

说明:
- 使用 YAML Front Matter 格式
- `number` 字段用于检测重复
- 模板渲染字段: `name`（标题）、`desc`（描述）

### 配置发布参数

对于需要额外字段的站点，在 ptool.toml 中配置:

```toml
[uploadTorrentAdditionalPayload]
type = """
{% if "分类1" in tags %}100
{% elif "分类2" in tags %}200
{% else %}999
{% endif %}
"""
```

### 图床配置

```toml
[[sites]]
type = "custom"
imageUploadUrl = 'https://pic.example.com/'
#imageUploadFileField = 'file'
#imageUploadResponseUrlField = 'url'
#imageUploadPayload = ''
```

## sites 命令

查看内置支持的 PT 站点:

```bash
# 列出所有内置站点
ptool sites

# 显示站点详细配置
ptool sites show mteam
```

## 站点种子信息显示

`status -t`、`batchdl`、`search` 等命令显示的站点种子列表包含以下字段:

| 字段 | 说明 |
|------|------|
| Name | 种子名称 |
| Size | 种子大小 |
| Free | 免费属性（见下表） |
| Time | 发布时间 |
| ↑S | 做种人数 |
| ↓L | 下载人数 |
| ✓C | 已完成人数 |
| ID | 种子 ID（格式: 站点名.种子ID） |
| P | 下载进度/历史状态 |

### Free 字段符号说明

| 符号 | 说明 |
|------|------|
| `2.0` | 上传量计算倍率 |
| `!` | HR 种子 |
| `$` | 付费种子 |
| `✓` | 免费种子 |
| `✕` | 非免费种子 |
| `(1d12h)` | 优惠剩余时间 |
| `N` | 中性种子 |
| `Z` | 零流量种子 |

### P 字段说明

| 值 | 说明 |
|----|----|
| `-` | 未曾下载 |
| `✓` | 曾下载或做种 |
| `*%` | 当前正在下载或做种 |

## 站点分组

在 ptool.toml 中定义分组:

```toml
[[groups]]
name = "acg"
sites = ["u2", "kamept"]
```

使用分组:

```bash
# 在分组的所有站点中搜索
ptool search acg clannad

# 刷流使用分组
ptool brush local acg
```

预置 `_all` 分组表示所有站点。

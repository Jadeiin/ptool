# 命令概览

ptool 的所有功能通过命令参数区分，基本格式:

```
ptool <command> [args...] [flags]
```

## 命令分类速查

### 刷流与辅种

| 命令 | 说明 |
|------|------|
| [brush](brush.md) | 自动刷流 |
| [iyuu](xseed.md) | 使用 IYUU 接口自动辅种 |
| [reseed](xseed.md) | 使用 Reseed 接口自动辅种 |
| [xseedadd](xseed.md) | 手动添加辅种种子 |
| [dynamicseeding](brush.md) | 全站动态保种 |

### BT 客户端控制

| 命令 | 说明 |
|------|------|
| [status](client.md) | 显示客户端状态 |
| [clientctl](client.md) | 读取/修改客户端配置 |
| [show](client.md) | 显示种子信息 |
| [add](client.md) | 添加种子 |
| [pause/resume](client.md) | 暂停/恢复种子 |
| [delete](client.md) | 删除种子 |
| [reannounce](client.md) | 强制汇报 |
| [recheck](client.md) | 强制检测 Hash |
| [export](client.md) | 导出种子 |

### 分类与标签管理

| 命令 | 说明 |
|------|------|
| [getcategories](client.md) | 获取所有分类 |
| [createcategory](client.md) | 创建分类 |
| [deletecategories](client.md) | 删除分类 |
| [setcategory](client.md) | 设置种子分类 |
| [gettags](client.md) | 获取所有标签 |
| [createtags](client.md) | 创建标签 |
| [deletetags](client.md) | 删除标签 |
| [addtags](client.md) | 添加标签 |
| [removetags](client.md) | 移除标签 |
| [renametag](client.md) | 重命名标签 |
| [checktag](client.md) | 检查标签是否存在 |

### Tracker 管理

| 命令 | 说明 |
|------|------|
| [edittracker](client.md) | 修改 Tracker |
| [addtrackers](client.md) | 添加 Tracker |
| [removetrackers](client.md) | 移除 Tracker |
| [markinvalidtracker](client.md) | 标记 Tracker 异常种子 |

### 种子文件工具

| 命令 | 说明 |
|------|------|
| [parsetorrent](torrent.md) | 显示种子信息 |
| [verifytorrent](torrent.md) | 校验种子与硬盘内容 |
| [maketorrent](torrent.md) | 制作种子 |
| [edittorrent](torrent.md) | 编辑种子文件 |
| [dltorrent](torrent.md) | 下载站点种子 |

### 站点功能

| 命令 | 说明 |
|------|------|
| [search](site.md) | 搜索种子 |
| [batchdl](site.md) | 批量下载种子 |
| [publish](site.md) | 发布种子 |
| [sites](site.md) | 查看内置站点列表 |

### 其他工具

| 命令 | 说明 |
|------|------|
| [stats](brush.md) | 刷流流量统计 |
| [findalone](torrent.md) | 查找未做种文件 |
| [partialdownload](torrent.md) | 拆包下载 |
| [movesavepath](client.md) | 修改保存路径 |
| [transfertorrent](client.md) | 转移种子客户端 |
| [hardlink](torrent.md) | 硬链接工具 |
| [cookiecloud](cookiecloud.md) | CookieCloud 同步 |
| [shell](shell.md) | 交互式终端 |
| [config](config.md) | 配置文件管理 |
| [version](shell.md) | 版本信息 |
| [alias](shell.md) | 执行命令别名 |

## 全局参数

| 参数 | 说明 |
|------|------|
| `--config string` | 指定配置文件路径 |
| `-v` | 输出更多日志信息 |
| `-vv` | 输出更详细的日志 |
| `-vvv` | 输出最详细的日志 |

## 批量选择种子的特殊参数

很多命令支持使用特殊值选择多个种子:

| 特殊值 | 说明 |
|--------|------|
| `_all` | 所有种子 |
| `_done` | 已下载完成的种子 |
| `_undone` | 未下载完成的种子 |
| `_active` | 当前活动的种子 |
| `_error` | 出错的种子 |
| `_downloading` | 正在下载的种子 |
| `_seeding` | 正在做种的种子 |
| `_paused` | 暂停的种子 |
| `_completed` | 已完成但未做种的种子 |

## 筛选参数

| 参数 | 说明 |
|------|------|
| `--category string` | 指定分类 |
| `--tag string` | 指定标签（逗号分隔多个） |
| `--filter string` | 种子名称包含的文字 |

## 命令组合示例

```bash
# 删除已下载完成超过 5 天的 RSS 分类种子
ptool show local --category rss --completed-before 5d --show-info-hash-only | ptool delete local --force -

# 暂停所有下载中的种子
ptool pause local _downloading

# 为所有做种中的种子添加标签
ptool addtags local mytag _seeding
```

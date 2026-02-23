# 刷流功能

刷流功能是 ptool 的核心功能，自动从 PT 站点获取种子添加到 BT 客户端，并自动删除旧的无价值种子。

## brush 命令

```bash
ptool brush <client> <site>... [flags]
```

### 参数

- `<client>`: BT 客户端名称
- `<site>`: PT 站点名称（可指定多个）

### 示例

```bash
# 单站点刷流
ptool brush local keepfrds

# 多站点刷流
ptool brush local keepfrds mteam

# 重复站点名增加权重
ptool brush local keepfrds keepfrds mteam
```

## 工作原理

### 选种规则（添加新种子）

刷流任务会按以下规则筛选种子:

- ✗ 不免费种子
- ✗ 存在 HnR 考查的种子
- ✗ 免费时间临近截止的种子
- ✗ "付费"种子（下载或汇报时扣除积分）
- ✗ 发布时间过久的种子
- ✓ 综合考虑做种/下载人数、种子大小等因素

### 删种规则（删除旧种子）

- 免费时间临近截止且未下载完成 → 删除或停止下载
- 硬盘空间不足（默认保留 5GiB）→ 删除无上传速度的种子
- 长时间无上传速度或上传/下载比例过低 → 删除

## 刷流种子管理

- 刷流种子默认放到 `_brush` 分类
- 程序**只管理** `_brush` 分类里的种子
- 更改种子分类可防止被自动删除
- 客户端里存在 `_noadd` 标签时，刷流不会添加新种子

## 常用参数

| 参数 | 说明 |
|------|------|
| `--max-downloading-torrents int` | 最大同时下载种子数 |
| `--max-torrents int` | 刷流任务种子总数上限 |
| `--min-disk-space string` | 最小磁盘剩余空间 |

## stats 命令

显示刷流流量统计:

```bash
# 启用统计（在 ptool.toml 中添加）
brushEnableStats = true

# 查看统计
ptool stats [client...]
```

统计信息包括:
- 下载流量总和
- 上传流量总和
- 仅统计 `_brush` 分类的种子

## dynamicseeding 命令（动态保种）

自动下载亟需保种的种子:

```bash
ptool dynamicseeding <client> <site>
```

### 配置

```toml
[[sites]]
type = "kamept"
cookie = "..."
dynamicSeedingSize = "500GiB"              # 动态保种使用空间
dynamicSeedingTorrentMaxSize = "20GiB"     # 单个种子大小上限
```

### 工作原理

1. 可用空间 = dynamicSeedingSize - 当前动态保种种子总大小
2. 可用空间充足时 → 下载亟需保种（有断种风险）的种子
3. 可用空间不足且有新的亟需保种种子 → 删除不再有风险的老种子
4. 做种人数 < 4 的种子不会被自动删除
5. 带有 `nodel` 标签的种子不会被自动删除

动态保种种子特征:
- 分类: `dynamic-seeding-<sitename>`
- 标签: `site:<sitename>`

## 定时任务设置

### Linux (cron)

```bash
# 每 10 分钟执行一次刷流
*/10 * * * * /path/to/ptool brush local keepfrds

# 每 30 分钟执行一次动态保种
*/30 * * * * /path/to/ptool dynamicseeding local kamept
```

### Windows (计划任务)

1. 打开任务计划程序 (`taskschd.msc`)
2. 创建基本任务
3. 设置触发器（例如每 10 分钟一次）
4. 设置操作: 启动程序 `ptool.exe`
5. 添加参数: `brush local keepfrds`

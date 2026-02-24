# 其他功能

## shell 命令

启动交互式终端环境:

```bash
ptool shell
```

特性:

- 支持所有 ptool 命令
- 完整的命令和参数自动补全
- 支持动态内容补全（客户端名、站点名等）
- 历史记录功能

### 与系统 shell 的区别

ptool 也支持 bash、powershell 等系统 shell 的自动补全:

```bash
# 查看补全脚本安装方法
ptool completion
```

但系统 shell 的补全功能有限，不支持动态内容（如客户端名、站点名）。

### 配置

```toml
shellMaxSuggestions = 5      # 最大建议数
shellMaxHistory = 500        # 历史记录数
```

## version 命令

显示版本信息:

```bash
ptool version

# 查看模仿浏览器详情
ptool version --show-impersonate chrome120
```

## alias 命令

执行命令别名:

```bash
ptool alias <name> [args...]
```

配置示例（ptool.toml）:

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

使用:

```bash
# 执行 st 别名
ptool alias st
# 等效于: ptool status local -t

# 执行 br 别名
ptool alias br
# 等效于: ptool brush local keepfrds

# 带参数
ptool alias br vps mteam
# 等效于: ptool brush vps mteam
```

注意:
- 别名无法覆盖内置命令
- 别名不能在系统 shell 直接使用，需通过 `ptool alias` 执行

## config 命令

配置文件管理:

```bash
# 创建配置文件
ptool config create

# 显示当前配置
ptool config show
```

## 模仿浏览器 (impersonate)

ptool 访问站点时会自动模拟浏览器环境:

- TLS ja3 指纹
- HTTP2 akamai_fingerprint 指纹
- HTTP headers

能够绕过大多数站点的 CF 盾。

### 配置

```toml
siteImpersonate = "chrome120"    # 设置模仿的浏览器
```

运行 `ptool version` 查看支持的浏览器列表。

### 查看浏览器详情

```bash
ptool version --show-impersonate chrome120
```

## 常用别名推荐

```toml
[[aliases]]
name = "st"
cmd = "status -t"
minArgs = 0
defaultArgs = "local"

[[aliases]]
name = "br"
cmd = "brush"
minArgs = 0
defaultArgs = "local keepfrds"

[[aliases]]
name = "xseed"
cmd = "iyuu xseed"
minArgs = 0
defaultArgs = "local"

[[aliases]]
name = "adds"
cmd = "add"
minArgs = 0
defaultArgs = "local"
```

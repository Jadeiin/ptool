# CookieCloud 功能

ptool 支持通过 [CookieCloud](https://github.com/easychen/CookieCloud) 服务器同步站点 Cookies 或导入站点信息。

## 配置

在 ptool.toml 中添加 CookieCloud 连接信息:

```toml
[[cookieclouds]]
name = "default"                    # 可选，名称
server = "https://cookiecloud.example.com"
uuid = "your_uuid"
password = "your_password"
# sites = ["keepfrds"]              # 可选，限制只同步指定站点
# proxy = "http://127.0.0.1:8080"   # 可选，代理服务器
```

可以添加任意多个 CookieCloud 连接。

## status 子命令

测试 CookieCloud 服务连接:

```bash
ptool cookiecloud status
```

此命令会测试所有配置的 CookieCloud 连接，验证配置正确性和服务器状态。

## sync 子命令

同步站点 Cookies:

```bash
ptool cookiecloud sync
```

工作原理:

1. 从 CookieCloud 服务器获取最新 Cookies
2. 测试 ptool.toml 中站点的当前 Cookie 和新 Cookie
3. **只在当前 Cookie 失效且新 Cookie 有效时**更新
4. 自动修改 ptool.toml 文件

说明:
- 不会更新所有站点的 Cookie，只更新需要更新的
- 不会添加新站点，只更新已有配置的站点

## import 子命令

导入站点:

```bash
ptool cookiecloud import
```

工作原理:

1. 从 CookieCloud 获取 Cookies
2. 筛选出内置支持但尚未配置的站点
3. 筛选出存在有效 Cookie 的站点
4. 自动添加这些站点的配置到 ptool.toml

说明:
- 不会更新已有站点的 Cookie
- 只会添加当前未配置的新站点

## get 子命令

查看 CookieCloud 中的网站 Cookie:

```bash
ptool cookiecloud get <site>...
```

示例:

```bash
# 查看指定站点
ptool cookiecloud get keepfrds

# 查看多个站点
ptool cookiecloud get keepfrds mteam

# 支持域名或 URL
ptool cookiecloud get pt.keepfrds.com
ptool cookiecloud get "https://pt.keepfrds.com/"
```

输出格式:

- 默认: HTTP "Cookie" 头格式
- `--format js`: JavaScript 代码格式，可直接在浏览器 Console 执行导入

```bash
# 以 JavaScript 格式输出
ptool cookiecloud get keepfrds --format js
```

## 完整工作流示例

```bash
# 1. 测试 CookieCloud 连接
ptool cookiecloud status

# 2. 导入新站点（首次使用）
ptool cookiecloud import

# 3. 设置定时任务同步 Cookies
# Linux cron: 每天同步一次
0 */6 * * * /path/to/ptool cookiecloud sync

# 4. 定期执行刷流等任务
*/10 * * * * /path/to/ptool brush local keepfrds
```

## 注意事项

1. **安全性**: CookieCloud 密码请妥善保管
2. **同步频率**: 建议每 6-12 小时同步一次，过于频繁可能被站点限制
3. **站点配置**: import 导入的站点可能需要进一步配置（如代理、超时等）
4. **Cookie 有效期**: 部分站点 Cookie 有效期较短，需要配合浏览器插件自动同步到 CookieCloud

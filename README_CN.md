# mcpkit

> **MCP 开发者的瑞士军刀。**
> 用一个快速、零依赖的二进制文件，测试、扫描、基准测试和探索你的 MCP 服务器。

[English](./README.md) | 中文

```
$ mcpkit test --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

🛠  mcpkit test — MCP 服务器协议合规测试
  Server: filesystem@0.1.0

ID       Check                           Status  Time   Detail
HND-001  Initialize succeeded            ✓ PASS  0µs    server returned initialize result
HND-002  Protocol version reported       ✓ PASS  0µs    protocolVersion: 2025-11-25
HND-003  Server info present             ✓ PASS  0µs    filesystem@0.1.0
TL-001   tools/list returns valid        ✓ PASS  514µs  server returned 2 tools
TL-002   Tool name format                ✓ PASS  0µs    name conforms to spec
...
```

## ✨ 为什么选择 mcpkit？

MCP 生态已有 **16,000+ 服务器**和 **1.5 亿+ SDK 下载量**——但缺少一个 Go 原生的、单二进制的开发者工具链。

**mcpkit 填补了这个空白。** 它是 MCP 协议缺失的 CLI 工具，就像 `curl` 是通用 HTTP 工具，`psql` 是 PostgreSQL 标准客户端一样。

## 🚀 快速开始

```bash
# 安装（Go 用户）
go install github.com/justcodeit404/mcpkit/cmd/mcpkit@latest

# 交互式探索服务器
mcpkit probe --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# 运行协议合规测试
mcpkit test --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# 安全扫描
mcpkit scan --command "npx -y @modelcontextprotocol/server-filesystem /tmp"

# 性能基准测试
mcpkit bench --command "./my-server" --method ping -n 1000
```

## 🧰 命令列表

| 命令 | 说明 |
|------|------|
| `mcpkit probe` | 交互式 REPL，探索 MCP 服务器 |
| `mcpkit test` | 协议合规测试（20 项检查） |
| `mcpkit scan` | 安全漏洞扫描（10 条规则，2 个等级） |
| `mcpkit bench` | 性能基准测试（含百分位统计） |
| `mcpkit fuzz` | 协议模糊测试（v0.3.0 开发中） |
| `mcpkit new` | 脚手架生成新 MCP 服务器（v0.2.0 开发中） |
| `mcpkit validate` | 验证 `mcp.json` 配置文件（v0.2.0 开发中） |

## ⚔️ 对比

| 功能 | MCP Inspector | mcp-server-doctor | MCPLint | **mcpkit** |
|------|:---:|:---:|:---:|:---:|
| 语言 | Node.js | Node.js | Rust | **Go** |
| 单二进制 | ❌ | ❌ | ✅ | ✅ |
| 交互式 REPL | ⚠️ Web UI | ❌ | ❌ | ✅ |
| 协议合规测试 | ⚠️ 部分 | ⚠️ 部分 | ✅ | ✅ |
| 安全扫描 | ❌ | ❌ | ⚠️ | ✅ |
| 性能基准测试 | ❌ | ⚠️ 基础 | ❌ | ✅ |
| 协议模糊测试 | ❌ | ❌ | ❌ | ✅ (v0.3.0) |
| CI/CD JSON 输出 | ⚠️ 部分 | ⚠️ | ❌ | ✅ |
| 跨平台 | ⚠️ 有限 | ⚠️ | ✅ | ✅ |
| 无需 npm/Node | ❌ | ❌ | ✅ | ✅ |

## 📦 安装

```bash
# Go install (需要 Go 1.23+)
go install github.com/justcodeit404/mcpkit/cmd/mcpkit@latest

# 直接下载二进制
# 见 https://github.com/justcodeit404/mcpkit/releases/latest

# macOS / Linux
curl -fsSL https://github.com/justcodeit404/mcpkit/releases/latest/download/mcpkit_linux_amd64.tar.gz | tar xz
sudo mv mcpkit /usr/local/bin/
```

## 🎯 mcpkit 的优势

- **零依赖** — 单个静态链接二进制，无需 Node.js、Python 或 npm
- **极速** — Go 编写，启动瞬间完成，微秒级基准测试精度
- **精美终端输出** — 使用 Charmbracelet lipgloss 设计，终端里也能赏心悦目
- **CI/CD 原生支持** — JSON 输出，可直接用于 GitHub Actions、GitLab CI、Jenkins
- **跨平台** — 同一份源码构建 Windows、macOS、Linux

## 🔍 `mcpkit test` 检查项（v0.1.0）

| 检查 ID | 类别 | 检查内容 |
|---------|------|---------|
| HND-001..005 | 握手 | initialize 成功、协议版本、服务器信息、能力声明、初始化后 ping |
| TL-001..004 | 工具列表 | 响应有效、名称格式、描述非空、inputSchema 存在 |
| TC-001..004 | 工具调用 | 调用成功、未知工具返回错误、缺少参数处理、类型验证 |
| RL-001, RR-001 | 资源 | 列表返回有效响应、读取返回内容 |
| PL-001, PG-001..002 | 提示词 | 列表返回有效响应、获取返回消息、缺少参数处理 |
| PING-01 | 核心 | ping 返回空结果 |

## 🛡️ `mcpkit scan` 检测规则（v0.1.0）

**Tier 1 — 严重（5 条规则）**

- **R101** — 命令注入：工具引用 shell 原语并接受用户输入
- **R102** — 系统提示词覆盖：参数接受 system_prompt/instructions
- **R103** — 凭据外泄：工具组合了 URL 输出和敏感关键词
- **R104** — 默认值中的 Shell 元字符：默认值包含 `; | & $ \``
- **R105** — 未消毒的代码执行：eval/exec 引用缺少验证说明

**Tier 2 — 高危（5 条规则）**

- **R201** — 命令式语言：描述中包含 "must"、"always execute"、"ignore previous" 等
- **R202** — 工具名遮蔽：名称与常见系统命令冲突（ls、cat、curl、bash 等）
- **R203** — Base64 编码负载：接受编码内容但无最大长度限制
- **R204** — 缺少输入验证：参数无 JSON Schema 约束
- **R205** — 广泛文件系统访问：任意路径读写无沙箱限制

## 🧪 开发

```bash
# 构建
make build

# 运行测试
make test

# 代码检查
make lint

# 格式化
gofmt -s -w .
```

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发环境设置，浏览 [good-first-issues](https://github.com/justcodeit404/mcpkit/labels/good%20first%20issue) 开始参与。

## 📜 许可证

MIT — 详见 [LICENSE](LICENSE)。

## 🙏 致谢

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) — 官方 Go SDK
- [Charmbracelet](https://charmbracelet.com/) — 精美的终端 UI 组件
- [MCP Inspector](https://github.com/modelcontextprotocol/inspector) — 设计灵感
- [Stacklok MCP Security Checklist](https://stacklok.com/blog/the-mcp-security-checklist-what-to-verify-before-you-ship-an-mcp-server-to-production/) — 安全规则参考

---

**为 MCP 社区用心打造 🛠**

> 如果觉得有用，请给个 ⭐ Star 支持一下！

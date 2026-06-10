# wechat-AI

> 将个人微信化身 AI 机器人，支持多模型接入，基于 [openwechat](https://github.com/eatmoreapple/openwechat) 开发。
> 本仓库 fork 自 [869413421/wechatbot](https://github.com/869413421/wechatbot)。

## 功能特性

- **多模型接入** — 支持切换不同的 AI 供应商（Moonshot / OpenAI / Claude / 文心一言 / 通义千问）
- **私聊自动回复** — 对白名单内好友的消息进行 AI 自动回复（支持上下文记忆）
- **群聊 @ 回复** — 在群中被 @ 时自动回复，可限定只回复指定群
- **上下文记忆** — 会话超时时间内（默认 60s）自动累积上下文
- **清空会话口令** — 发送配置的令牌即可重置当前对话上下文
- **好友自动通过** — 可选自动同意好友申请
- **回复速率控制** — 可配置回复间隔（如 `1s`），避免触发风控
- **多方式部署** — 二进制直接运行 / Docker / Supervisor 进程守护

## 前置条件

- Go 1.16+
- AI 供应商的 API Key
- 微信实名认证账号

## 支持的 AI 供应商

| 供应商 | provider 值 | 默认模型 | 认证方式 | 备注 |
|--------|-------------|----------|----------|------|
| **Moonshot** (月之暗面) | `moonshot` | `moonshot-v1-8k` | `api_key` | 默认供应商 |
| **OpenAI** | `openai` | `gpt-3.5-turbo` | `api_key` | 兼容 DeepSeek / Yi 等 |
| **Anthropic Claude** | `claude` | `claude-3-5-sonnet-20241022` | `api_key` | 需配置 `x-api-key` |
| **Baidu 文心一言** | `baidu` | `ernie-4.0-8k-latest` | `api_key` + `secret_key` | 千帆大模型平台 |
| **Alibaba 通义千问** | `qwen` | `qwen-turbo` | `api_key` | DashScope API |
| 其他 OpenAI 兼容 API | `openai` | 自定义 | `api_key` + `base_url` | DeepSeek / Yi / Groq 等 |

## 配置说明

### 完整配置项

```json
{
  "provider": "moonshot",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "secret_key": "",
  "base_url": "",
  "auto_pass": true,
  "session_timeout": 60,
  "max_tokens": 256,
  "model": "moonshot-v1-8k",
  "temperature": 0.2,
  "reply_prefix": "",
  "session_clear_token": "清空会话",
  "system_prompt": "你是一个性格温暖、自带幽默感且充满好奇心的AI伙伴。",
  "whitelist_users": [],
  "whitelist_groups": [],
  "reply_interval": "1s"
}
```

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `provider` | string | `moonshot` | AI 供应商：`moonshot` / `openai` / `claude` / `baidu` / `qwen` |
| `api_key` | string | — | API 密钥（必填，Baidu 模式下为 Client ID） |
| `secret_key` | string | — | 附加密钥（仅 Baidu 模式需要，作为 Client Secret） |
| `base_url` | string | — | OpenAI 兼容模式的 API 地址，末尾需带 `/v1/` |
| `auto_pass` | bool | `false` | 是否自动通过好友申请 |
| `session_timeout` | number | `60` | 会话上下文保留时间（秒） |
| `max_tokens` | number | `512` | AI 回复最大 token 数 |
| `model` | string | 各供应商默认 | 模型名称 |
| `temperature` | float | `0.9` | 生成温度 (0~1) |
| `reply_prefix` | string | `""` | 私聊回复前缀 |
| `session_clear_token` | string | `"下一个问题"` | 清空上下文口令 |
| `system_prompt` | string | `""` | 系统提示词，设定 AI 角色/风格 |
| `whitelist_users` | array | `[]` | 私聊白名单昵称（空=不限制） |
| `whitelist_groups` | array | `[]` | 群聊白名单名称（空=不限制） |
| `reply_interval` | string | `"1s"` | 回复间隔，如 `"500ms"` `"2s"` |

### 环境变量

环境变量优先级高于配置文件：

| 变量 | 对应配置 |
|------|----------|
| `PROVIDER` | `provider` |
| `APIKEY` | `api_key` |
| `SECRET_KEY` | `secret_key` |
| `BASE_URL` | `base_url` |
| `AUTO_PASS` | `auto_pass` |
| `SESSION_TIMEOUT` | `session_timeout` |
| `MODEL` | `model` |
| `MAX_TOKENS` | `max_tokens` |
| `TEMPREATURE` | `temperature` |
| `REPLY_PREFIX` | `reply_prefix` |
| `SESSION_CLEAR_TOKEN` | `session_clear_token` |
| `REPLY_INTERVAL` | `reply_interval` |

## 供应商接入指南

### 1. Moonshot（月之暗面）— 默认

```json
{
  "provider": "moonshot",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "model": "moonshot-v1-8k"
}
```

- 注册：[Moonshot AI 开放平台](https://platform.moonshot.cn/)
- 模型参考：`moonshot-v1-8k` / `moonshot-v1-32k` / `moonshot-v1-128k`

### 2. OpenAI

```json
{
  "provider": "openai",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "model": "gpt-4o"
}
```

### 3. DeepSeek（深度求索）

DeepSeek 兼容 OpenAI 接口，使用 `openai` provider 并配置 `base_url`：

```json
{
  "provider": "openai",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "base_url": "https://api.deepseek.com/v1/",
  "model": "deepseek-chat"
}
```

### 4. 零一万物 Yi

```json
{
  "provider": "openai",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "base_url": "https://api.01.ai/v1/",
  "model": "yi-large"
}
```

### 5. Anthropic Claude

```json
{
  "provider": "claude",
  "api_key": "sk-ant-xxxxxxxxxxxxxxxxxxxx",
  "model": "claude-3-5-sonnet-20241022"
}
```

- `base_url` 可选，默认 `https://api.anthropic.com/v1/`
- 支持覆盖 `system_prompt`

### 6. Baidu 文心一言（千帆）

```json
{
  "provider": "baidu",
  "api_key": "你的 Client ID",
  "secret_key": "你的 Client Secret",
  "model": "ernie-4.0-8k-latest"
}
```

- 注册：[百度千帆大模型平台](https://console.bce.baidu.com/qianfan/)
- 自动获取 `access_token`，过期自动刷新
- 可用模型：`ernie-4.0-8k-latest` / `ernie-3.5-8k-latest` / `ernie-speed-8k` 等

### 7. Alibaba 通义千问（DashScope）

```json
{
  "provider": "qwen",
  "api_key": "sk-xxxxxxxxxxxxxxxxxxxx",
  "model": "qwen-turbo"
}
```

- 注册：[阿里云 DashScope](https://dashscope.aliyun.com/)
- 可用模型：`qwen-turbo` / `qwen-plus` / `qwen-max` 等

### 8. 其他 OpenAI 兼容 API

任意兼容 `/v1/chat/completions` 格式的 API 均可使用 `openai` provider：

```json
{
  "provider": "openai",
  "api_key": "你的 API Key",
  "base_url": "https://your-api-endpoint/v1/",
  "model": "your-model-name"
}
```

适用场景：Groq / Together AI / Perplexity / Local LLM (vLLM / Ollama) 等。

## 快速开始

```bash
cp config.dev.json config.json
# 编辑 config.json 填入 api_key 和 provider
go run main.go
```

终端显示二维码后，使用微信扫码登录。

后台运行：
```bash
go build -o wechatbot ./main.go
nohup ./wechatbot &> run.log &
tail -f run.log
```

## Docker 部署

### 环境变量方式

```bash
docker run -itd --name wechatbot --restart=always \
  -e PROVIDER=moonshot \
  -e APIKEY=sk-xxxxxxxxxxxxxxxxxxxx \
  -e MODEL=moonshot-v1-8k \
  -e AUTO_PASS=true \
  -e SESSION_TIMEOUT=60s \
  -e MAX_TOKENS=256 \
  -e TEMPREATURE=0.2 \
  -e SYSTEM_PROMPT=你的prompt \
  -e SESSION_CLEAR_TOKEN=清空会话 \
  docker.mirrors.sjtug.sjtu.edu.cn/qingshui869413421/wechatbot:latest
```

### 配置文件挂载

```bash
docker run -itd --name wechatbot \
  -v `pwd`/config.json:/app/config.json \
  docker.mirrors.sjtug.sjtu.edu.cn/qingshui869413421/wechatbot:latest
```

查看二维码：
```bash
docker exec -it wechatbot bash
tail -f -n 50 /app/run.log
```

## 项目结构

```
├── main.go                     # 入口
├── bootstrap/bootstrap.go      # 启动流程
├── config/config.go            # 配置加载
├── pkg/
│   └── llm/
│       └── llm.go              # Provider 接口 + ChatMessage 类型
├── ai/
│   ├── ai.go                   # 供应商注册表 + Completions()
│   ├── gpt35.go                # 旧版 GPT-3.5（未使用）
│   ├── moonshot/
│   │   └── provider.go         # 月之暗面 Moonshot
│   ├── openai/
│   │   └── provider.go         # OpenAI / DeepSeek / Yi 等兼容 API
│   ├── claude/
│   │   └── provider.go         # Anthropic Claude
│   ├── baidu/
│   │   └── provider.go         # 百度文心一言
│   └── qwen/
│       └── provider.go         # 阿里通义千问
├── handlers/
│   ├── handler.go              # 消息分发路由
│   ├── user_msg_handler.go     # 私聊处理
│   ├── group_msg_handler.go    # 群聊 @ 处理
│   └── token_msg_handler.go    # 清空口令处理
├── service/user.go             # 用户会话上下文服务
├── pkg/logger/logger.go        # 日志工具
├── Dockerfile
├── Makefile
├── supervisord.conf
└── storage.json                # 微信登录缓存（自动生成）
```

## 扩展新供应商

新增供应商只需三步：

1. 在 `ai/` 下创建文件，实现 `Provider` 接口：

```go
type MyProvider struct{}

func (p *MyProvider) Name() string { return "myprovider" }

func (p *MyProvider) Chat(messages []ChatMessage) (string, error) {
    // 调用 API 并返回回复文本
}
```

2. 在 `ai/ai.go` 的 `providers` map 中注册：

```go
var providers = map[string]Provider{
    "moonshot": &MoonshotProvider{},
    // ...
    "myprovider": &MyProvider{},
}
```

3. 配置中使用 `"provider": "myprovider"` 即可，无需改动 handlers。

## 常见问题

- **`login error: write storage.json: bad file descriptor`** — 删除 `storage.json` 重新登录
- **`301 response missing Location header`** — 检查 PC 微信能否正常登录
- **二维码无法扫描** — 缩小终端窗口尺寸，让二维码像素更清晰
- **机器人答非所问** — 发送 `session_clear_token` 配置的口令清空上下文

## 注意事项

- 仅供个人学习娱乐使用，滥用可能导致微信账号被限制
- 本项目不做敏感信息过滤，请自行注意收发内容
- 首次登录需扫码，`storage.json` 会保存会话以便后续热登录

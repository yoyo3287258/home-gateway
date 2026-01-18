# Home Smart Control API Gateway

基于 Go 语言开发的智能家居控制网关，专为 Rock 5B (ARM64) 等边缘设备设计。通过集成 LLM（大语言模型）实现智能意图识别和参数提取，支持多渠道（HTTP, Telegram）接入和 Kafka 异步处理。

## ✨ 特性

- **智能意图识别**：集成 OpenAI 兼容的 LLM API，自动识别用户指令意图并提取参数。
- **多渠道支持**：目前支持 HTTP API 和 Telegram Bot，易于扩展更多渠道。
- **配置热重载**：支持不重启服务的情况下动态更新处理器配置。
- **安全机制**：
  - API Token 认证
  - IP 白名单 (支持 CIDR)
  - 接口速率限制
  - Telegram Webhook 签名验证
- **异步处理**：基于 Kafka 的请求/响应模型，解耦指令接收与执行。
- **自动更新**：内置 Git Release 自动检查和更新功能。
- **ARM64 优化**：针对边缘设备（如 Rock 5B）优化，支持跨平台编译。

## 🚀 快速开始

### 1. 配置文件

在 `configs/` 目录下创建 `config.yaml`（参考示例）：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

llm:
  base_url: "https://api.openai.com/v1"
  api_key: "${LLM_API_KEY}"
  model: "gpt-4o-mini"

kafka:
  brokers: ["localhost:9092"]
  request_topic: "home.request"
  response_topic: "home.response"

security:
  api_token: "your-secret-token"
  ip_whitelist: ["192.168.1.0/24"]
```

### 2. 处理器配置

在 `configs/processors/` 目录下添加 YAML 文件定义技能（如 `lighting.yaml`）：

```yaml
processors:
  - id: "light_living_room"
    name: "客厅灯"
    description: "控制客厅的主灯"
    group: "lighting"
    keywords: ["客厅灯", "大灯"]
    parameters:
      - name: "action"
        type: "enum"
        values: ["on", "off"]
        required: true
    enabled: true
```

### 3. 运行

```bash
# 设置环境变量（推荐）
export LLM_API_KEY="sk-..."

# 启动服务
./home-gateway
```

## 📚 API 文档

### 通用指令接口

`POST /api/v1/command`

**Header:**
- `Authorization: Bearer <your-api-token>`
- `Content-Type: application/json`

**Body:**
```json
{
  "content": "帮我把客厅的灯打开",
  "user_id": "user123" // 可选
}
```

**Response:**
```json
{
  "message": "操作成功",
  "data": { ... },
  "trace_id": "12345..."
}
```

### 配置重载

`POST /api/v1/config/reload`

## 🛠️ 开发与构建

### 本地运行
```bash
go run ./cmd/gateway
```

### 编译
```bash
# Windows
go build -o home-gateway.exe ./cmd/gateway

# Linux ARM64
set GOOS=linux
set GOARCH=arm64
go build -o home-gateway-linux-arm64 ./cmd/gateway
```

## 📦 部署

项目包含 GitHub Actions 工作流，Tag 推送（如 `v1.0.0`）会自动构建多平台二进制文件并发布 Release。

## 📄 License
MIT

# Socratic Tutor Harness (苏格拉底代码学习导师服务)

这是一个低成本、轻量级的无头 Web API 服务，专为对接 Dify/Coze 等大模型平台（LLMOps）自定义工具（Custom Tools）而设计。服务基于 Go 语言原生标准库构建，用于以“苏格拉底式提问”引导学习者进行代码审计、SRE 与 DevOps 习惯纠偏。

## 项目核心设计与 SRE 实践

本项目不仅实现了大模型 Prompt 拼接与会话交互，还在 Go 后端逻辑中融入了多项 SRE 与安全开发规范：

1. **显式 Context 超时与防泄漏传播**：
   - 客户端请求的 Context 超时时间（125 秒）显式传递给内部的会话压缩引擎与外部 LLM API 客户端。
   - 当客户端提前断开连接或触发超时，Go HTTP 客户端会立即中止向第三方大模型的出站请求，彻底防止 Goroutine 泄露与文件描述符（FD）积压。
2. **严谨的服务器超时金字塔（Timeout Hierarchy）**：
   - `ReadTimeout`（10秒）：限制客户端 Header/Body 读取时长，防御 Slowloris 慢速请求攻击。
   - `WriteTimeout`（130秒）：必须大于 Handler Context 超时时间（125秒），确保在 Handler 处理超时（504）时，服务器依然能将优雅的 JSON 错误写入 TCP Socket 并正常关闭连接。
   - `IdleTimeout`（60秒）：长连接空闲自动回收，释放内存资源。
3. **会话压缩机制（Session Compression）**：
   - 使用 SQLite 存储会话历史。
   - 当单次会话的历史报文达到 Payload 阈值（30KB）时，自动将历史数据（除最近 8 条外）提取为 `Session_Summary` 摘要并持久化，供大模型后续交互时作为紧凑上下文，大幅降低 Token 消耗并保持长期记忆。
4. **防御性校验与防路径穿越（SSRF & Traversal Protection）**：
   - 严禁直接使用包含 `../` 的参数作为路径读取本地 Skill，并在构建 System Prompt 时使用 `filepath.Clean` 及前缀白名单机制对路径进行严格物理隔离限制。
   - 对 `SessionID` 和 `Skill` 进行强类型正则字符过滤（仅允许字母、数字、下划线及中划线）。
   - 限制 Request Body 最大 64KB（`http.MaxBytesReader`），防止恶意超大 Payloads 撑爆系统内存。

## 目录结构说明

```text
.
├── api/
│   └── openapi.yaml           # 标准 OpenAPI 3.0 接口定义（Dify 导入用）
├── cmd/
│   └── tutor/
│       └── main.go            # 应用程序入口，负责路由注册、数据库初始化与 HTTP Server 绑定
├── data/
│   └── tutor.db               # 本地 SQLite3 数据库（存储会话历史与压缩摘要）
├── internal/
│   └── tutor/
│       ├── client.go          # 第三方大模型 HTTP API 客户端实现（透传 Context）
│       ├── context.go         # 会话 Payload 估算与 LLM 自动化会话压缩逻辑
│       ├── handler.go         # HTTP 处理器实现（含 /api/v1/socratic/ask 与 /healthz 探针）
│       ├── model.go           # 核心请求与响应数据模型结构体定义
│       ├── prompt.go          # 导师 System Prompt、Memory 快照及 Skill 包物理装配逻辑
│       └── storage.go         # SQLite3 数据库建表、持久化读写与 slog 文本日志追加
└── prompts/
    ├── memory.md              # 记录学习者技术栈、痛点与学习目标进度的快照
    ├── system.md              # 核心苏格拉底系统 Prompt
    └── skills/                # 各种专业技能包的 Markdown 定义
        ├── code_learn.md      # 代码学习审计基础技能包
        └── go_learn.md        # Go 标准库与安全防御进阶技能包
```

## 接口定义

### 1. 核心问答接口
* **路径**: `POST /api/v1/socratic/ask`
* **请求体 (JSON)**:
  ```json
  {
    "session_id": "session_go_001",
    "question": "在 Go 语言中如何避免 TCP 连接泄露？",
    "skill": "go_learn"
  }
  ```
* **响应体 (200 OK)**:
  ```json
  {
    "status": "ok",
    "session_id": "session_go_001",
    "answer": "苏格拉底引导式回答..."
  }
  ```
* **响应体 (400 Bad Request)**:
  ```json
  {
    "status": "error",
    "error_code": "invalid_request",
    "message": "错误细节描述..."
  }
  ```

### 2. 存活与就绪探针（K8s 自愈探针）
* **路径**: `GET /healthz`
* **响应体 (200 OK)**:
  ```json
  {
    "status": "ok"
  }
  ```

## 快速开始

### 1. 导入环境变量
启动服务前，需要配置第三方大模型接口密钥和请求参数：

```bash
export Secret_Token_Key="your_openai_or_deepseek_api_key"
export Secret_Token_Url="https://api.example.com/v1/chat/completions"
export Secret_Token_Model="gemini-3.5-flash"
```

### 2. 编译并运行服务
```bash
go build -o socratic-tutor-service cmd/tutor/main.go
./socratic-tutor-service -addr 0.0.0.0:8083
```

### 3. Dify 接入配置
1. 在 Dify 中选择 **"工具" -> "自定义工具"**。
2. 复制项目下的 `api/openapi.yaml` 文件的 YAML 内容。
3. 修改配置中的 `servers.url` 为你服务运行的真实局域网/公网 IP 地址。
4. 如果部署在内网，需在 Dify 后台配置环境变量 `SSRF_ALLOW_PRIVATE_IP=true` 以放行内网 IP。

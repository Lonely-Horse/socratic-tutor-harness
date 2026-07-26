# 个人技术资产与已完成项目档案

## 一、 已掌握技术栈 (Tech Stack)

### 1. 核心编程语言与底功
* **C 语言**：标准库（stdlib）、系统级 I/O、内存管理（`malloc`/`free`）、网络 Socket 基础。
* **Go 语言**：精通原生标准库开发，包含：
  * `net/http`（Web 服务与 Client 连接池管理）
  * `html/template`（服务端 HTML 渲染）
  * `encoding/json`（流式 JSON 解析与序列化）
  * `flag`（命令行参数解析）
  * `time`（超时控制与时间处理）
  * `sync`（`sync.Mutex` 并发安全控制）
* **Python**：掌握基础语法，熟悉 **Flask** 和 **FastAPI** 框架，能够快速搭建轻量级 Web 服务或辅助脚本。
* **Shell 脚本**：掌握基础语法（变量、判断、循环、管道与重定向），能够编写和修改基础运维自动化脚本。

### 2. 云原生与运维基础设施
* **容器化**：Docker、Docker Compose 编排。
* **自动化运维**：Ansible（Playbook 剧本编写、`base` 初始化、`deploy` 部署、`nginx` 配置管理）。
* **网络与安全边界**：Nginx 反向代理、路径级访问控制（ACL）、Tailscale 虚拟内网集成。

---

## 二、 已完成项目 (Completed Projects)

### 1. `Socratic Tutor` 智能体 Skill 接口
* **完成时间**：2026年7月。
* **技术实现**：基于 **Go 原生标准库** 实现。
* **项目描述**：编写高性能的 Web API 服务，作为大模型（Dify 等 AI 平台）的自定义 Skill。以苏格拉底式的引导逻辑为核心，接收 JSON 请求，处理并发调度，并安全输出结构化数据。

### 2. `exam-prep` 考研打卡与知识检索助手
* **完成时间**：2026年7月。
* **技术实现**：
  * 基于纯 Go 原生标准库实现。
  * `sync.Mutex` 保证高并发下内存与本地 JSON 文件的读写安全。
  * `json.NewDecoder` 流式解析，防御大文件内存过载。
* **双轨制架构**：
  * 人类轨道：`/dashboard`（使用 `html/template` 进行 HTML 服务端页面渲染）。
  * AI 轨道：`/api/skill`（干净的 JSON API，遵循 OpenAPI 规范）。
* **AI 平台对接**：完成 Dify 平台集成，实现 AI 调用 Go 接口的 Skill 完整链路。

### 3. FastAPI 个人博客（边缘业务节点）
* **架构**：Docker Compose 容器化运行。
* **网络安全防护**：公网入口部署 Nginx 反向代理，配置路径级访问控制规则（公网访问 `/metrics`、`/docs`、`/redoc` 返回 403，仅允许内网 Tailscale 抓取与访问）。

### 4. LLM Wiki（已归档）
* **状态**：早期探索痕迹，目前已归档不维护。
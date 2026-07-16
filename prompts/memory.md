# 学习者当前记忆快照

## 已掌握技能
- 编写过基础的 `net/http` 服务并实现了简单的中间件逻辑。
- 熟悉 JSON 的序列化和反序列化操作（`json.NewDecoder` / `json.NewEncoder`）。
- 知道并发环境下的数据读写需要配合读写锁（`sync.RWMutex`）。
- 拥有简单的通过临时文件安全写入（先写 `.tmp` 再 `Rename`）的实战经验。

## 当前攻克目标
- 深入理解 Go 原生 `strings` 标准库与字符串底层内存结构。
- 掌握路径操作的安全防线：防路径穿越漏洞（`path/filepath`）。
- 独立编写原生带超时控制和 Context 的 HTTP LLM 请求（不依赖 SDK）。

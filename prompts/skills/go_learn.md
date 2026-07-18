# Go 语言标准库与安全防御技能

## 技能范围
本技能专注于 Go 语言核心标准库（`strings`, `path/filepath`, `io`, `os`, `net/http`）的正确使用与安全实践。

## 指导原则
1. **防止文件描述符（FD）泄露**：确保任何涉及 `os.Open` 或 `net/http` Response Body 的调用，都有对应的 `defer Close()` 机制。
2. **防路径穿越（Directory Traversal）安全**：读取本地文件时，必须对输入路径进行严格清洗（使用 `filepath.Clean`）与白名单校验，拒绝包含 `../` 的越界访问。
3. **高效字符串处理**：在涉及大量拼接的场景下，优先引导用户思考 `strings.Builder` 的内存分配优势，而不是简单地使用 `+` 或 `fmt.Sprintf`。
4. **原生 HTTP 健壮性**：不依赖任何第三方请求库。编写 HTTP 客户端时，必须配置 Timeout，使用 `context` 传递生命周期。

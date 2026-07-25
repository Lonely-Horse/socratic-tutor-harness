package tutor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func estimatePayload(systemPrompt string, contextMsgs []Message, currentQuestion string) int {
	n := len(systemPrompt) + len(currentQuestion)
	for _, m := range contextMsgs {
		n = n + len(m.Role) + len(m.Content)
	}
	return n
}

func splitForCompress(all []StoredMessage) (toCompress []StoredMessage, recent []StoredMessage) {
	if len(all) <= RecentRawLimit {
		return nil, all
	}
	toCompress = all[:len(all)-RecentRawLimit]
	recent = all[len(all)-RecentRawLimit:]
	return toCompress, recent
}

func storedToMessages(in []StoredMessage) []Message {
	var message []Message
	for _, storedmessage := range in {
		message = append(message, Message{
			Role:    storedmessage.Role,
			Content: storedmessage.Content,
		})
	}
	return message
}

func buildCompressMaterial(message []StoredMessage) string {
	var b strings.Builder

	b.WriteString("以下是需要压缩的旧对话原文。请只根据这些内容生成摘要，不要发明事实，不要捏造虚假信息。\n\n")
	for _, msg := range message {
		b.WriteString("[")
		b.WriteString(msg.Role)
		b.WriteString("]")
		b.WriteString(msg.Content)
		b.WriteString("\n\n")
	}
	return b.String()
}

func CompressSession(ctx context.Context, db *sql.DB, sessionID string, toCompress []StoredMessage) error {
	const compressSystemPrompt = `
	你是一个会话压缩器，不是回答问题的导师。

	你的任务：
	把一段旧的学习对话压缩成 session_summary，供后续 LLM 在同一 session 中恢复上下文。

	压缩原则：
	1. 用户内容优先：
	- 保留用户的根问题、目标、约束、纠正、否定、前后衔接问题。
	- 尽量贴近用户原意，不要改写成用户没有表达过的结论。
	2. Assistant 内容大幅压缩：
	- 只保留关键误区、最终澄清点、未完成的思考题或下一步任务。
	- 删除重复解释、寒暄、长篇示例、过程性铺垫。
	3. 禁止发明：
	- 不要声称“用户已经掌握某技术”，除非原文中有明确证据。
	- 如果用户口头说会，但后续使用仍出错，应记录为“仍需巩固”，不要记为已掌握。
	4. 冲突优先级：
	- 近端原文 > 本摘要 > memory 全局记忆。
	- 摘要中如涉及不确定结论，要写成“可能/待确认”，不要写死。
	5. 输出格式：
	- 只输出摘要正文，不要解释你如何压缩。
	- 不要输出 JSON。
	- 不要输出 Markdown 表格。
	- 控制在约 800-1200 中文字以内。

	摘要应包含：
	- 本 session 的主题/主线
	- 用户的根问题和重要纠正
	- 已澄清的关键点
	- 仍未完成或需要继续追问的问题
	- 对 assistant 侧长回答的极简概括
	`
	if sessionID == "" {
		return fmt.Errorf("The sessionID is empty")
	}
	if len(toCompress) == 0 {
		return nil
	}

	material := buildCompressMaterial(toCompress)
	summaryText, err := AskLLM(ctx, compressSystemPrompt, material, nil)
	if err != nil {
		return err
	}
	if summaryText == "" {
		return fmt.Errorf("The summaryText is empty")
	}
	untilID := toCompress[len(toCompress)-1].ID

	err = SaveSummary(db, sessionID, summaryText, untilID)
	if err != nil {
		return err
	}

	return nil
}

func BuildModelContext(ctx context.Context, db *sql.DB, sessionID, systemPrompt, currentQuestion string) ([]Message, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("The sessionID is empty")
	}

	all, err := LoadMessagesWithID(db, sessionID)
	if err != nil {
		return nil, err
	}

	sum, err := LoadSummary(db, sessionID)
	if err != nil {
		return nil, err
	}

	var tail []StoredMessage
	if sum.SummaryText == "" {
		tail = all
	} else {
		for _, msg := range all {
			if sum.UntilID < msg.ID {
				tail = append(tail, msg)
			}
		}
	}

	var contextMsgs []Message
	if sum.SummaryText != "" {
		contextMsgs = append(contextMsgs, Message{
			Role:    "user",
			Content: "Session_Summary\n" + sum.SummaryText,
		})
	}

	contextMsgs = append(contextMsgs, storedToMessages(tail)...)

	size := estimatePayload(systemPrompt, contextMsgs, currentQuestion)
	if size < PayloadLimitBytes {
		return contextMsgs, nil
	}

	toCompress, recent := splitForCompress(all)
	if len(toCompress) == 0 {
		return contextMsgs, nil
	}

	err = CompressSession(ctx, db, sessionID, toCompress)
	if err != nil {
		AppendEventLog(
			"data/tutor.log",
			"context_compress",
			"session", sessionID,
			"estimated_size", size,
			"history_count", len(all),
			"to_comporess_count", len(toCompress),
			"recent_count", len(recent),
			"ok", false,
			"err", "compress_failed",
		)
		return contextMsgs, nil
	}

	newSum, err := LoadSummary(db, sessionID)
	if err != nil {
		return contextMsgs, nil
	}

	var rebuilt []Message
	if newSum.SummaryText != "" {
		rebuilt = append(rebuilt, Message{
			Role:    "user",
			Content: "Session_Summary\n" + newSum.SummaryText,
		})
	}
	rebuilt = append(rebuilt, storedToMessages(recent)...)

	err = AppendEventLog(
		"data/tutor.log",
		"context_compress",
		"session", sessionID,
		"estimated_size", size,
		"history_count", len(all),
		"to_compress_count", len(toCompress),
		"recent_count", len(recent),
		"ok", true,
	)
	if err != nil {
		return rebuilt, err
	}

	return rebuilt, nil
}

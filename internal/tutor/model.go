package tutor

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Payload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ApiResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type SessionSummary struct {
	SessionID   string
	SummaryText string
	UntilID     int64
}

type StoredMessage struct {
	ID      int64
	Role    string
	Content string
}

type AskRequest struct {
	SessionID string `json:"session_id"`
	Question  string `json:"question"`
	Skill     string `json:"skill"`
}

type AskResponse struct {
	Status    string `json:"status"`
	SessionID string `json:"session_id,omitempty"`
	Answer    string `json:"answer,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

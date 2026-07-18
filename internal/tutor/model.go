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

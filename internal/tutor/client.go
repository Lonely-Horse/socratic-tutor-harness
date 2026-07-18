package tutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func AskLLM(SystemPrompt, UserQuestion string, Data []Message) (string, error) {
	apiKey := os.Getenv("Secret_Token_Key")
	apiUrl := os.Getenv("Secret_Token_Url")
	if apiKey == "" || apiUrl == "" {
		return "", fmt.Errorf("The Secret_Token_Key or Secret_Token_Url is not find in enviroment")
	}

	var tmpData []Message
	tmpData = append(tmpData, Message{Role: "system", Content: SystemPrompt})
	tmpData = append(tmpData, Data...)
	tmpData = append(tmpData, Message{Role: "user", Content: UserQuestion})

	Pay1 := Payload{
		Model:    "grok-4.5",
		Messages: tmpData,
	}

	Pay2, err := json.Marshal(Pay1)
	if err != nil {
		return "", fmt.Errorf("The struct isn't Marshal,detail: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client := &http.Client{}
	var apiResp ApiResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiUrl, bytes.NewBuffer(Pay2))
	if err != nil {
		return "", fmt.Errorf("The request failed,detail: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("The request isn't do,detail: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&apiResp)
	if err != nil {
		return "", fmt.Errorf("The decoder isn't decoder,detail: %s", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("empty choices returned form API")
	}

	return apiResp.Choices[0].Message.Content, nil
}

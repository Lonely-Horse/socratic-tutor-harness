package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func buildSystemPrompt() (prompt string, err error) {
	var file strings.Builder
	SystemPath := filepath.Clean(filepath.Join("prompts/", "system.md"))
	SkillPath := filepath.Clean(filepath.Join("prompts/", "skill.md"))
	MemoryPath := filepath.Clean(filepath.Join("prompts/", "memory.md"))

	Systemfile, err := os.ReadFile(SystemPath)
	if err != nil {
		return "", err
	}

	Skillfile, err := os.ReadFile(SkillPath)
	if err != nil {
		return "", err
	}

	Memoryfile, err := os.ReadFile(MemoryPath)
	if err != nil {
		return "", err
	}

	file.Write(Systemfile)
	file.WriteString("\n")
	file.Write(Skillfile)
	file.WriteString("\n")
	file.Write(Memoryfile)

	return file.String(), err
}

func askLLM(SystemPrompt, UserQuestion string) (string, error) {
	apiKey := os.Getenv("Secret_Token_Key")
	apiUrl := os.Getenv("Secret_Token_Url")
	if apiKey == "" || apiUrl == "" {
		return "", fmt.Errorf("The Secret_Token_Key or Secret_Token_Url is not find in enviroment\n")
	}
	Pay1 := Payload{
		Model: "gemini-3.5-flash",
		Messages: []Message{
			{
				Role:    "system",
				Content: SystemPrompt,
			},
			{
				Role:    "user",
				Content: UserQuestion,
			},
		},
	}

	Pay2, err := json.Marshal(Pay1)
	if err != nil {
		return "", fmt.Errorf("The struct isn't Marshal,detail: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
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

func appendHistory(UserQuestion, answer, filepath string) error {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	var record strings.Builder
	record.WriteString("\n## 提问时间:" + currentTime + "\n")
	record.WriteString("### User:\n" + UserQuestion + "\n\n")
	record.WriteString("### Assistant:\n" + answer + "\n")
	record.WriteString("---\n")

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(record.String())
	return err
}

func main() {
	var question string
	historypath := filepath.Clean(filepath.Join("prompts", "history.md"))
	flag.StringVar(&question, "question", "你好", "问题")
	flag.Parse()

	prompt, err := buildSystemPrompt()
	if err != nil {
		fmt.Printf("The question have some error, detail: %s", err)
		return
	}

	Answer, err := askLLM(prompt, question)
	if err != nil {
		fmt.Printf("The askLLM have some problem,detail:%s", err)
		return
	}

	err = appendHistory(question, Answer, historypath)
	if err != nil {
		fmt.Printf("The appendHistory have some problem,detail: %s", err)
		return
	}

	fmt.Printf("The LLM's answer:\n%s\n", Answer)
}

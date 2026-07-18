package main

import (
	"flag"
	"fmt"
	"socratic-tutor-harness/internal/tutor"
)

func main() {
	var question string
	var skill string
	var session string
	flag.StringVar(&question, "question", "你好", "问题")
	flag.StringVar(&skill, "skill", "init", "技能")
	flag.StringVar(&session, "session", "default", "话题")
	flag.Parse()

	DataBasePath := "prompts/data.db"
	db, err := tutor.BuildDatabase(DataBasePath)
	if err != nil {
		fmt.Printf("The question have some error\ndetail: %s", err)
		return
	}
	defer db.Close()

	prompt, err := tutor.BuildSystemPrompt(skill)
	if err != nil {
		fmt.Printf("The question have some error,\ndetail: %s\n", err)
		return
	}

	data, err := tutor.LoadMessages(db, session)
	if err != nil {
		fmt.Printf("The LoadMessage have some problem,detail: %s", err)
		return
	}

	answer, err := tutor.AskLLM(prompt, question, data)
	if err != nil {
		fmt.Printf("The askLLM have some problem\ndetail:%s\n", err)
		return
	}

	err = tutor.SaveMessage(db, session, "user", question)
	if err != nil {
		fmt.Printf("The SaveMessage to user model have some problem,detail: %s\n", err)
		return
	}
	err = tutor.SaveMessage(db, session, "assistant", answer)
	if err != nil {
		fmt.Printf("The SaveMessage to assisant model have some problem,detail: %s\n", err)
		return
	}

	fmt.Printf("The LLM's answer:\n%s\n", answer)
}

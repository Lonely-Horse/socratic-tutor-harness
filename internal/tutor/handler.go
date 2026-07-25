package tutor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func IsSafeName(s string) bool {
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch <= '9' && ch >= '0') || (ch == '_' || ch == '-')) {
			return false
		}
	}
	return true
}

func ValidateAskRequest(req AskRequest) error {
	if req.SessionID == "" {
		return fmt.Errorf("The sessionid is empty,please enter the sessionid")
	}
	if req.Question == "" {
		return fmt.Errorf("The questionis empty,please enter question")
	}
	if !IsSafeName(req.SessionID) {
		return fmt.Errorf("The sessionid have illegal chars")
	}
	if len(req.SessionID) > 256 {
		return fmt.Errorf("The sessionid is too long")
	}
	if len(req.Question) > 8192 {
		return fmt.Errorf("The question is too long")
	}
	if req.Skill != "" {
		if !IsSafeName(req.Skill) {
			return fmt.Errorf("The skill have illegal chars")
		}
	}
	return nil
}

type SocraticServer struct {
	DB *sql.DB
}

func (s *SocraticServer) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error":"The method not allowed"}`))
}

func (s *SocraticServer) HandleAsk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"The method not allowed"}`))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

	var req AskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			w.Write([]byte(`{"status":"error","err_code":"payload_too_large","message":"request body limit is 64kb"}`))
			return
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"error","error_code":"invalid_request","message":"invalid json format"}`))
		return
	}

	err = ValidateAskRequest(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"status":"error","error_code":"invalid_request","error":"%s"}`, err.Error())))
		return
	}

	if req.Skill == "" {
		req.Skill = "code_learn"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 125*time.Second)
	defer cancel()

	systemPrompt, err := BuildSystemPrompt(req.Skill)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","error_code":"internal_error","message":"failed to build system prompt"}`))
		return
	}

	data, err := BuildModelContext(ctx, s.DB, req.SessionID, systemPrompt, req.Question)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status":"error","error_code":"internal_error","message":"failed to build the context"}`))
		return
	}

	answer, err := AskLLM(ctx, systemPrompt, req.Question, data)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			w.WriteHeader(http.StatusGatewayTimeout)
			w.Write([]byte(`{"status":"error","error_code":"llm_timeout","message":"upstream model time out"}`))
			return
		}
		fmt.Printf("The AskLLM error:\n%s", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"status":"error","error_code":"upstream_error","message":"failed to request upstream model"}`))
		return
	}

	err = SaveMessage(s.DB, req.SessionID, "user", req.Question)
	if err != nil {
		_ = AppendEventLog("data/tutor.log", "save_message_failed", "user", "err", err.Error())
	}

	err = SaveMessage(s.DB, req.SessionID, "assistant", answer)
	if err != nil {
		_ = AppendEventLog("data/tutor.log", "save_message_failed", "role", "assistant", "err", err.Error())
	}

	resp := AskResponse{
		Status:    "ok",
		SessionID: req.SessionID,
		Answer:    answer,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

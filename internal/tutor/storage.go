package tutor

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/glebarez/go-sqlite"
)

const (
	PayloadLimitBytes = 30 * 1024
	RecentRawLimit    = 8
)

func BuildDatabase(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("The dbPath is empty")
	}

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("The database isn't open,detail: %s", err)
	}

	err = database.Ping()
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("The database didn't ping.detail: %s", err)
	}

	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP 
	);`)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("The database open/created failed,detail: %s", err)
	}
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS session_summary(
		session_id TEXT PRIMARY KEY,
		summary_text TEXT NOT NULL,
		until_id INTEGER NOT NULL,
		update_at DATETIME DEFAULT CURRENT_TIMESTAMP 
	);`)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("The database open/created failed,detail: %s", err)
	}

	return database, nil
}

func SaveMessage(db *sql.DB, sessionID, role, content string) error {
	if sessionID == "" || content == "" {
		return fmt.Errorf("The Sessionid or Content is empty")
	}

	var roles = [2]string{"user", "assistant"}
	var flag int
	for _, Role := range roles {
		if Role == role {
			flag = 1
			break
		}
	}
	if flag != 1 {
		return fmt.Errorf("The Role didn't exist,Please enter exist Role")
	}

	_, err := db.Exec(`INSERT INTO history (session_id,role,content) VALUES (?,?,?)`, sessionID, role, content)
	if err != nil {
		return fmt.Errorf("The err is %s", err)
	}

	return nil
}

func LoadMessages(db *sql.DB, sessionID string) ([]Message, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("The sessionId is empty")
	}

	rows, err := db.Query(`SELECT role,content FROM history WHERE session_id = ? ORDER BY id ASC`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("The Query have some problem,detail: %s", err)
	}
	defer rows.Close()

	var message []Message
	var role, content string
	for rows.Next() {
		err := rows.Scan(&role, &content)
		if err != nil {
			return nil, fmt.Errorf("The rows didn't Scan,detail: %s", err)
		}

		message = append(message, Message{
			Role:    role,
			Content: content,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("The rows have some error,detail: %s", err)
	}

	return message, nil
}

func LoadMessagesWithID(db *sql.DB, sessionID string) ([]StoredMessage, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("The sessionID is rmpty")
	}

	rows, err := db.Query(`SELECT id,role,content FROM history WHERE session_id = ? ORDER BY id ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var id int64
	var role, content string
	var list []StoredMessage
	for rows.Next() {
		err := rows.Scan(&id, &role, &content)
		if err != nil {
			return nil, err
		}
		list = append(list, StoredMessage{
			ID:      id,
			Role:    role,
			Content: content,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return list, nil
}

func SaveSummary(db *sql.DB, sessionID, summaryText string, untilID int64) error {
	if sessionID == "" || summaryText == "" {
		return fmt.Errorf("The sessionID or summaryText is empty")
	}

	_, err := db.Exec(`INSERT INTO session_summary (session_id,summary_text,until_id) 
		VALUES (?,?,?)
		ON CONFLICT(session_id) DO UPDATE SET 
		  summary_text = excluded.summary_text,
		  until_id = excluded.until_id,
		  update_at = CURRENT_TIMESTAMP`,
		sessionID, summaryText, untilID)
	if err != nil {
		return err
	}
	return nil
}

func LoadSummary(db *sql.DB, sessionID string) (SessionSummary, error) {
	if sessionID == "" {
		return SessionSummary{}, fmt.Errorf("The sessionID is empty")
	}

	row := db.QueryRow(`SELECT summary_text,until_id FROM session_summary WHERE session_id = ?`, sessionID)

	var summary_text string
	var until_id int64
	err := row.Scan(&summary_text, &until_id)
	if err == sql.ErrNoRows {
		return SessionSummary{}, nil
	}
	if err != nil {
		return SessionSummary{}, err
	}

	sessionsummary := SessionSummary{
		SessionID:   sessionID,
		SummaryText: summary_text,
		UntilID:     until_id,
	}

	return sessionsummary, nil
}

func AppendEventLog(logPath string, event string, attrs ...any) error {
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	handler := slog.NewTextHandler(file, nil)
	logger := slog.New(handler)

	logger.Info(event, attrs...)

	return nil
}

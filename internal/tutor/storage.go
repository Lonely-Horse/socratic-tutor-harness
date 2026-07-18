package tutor

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
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

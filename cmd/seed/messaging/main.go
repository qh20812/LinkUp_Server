package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type userRow struct {
	ID    int
	Email string
}

type chatRow struct {
	ID   int
	Type string
}

type participantSeed struct {
	ChatID int
	UserID int
	Role   string
}

type messageSeed struct {
	ChatID   int
	SenderID int
	Content  string
	MediaID  sql.NullInt64
	EmojiID  sql.NullInt64
}

type callSeed struct {
	ChatID   sql.NullInt64
	CallerID int
	CallType string
	IsGroup  bool
	Status   string
}

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	conn, err := db.ConnectDb(config.GetEnv())
	if err != nil {
		log.Fatalf("DB connection: failed (%v)", err)
	}
	defer conn.Close()

	if err := ensurePhase5Tables(conn); err != nil {
		log.Fatalf("ensure phase5 tables failed: %v", err)
	}

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) < 5 {
		log.Fatalf("need at least 5 users for phase5 seeding, found %d", len(users))
	}

	chats, err := fetchChats(conn)
	if err != nil {
		log.Fatalf("fetch chats failed: %v", err)
	}
	if len(chats) == 0 {
		log.Fatalf("no chats found for phase5 seeding")
	}

	participants := buildParticipants(users, chats)
	participantsInserted, err := seedChatParticipants(conn, participants)
	if err != nil {
		log.Fatalf("seed chat participants failed: %v", err)
	}

	messages := buildMessages(users, chats)
	messagesInserted, err := seedMessages(conn, messages)
	if err != nil {
		log.Fatalf("seed messages failed: %v", err)
	}

	calls := buildCalls(users, chats)
	callsInserted, err := seedCalls(conn, calls)
	if err != nil {
		log.Fatalf("seed calls failed: %v", err)
	}

	fmt.Printf("Seed phase5: success (chat_participants=%d, messages=%d, calls=%d)\n",
		participantsInserted,
		messagesInserted,
		callsInserted,
	)
}

func ensurePhase5Tables(conn *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chat_participants (
			id INT AUTO_INCREMENT PRIMARY KEY,
			chat_id INT NOT NULL,
			user_id INT NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY chat_participant_unique (chat_id, user_id),
			CONSTRAINT fk_chat_participants_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			CONSTRAINT fk_chat_participants_user FOREIGN KEY (user_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INT AUTO_INCREMENT PRIMARY KEY,
			chat_id INT NOT NULL,
			sender_id INT NOT NULL,
			content VARCHAR(2000),
			media_id INT NULL,
			emoji_id INT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_messages_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users(id),
			CONSTRAINT fk_messages_media FOREIGN KEY (media_id) REFERENCES media(id),
			CONSTRAINT fk_messages_emoji FOREIGN KEY (emoji_id) REFERENCES emojis(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS calls (
			id INT AUTO_INCREMENT PRIMARY KEY,
			chat_id INT NULL,
			caller_id INT NOT NULL,
			call_type VARCHAR(20) NOT NULL DEFAULT 'voice',
			is_group BOOLEAN NOT NULL DEFAULT FALSE,
			status VARCHAR(20) NOT NULL DEFAULT 'completed',
			started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			ended_at TIMESTAMP NULL,
			CONSTRAINT fk_calls_chat FOREIGN KEY (chat_id) REFERENCES chats(id),
			CONSTRAINT fk_calls_caller FOREIGN KEY (caller_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, q := range queries {
		if _, err := conn.Exec(q); err != nil {
			return fmt.Errorf("create phase5 table: %w", err)
		}
	}

	return nil
}

func fetchUsers(conn *sql.DB) ([]userRow, error) {
	rows, err := conn.Query("SELECT id, email FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func fetchChats(conn *sql.DB) ([]chatRow, error) {
	rows, err := conn.Query("SELECT id, type FROM chats ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query chats: %w", err)
	}
	defer rows.Close()

	chats := make([]chatRow, 0)
	for rows.Next() {
		var c chatRow
		if err := rows.Scan(&c.ID, &c.Type); err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}
		chats = append(chats, c)
	}

	return chats, rows.Err()
}

func buildParticipants(users []userRow, chats []chatRow) []participantSeed {
	items := make([]participantSeed, 0)
	for i, chat := range chats {
		if strings.EqualFold(chat.Type, "group") {
			for j := 0; j < 5; j++ {
				items = append(items, participantSeed{ChatID: chat.ID, UserID: users[(i+j)%len(users)].ID, Role: "member"})
			}
			items = append(items, participantSeed{ChatID: chat.ID, UserID: users[(i+5)%len(users)].ID, Role: "admin"})
		} else {
			first := users[i%len(users)].ID
			second := users[(i+1)%len(users)].ID
			if first == second {
				second = users[(i+2)%len(users)].ID
			}
			items = append(items, participantSeed{ChatID: chat.ID, UserID: first, Role: "member"})
			items = append(items, participantSeed{ChatID: chat.ID, UserID: second, Role: "member"})
		}
	}
	return items
}

func buildMessages(users []userRow, chats []chatRow) []messageSeed {
	texts := []string{
		"Xin chào, bạn có rảnh không?",
		"Hôm nay mình check lại thông tin seed data.",
		"Bài đăng mới đã được tạo thành công.",
		"Đừng quên review tài liệu kỹ thuật.",
		"Ảnh và video seed đã sẵn sàng.",
	}
	items := make([]messageSeed, 0)
	for i, chat := range chats {
		for j := 0; j < 4; j++ {
			sender := users[(i+j)%len(users)]
			items = append(items, messageSeed{
				ChatID:   chat.ID,
				SenderID: sender.ID,
				Content:  texts[(i+j)%len(texts)],
				MediaID:  sql.NullInt64{Valid: false},
				EmojiID:  sql.NullInt64{Valid: false},
			})
		}
	}
	return items
}

func buildCalls(users []userRow, chats []chatRow) []callSeed {
	statuses := []string{"completed", "missed", "declined"}
	items := make([]callSeed, 0)
	for i, chat := range chats {
		caller := users[i%len(users)]
		items = append(items, callSeed{
			ChatID:   sql.NullInt64{Int64: int64(chat.ID), Valid: true},
			CallerID: caller.ID,
			CallType: "voice",
			IsGroup:  strings.EqualFold(chat.Type, "group"),
			Status:   statuses[i%len(statuses)],
		})
		if strings.EqualFold(chat.Type, "group") {
			items = append(items, callSeed{
				ChatID:   sql.NullInt64{Int64: int64(chat.ID), Valid: true},
				CallerID: users[(i+1)%len(users)].ID,
				CallType: "video",
				IsGroup:  true,
				Status:   statuses[(i+1)%len(statuses)],
			})
		}
	}
	return items
}

func seedChatParticipants(conn *sql.DB, items []participantSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ChatID, item.UserID, item.Role})
	}
	return bulkInsertIgnore(conn, "chat_participants", []string{"chat_id", "user_id", "role"}, values)
}

func seedMessages(conn *sql.DB, items []messageSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ChatID, item.SenderID, item.Content, item.MediaID, item.EmojiID})
	}
	return bulkInsertIgnore(conn, "messages", []string{"chat_id", "sender_id", "content", "media_id", "emoji_id"}, values)
}

func seedCalls(conn *sql.DB, items []callSeed) (int64, error) {
	values := make([][]any, 0, len(items))
	for _, item := range items {
		values = append(values, []any{item.ChatID, item.CallerID, item.CallType, item.IsGroup, item.Status})
	}
	return bulkInsertIgnore(conn, "calls", []string{"chat_id", "caller_id", "call_type", "is_group", "status"}, values)
}

func bulkInsertIgnore(conn *sql.DB, table string, columns []string, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT IGNORE INTO ")
	builder.WriteString(table)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(") VALUES ")

	args := make([]any, 0, len(rows)*len(columns))
	for i, row := range rows {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for j := range columns {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("?")
		}
		builder.WriteString(")")
		args = append(args, row...)
	}

	result, err := conn.Exec(builder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed rows into %s: %w", table, err)
	}

	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", table, err)
	}
	return inserted, nil
}

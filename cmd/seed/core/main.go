package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type roleSeed struct {
	Name        string
	Description string
}

type violationRuleSeed struct {
	Title       string
	Description string
}

type emojiSeed struct {
	Code     string
	ImageURI string
}

type chatSeed struct {
	Type      string
	Name      string
	AvatarURI string
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

	if err := ensureCoreTables(conn); err != nil {
		log.Fatalf("ensure core tables failed: %v", err)
	}

	rolesInserted, err := seedRoles(conn, buildRoles())
	if err != nil {
		log.Fatalf("seed roles failed: %v", err)
	}

	violationsInserted, err := seedViolationRules(conn, buildViolationRules())
	if err != nil {
		log.Fatalf("seed violation rules failed: %v", err)
	}

	emojisInserted, err := seedEmojis(conn, buildEmojis())
	if err != nil {
		log.Fatalf("seed emojis failed: %v", err)
	}

	chatsInserted, err := seedChats(conn, buildChats(60))
	if err != nil {
		log.Fatalf("seed chats failed: %v", err)
	}

	fmt.Printf(
		"Seed core data: success (roles=%d, violation_rules=%d, emojis=%d, chats=%d)\n",
		rolesInserted,
		violationsInserted,
		emojisInserted,
		chatsInserted,
	)
}

func ensureCoreTables(conn *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL UNIQUE,
			description VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS violation_rules (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(150) NOT NULL UNIQUE,
			description VARCHAR(1000),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS emojis (
			id INT AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(50) NOT NULL UNIQUE,
			image_uri VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS chats (
			id INT AUTO_INCREMENT PRIMARY KEY,
			type VARCHAR(20) NOT NULL DEFAULT 'direct',
			name VARCHAR(100) NOT NULL UNIQUE,
			avatar_uri VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, statement := range statements {
		if _, err := conn.Exec(statement); err != nil {
			return fmt.Errorf("create core table: %w", err)
		}
	}

	return nil
}

func buildRoles() []roleSeed {
	return []roleSeed{
		{Name: "SUPER_ADMIN", Description: "Quyền quản trị toàn hệ thống."},
		{Name: "ADMIN", Description: "Quản lý hệ thống và nội dung."},
		{Name: "USER", Description: "Người dùng thông thường của hệ thống."},
		{Name: "COMM_CHAT", Description: "Quản lý cộng đồng và chat nội bộ."},
	}
}

func buildViolationRules() []violationRuleSeed {
	baseRules := []string{
		"Spam nội dung",
		"Spam liên kết",
		"Spam bình luận",
		"Quấy rối người dùng",
		"Ngôn từ thù ghét",
		"Kích động bạo lực",
		"Đe doạ an toàn cá nhân",
		"Nội dung khiêu dâm",
		"Nội dung bạo lực",
		"Xâm hại trẻ em",
		"Lừa đảo tài chính",
		"Mạo danh tài khoản",
		"Phishing / lừa lấy thông tin",
		"Phát tán mã độc",
		"Vi phạm bản quyền",
		"Phát tán thông tin cá nhân",
		"Tài liệu riêng tư bị rò rỉ",
		"Nội dung sai lệch",
		"Gian lận tương tác",
		"Bot hoạt động tự động",
		"Rác nội dung",
		"Off-topic kéo spam",
		"Quảng cáo trái phép",
		"Cờ bạc trái phép",
		"Nội dung ma tuý",
		"Nội dung tự hại",
		"Ngôn từ xúc phạm cá nhân",
		"Bình luận công kích",
		"Tài khoản giả mạo doanh nghiệp",
		"Lạm dụng tính năng báo cáo",
	}

	rules := make([]violationRuleSeed, 0, len(baseRules))
	for index, title := range baseRules {
		rules = append(rules, violationRuleSeed{
			Title:       title,
			Description: fmt.Sprintf("Quy tắc kiểm duyệt #%02d cho hành vi: %s.", index+1, strings.ToLower(title)),
		})
	}

	return rules
}

func buildEmojis() []emojiSeed {
	common := []emojiSeed{
		{Code: ":like:", ImageURI: "/seeds/emojis/like.png"},
		{Code: ":heart:", ImageURI: "/seeds/emojis/heart.png"},
		{Code: ":haha:", ImageURI: "/seeds/emojis/haha.png"},
		{Code: ":wow:", ImageURI: "/seeds/emojis/wow.png"},
		{Code: ":sad:", ImageURI: "/seeds/emojis/sad.png"},
		{Code: ":angry:", ImageURI: "/seeds/emojis/angry.png"},
		{Code: ":clap:", ImageURI: "/seeds/emojis/clap.png"},
		{Code: ":fire:", ImageURI: "/seeds/emojis/fire.png"},
		{Code: ":party:", ImageURI: "/seeds/emojis/party.png"},
		{Code: ":ok:", ImageURI: "/seeds/emojis/ok.png"},
		{Code: ":thumbsup:", ImageURI: "/seeds/emojis/thumbsup.png"},
		{Code: ":thumbsdown:", ImageURI: "/seeds/emojis/thumbsdown.png"},
		{Code: ":love:", ImageURI: "/seeds/emojis/love.png"},
		{Code: ":smile:", ImageURI: "/seeds/emojis/smile.png"},
		{Code: ":grin:", ImageURI: "/seeds/emojis/grin.png"},
		{Code: ":wink:", ImageURI: "/seeds/emojis/wink.png"},
		{Code: ":sunglasses:", ImageURI: "/seeds/emojis/sunglasses.png"},
		{Code: ":coffee:", ImageURI: "/seeds/emojis/coffee.png"},
		{Code: ":rocket:", ImageURI: "/seeds/emojis/rocket.png"},
		{Code: ":sparkles:", ImageURI: "/seeds/emojis/sparkles.png"},
	}

	emojis := make([]emojiSeed, 0, 120)
	emojis = append(emojis, common...)
	for index := 1; index <= 100; index++ {
		emojis = append(emojis, emojiSeed{
			Code:     fmt.Sprintf(":emoji_%03d:", index),
			ImageURI: fmt.Sprintf("/seeds/emojis/emoji-%03d.png", index),
		})
	}

	return emojis
}

func buildChats(total int) []chatSeed {
	chats := make([]chatSeed, 0, total)
	for index := 1; index <= total; index++ {
		chatType := "direct"
		if index%5 == 0 {
			chatType = "group"
		}

		chats = append(chats, chatSeed{
			Type:      chatType,
			Name:      fmt.Sprintf("Seed Chat %03d", index),
			AvatarURI: fmt.Sprintf("/seeds/chats/chat-%03d.png", index),
		})
	}

	return chats
}

func seedRoles(conn *sql.DB, roles []roleSeed) (int64, error) {
	values := make([][]any, 0, len(roles))
	for _, item := range roles {
		values = append(values, []any{item.Name, item.Description})
	}
	if _, err := bulkInsertIgnore(conn, "roles", []string{"name", "description"}, values); err != nil {
		return 0, err
	}

	if err := syncAllowedRoles(conn, roles); err != nil {
		return 0, err
	}

	return countRows(conn, "roles")
}

func syncAllowedRoles(conn *sql.DB, roles []roleSeed) error {
	if len(roles) == 0 {
		return nil
	}

	allowed := make([]string, 0, len(roles))
	for _, item := range roles {
		allowed = append(allowed, fmt.Sprintf("'%s'", item.Name))
	}
	query := fmt.Sprintf("DELETE FROM roles WHERE name NOT IN (%s)", strings.Join(allowed, ", "))
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("delete obsolete roles: %w", err)
	}
	return nil
}

func countRows(conn *sql.DB, table string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if err := conn.QueryRow(query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count rows for %s: %w", table, err)
	}
	return count, nil
}

func seedViolationRules(conn *sql.DB, rules []violationRuleSeed) (int64, error) {
	values := make([][]any, 0, len(rules))
	for _, item := range rules {
		values = append(values, []any{item.Title, item.Description})
	}
	return bulkInsertIgnore(conn, "violation_rules", []string{"title", "description"}, values)
}

func seedEmojis(conn *sql.DB, emojis []emojiSeed) (int64, error) {
	values := make([][]any, 0, len(emojis))
	for _, item := range emojis {
		values = append(values, []any{item.Code, item.ImageURI})
	}
	return bulkInsertIgnore(conn, "emojis", []string{"code", "image_uri"}, values)
}

func seedChats(conn *sql.DB, chats []chatSeed) (int64, error) {
	values := make([][]any, 0, len(chats))
	for _, item := range chats {
		values = append(values, []any{item.Type, item.Name, item.AvatarURI})
	}
	return bulkInsertIgnore(conn, "chats", []string{"type", "name", "avatar_uri"}, values)
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
	for index, row := range rows {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for columnIndex := range columns {
			if columnIndex > 0 {
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

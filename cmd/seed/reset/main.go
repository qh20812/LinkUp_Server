package reset

import (
	"fmt"
	"strings"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

// Run xoá toàn bộ bảng trong database hiện tại. Danh sách bảng được đọc từ
// information_schema thay vì liệt kê tay — nhờ vậy bảng do GORM AutoMigrate
// tạo (story, e2e key, session…) cũng được dọn, không còn sót dữ liệu cũ.
func Run(env config.Env) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("reset: connect: %w", err)
	}
	defer database.Close()

	rows, err := database.Query(
		`SELECT table_name FROM information_schema.tables
		 WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'`,
	)
	if err != nil {
		return fmt.Errorf("reset: list tables: %w", err)
	}

	tables := make([]string, 0, 64)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("reset: scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reset: iterate tables: %w", err)
	}

	if err := internal.Exec(database, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("reset: disable fk: %w", err)
	}

	for _, name := range tables {
		if strings.ContainsRune(name, '`') {
			return fmt.Errorf("reset: bảng có tên không hợp lệ: %q", name)
		}
		if err := internal.Exec(database, fmt.Sprintf("DROP TABLE IF EXISTS `%s`", name)); err != nil {
			return fmt.Errorf("reset: drop %s: %w", name, err)
		}
	}

	if err := internal.Exec(database, "SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return fmt.Errorf("reset: enable fk: %w", err)
	}

	fmt.Printf("reset: đã xoá %d bảng\n", len(tables))
	return nil
}

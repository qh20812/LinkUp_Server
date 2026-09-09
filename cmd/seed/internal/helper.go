package internal

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"linkup/config"
	"linkup/db"
)

func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Connect(env config.Env) (*sql.DB, error) {
	return db.ConnectDb(env)
}

func Ptr(s string) *string {
	return &s
}

func PtrTime(t time.Time) *time.Time {
	return &t
}

func Exec(db *sql.DB, query string, args ...any) error {
	_, err := db.Exec(query, args...)
	return err
}

// ContentUserIDs trả về danh sách user thường (bỏ 2 hệ thống role đầu: superadmin, admin).
func ContentUserIDs(state *SeedState) []string {
	return state.UserIDs[2:]
}

// ContentIndex trong [2, 19]: ánh xạ index vòng cho user thường.
func ContentIndex(i int) int {
	return 2 + (i % 18)
}

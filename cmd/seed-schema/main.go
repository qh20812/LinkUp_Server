// Command seed-schema tạo schema DB trống (reset + schema, KHÔNG seed data).
// Dùng 1 lần khi dựng MySQL local mới (vd trên VPS). Chạy từ thư mục
// chứa .env hoặc override bằng biến môi trường shell (shell thắng file).
package main

import (
	"log"

	"linkup/cmd/seed/reset"
	"linkup/cmd/seed/schema"
	"linkup/config"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	env := config.GetEnv()

	log.Println("Running seed-schema: reset")
	if err := reset.Run(env); err != nil {
		log.Fatalf("seed-schema reset failed: %v", err)
	}
	log.Println("Running seed-schema: schema")
	if err := schema.Run(env); err != nil {
		log.Fatalf("seed-schema schema failed: %v", err)
	}

	log.Println("seed-schema completed: empty schema ready")
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"linkup/cmd/seed/contribution"
	"linkup/cmd/seed/core"
	"linkup/cmd/seed/extended"
	"linkup/cmd/seed/internal"
	"linkup/cmd/seed/messaging"
	"linkup/cmd/seed/moderation"
	"linkup/cmd/seed/profiles"
	"linkup/cmd/seed/relationships"
	"linkup/cmd/seed/reset"
	"linkup/cmd/seed/schema"
	"linkup/cmd/seed/social"
	"linkup/cmd/seed/users"
	"linkup/config"
	"linkup/seeddata"
)

func main() {
	checkMedia := flag.Bool("check-media", false, "chỉ kiểm tra toàn bộ URI ảnh/video trong seeddata rồi thoát")
	flag.Parse()

	// Kiểm tra media không cần DB: dùng trước khi chụp ảnh báo cáo để chắc
	// không có ảnh nào 404.
	if *checkMedia {
		if err := checkMediaURIs(); err != nil {
			log.Fatalf("media check failed: %v", err)
		}
		return
	}

	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}
	env := config.GetEnv()

	// Kiểm tra chéo seeddata trước khi ghi bất cứ thứ gì — lỗi tham chiếu
	// (bài viết của user lạ, report trỏ tới cộng đồng không có…) sẽ lộ ra ngay.
	if err := seeddata.Validate(); err != nil {
		log.Fatalf("seeddata không hợp lệ: %v", err)
	}

	state := &internal.SeedState{}
	state.Init()

	steps := []struct {
		name string
		run  func() error
	}{
		{"reset", func() error { return reset.Run(env) }},
		{"schema", func() error { return schema.Run(env) }},
		{"users", func() error { return users.Run(env, state) }},
		{"core", func() error { return core.Run(env, state) }},
		{"profiles", func() error { return profiles.Run(env, state) }},
		// Cộng đồng phải có trước bài viết: posts.community_id là FK thật.
		{"relationships", func() error { return relationships.Run(env, state) }},
		{"social", func() error { return social.Run(env, state) }},
		{"contribution", func() error { return contribution.Run(env, state) }},
		{"messaging", func() error { return messaging.Run(env, state) }},
		{"moderation", func() error { return moderation.Run(env, state) }},
		{"extended", func() error { return extended.Run(env, state) }},
	}

	for _, step := range steps {
		log.Printf("Running seed: %s", step.name)
		if err := step.run(); err != nil {
			log.Fatalf("seed %s failed: %v", step.name, err)
		}
	}

	log.Println("All seed steps completed successfully")
}

// checkMediaURIs gửi HEAD tới mọi URI trong seeddata để phát hiện ảnh/video
// hỏng trước khi chụp ảnh báo cáo. Chỉ in ra mục lỗi và tổng kết.
func checkMediaURIs() error {
	uris := seeddata.AllMediaURIs()

	client := &http.Client{Timeout: 25 * time.Second}

	type result struct {
		uri    string
		status int
		err    error
	}

	// Wikimedia chặn khá gắt khi bị gọi dồn dập (HTTP 429), nên chỉ dùng 2
	// luồng, giãn cách giữa các request và thử lại vài lần. 429 không có nghĩa
	// là ảnh hỏng — chỉ là chưa kiểm tra được, nên tách khỏi lỗi thật.
	check := func(uri string) result {
		var last result
		for attempt := 0; attempt < 4; attempt++ {
			req, err := http.NewRequest(http.MethodHead, uri, nil)
			if err != nil {
				return result{uri: uri, err: err}
			}
			req.Header.Set("User-Agent", "LinkUpSeedCheck/1.0 (capstone; lien he: hotro@linkup.vn)")
			req.Header.Set("Accept", "image/*,video/*;q=0.9,*/*;q=0.8")

			resp, err := client.Do(req)
			if err != nil {
				last = result{uri: uri, err: err}
			} else {
				resp.Body.Close()
				last = result{uri: uri, status: resp.StatusCode}
				if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
					return last
				}
			}
			time.Sleep(time.Duration(3*(attempt+1)) * time.Second)
		}
		return last
	}

	jobs := make(chan string)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for uri := range jobs {
				time.Sleep(250 * time.Millisecond)
				results <- check(uri)
			}
		}()
	}

	go func() {
		for _, u := range uris {
			jobs <- u
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	failures, throttled := 0, 0
	for r := range results {
		switch {
		case r.err != nil:
			failures++
			fmt.Printf("FAIL %s — %v\n", r.uri, r.err)
		case r.status == http.StatusTooManyRequests:
			throttled++
			fmt.Printf("BỎ QUA (bị giới hạn tốc độ) %s — HTTP 429\n", r.uri)
		case r.status != http.StatusOK && r.status != http.StatusPartialContent:
			failures++
			fmt.Printf("FAIL %s — HTTP %d\n", r.uri, r.status)
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d/%d URI lỗi", failures, len(uris))
	}
	if throttled > 0 {
		fmt.Printf("OK: %d/%d URI trả 200/206, %d URI bị giới hạn tốc độ (chạy lại sau vài phút để kiểm tra nốt)\n",
			len(uris)-throttled, len(uris), throttled)
		return nil
	}
	fmt.Printf("OK: %d/%d URI trả 200/206\n", len(uris), len(uris))
	return nil
}

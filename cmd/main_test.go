package main

import (
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

// freeAddr trả về một địa chỉ 127.0.0.1 vừa được nhả cổng, dùng làm đầu vào cho
// listenWithRetry. Dùng "127.0.0.1" thay vì ":0" để test không phụ thuộc mọi
// interface mạng của máy.
func freeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("không lấy được cổng trống: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("không nhả được cổng: %v", err)
	}
	return addr
}

// occupy giữ cổng giống một instance server đang chạy.
func occupy(t *testing.T, addr string) net.Listener {
	t.Helper()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("không chiếm được cổng %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return ln
}

func TestListenWithRetry_FreePort(t *testing.T) {
	addr := freeAddr(t)

	ln, err := listenWithRetry(addr, 1, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("cổng trống nhưng vẫn lỗi: %v", err)
	}
	defer ln.Close()

	if ln.Addr().String() != addr {
		t.Fatalf("expected listener trên %s, got %s", addr, ln.Addr().String())
	}
}

func TestListenWithRetry_BusyPortReportsActionableError(t *testing.T) {
	addr := freeAddr(t)
	occupy(t, addr)

	start := time.Now()
	ln, err := listenWithRetry(addr, 3, 50*time.Millisecond)
	elapsed := time.Since(start)

	if ln != nil {
		ln.Close()
		t.Fatalf("không được trả listener khi cổng đang bận")
	}
	if err == nil {
		t.Fatalf("expected lỗi khi cổng đang bận")
	}

	// Thông điệp phải nói được việc cần làm, thay vì chỉ lỗi socket thô.
	msg := err.Error()
	if !strings.Contains(msg, addr) {
		t.Fatalf("thông điệp thiếu địa chỉ %s: %s", addr, msg)
	}
	if !strings.Contains(msg, "đang được tiến trình khác sử dụng") {
		t.Fatalf("thông điệp không hướng dẫn được gì: %s", msg)
	}
	// Lỗi gốc vẫn phải còn trong chuỗi để debug bằng log.
	if !strings.Contains(msg, "bind") && !strings.Contains(msg, "Only one usage") {
		t.Fatalf("thông điệp đã nuốt mất lỗi gốc: %s", msg)
	}
	if !isAddrInUse(err) {
		t.Fatalf("lỗi trả về phải được nhận diện là cổng đang bận: %v", err)
	}

	// 3 lần thử, 2 lần chờ => tối thiểu ~100ms; không được thử vô hạn.
	if elapsed > 2*time.Second {
		t.Fatalf("retry kéo dài quá lâu: %s", elapsed)
	}
}

func TestListenWithRetry_SucceedsAfterPortReleased(t *testing.T) {
	// Mô phỏng air restart: tiến trình cũ nhả cổng chậm hơn vài trăm ms.
	addr := freeAddr(t)
	old := occupy(t, addr)

	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = old.Close()
	}()

	ln, err := listenWithRetry(addr, 20, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("phải bind được sau khi cổng được nhả: %v", err)
	}
	defer ln.Close()
}

func TestListenWithRetry_NonBindErrorFailsFast(t *testing.T) {
	// Cổng không hợp lệ: không phải lỗi "đang bận" nên phải trả về ngay, không
	// chờ hết số lần thử.
	start := time.Now()
	ln, err := listenWithRetry("127.0.0.1:99999", 5, 200*time.Millisecond)
	elapsed := time.Since(start)

	if ln != nil {
		ln.Close()
	}
	if err == nil {
		t.Fatalf("expected lỗi với cổng không hợp lệ")
	}
	if strings.Contains(err.Error(), "đang được tiến trình khác sử dụng") {
		t.Fatalf("cổng không hợp lệ không được báo là đang bận: %s", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("lỗi không phải do bận cổng nhưng vẫn bị retry: %s", elapsed)
	}
}

func TestListenWithRetry_ClampsAttempts(t *testing.T) {
	// attempts < 1 vẫn phải thử đúng một lần, không trả lỗi rỗng.
	addr := freeAddr(t)

	ln, err := listenWithRetry(addr, 0, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("expected bind thành công với attempts=0: %v", err)
	}
	defer ln.Close()
}

func TestIsAddrInUse(t *testing.T) {
	if isAddrInUse(nil) {
		t.Fatalf("nil không phải lỗi bận cổng")
	}
	if isAddrInUse(errors.New("some other error")) {
		t.Fatalf("lỗi khác không được nhận diện là bận cổng")
	}

	addr := freeAddr(t)
	occupy(t, addr)
	_, err := net.Listen("tcp", addr)
	if err == nil {
		t.Skip("môi trường cho phép bind trùng cổng; bỏ qua kiểm tra errno")
	}
	if !isAddrInUse(err) {
		t.Fatalf("lỗi bind trùng cổng phải được nhận diện: %v", err)
	}
}

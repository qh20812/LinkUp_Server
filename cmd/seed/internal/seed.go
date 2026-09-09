package internal

type SeedUser struct {
	DisplayName string
	Username    string
	Email       string
	Status      string
}

// SeedUsers là nguồn dữ liệu duy nhất cho người dùng seed.
// Index 0 và 1 là hệ thống role (SUPER_ADMIN, ADMIN) — username theo role, không có tên cá nhân.
// Index 2–19 là người dùng thường: username là tên tiếng Việt bỏ dấu, giữ đúng tên hiển thị.
var SeedUsers = []SeedUser{
	// Hệ thống roles
	{DisplayName: "Quản Trị Viên Tối Cao", Username: "superadmin", Email: "superadmin@example.com", Status: "active"},
	{DisplayName: "Quản Trị Viên", Username: "admin", Email: "admin1@example.com", Status: "active"},
	// Người dùng thường — username khớp display name bỏ dấu
	{DisplayName: "Lê Hoàng Cường", Username: "le_hoang_cuong", Email: "alice@example.com", Status: "active"},
	{DisplayName: "Phạm Minh Đức", Username: "pham_minh_duc", Email: "bob@example.com", Status: "active"},
	{DisplayName: "Hoàng Thị Mai", Username: "hoang_thi_mai", Email: "charlie@example.com", Status: "active"},
	{DisplayName: "Đỗ Gia Huy", Username: "do_gia_huy", Email: "diana@example.com", Status: "active"},
	{DisplayName: "Vũ Ngọc Linh", Username: "vu_ngoc_linh", Email: "eve@example.com", Status: "banned"},
	{DisplayName: "Bùi Thanh Tùng", Username: "bui_thanh_tung", Email: "frank@example.com", Status: "active"},
	{DisplayName: "Đặng Hồng Nhung", Username: "dang_hong_nhung", Email: "grace@example.com", Status: "active"},
	{DisplayName: "Ngô Tuấn Kiệt", Username: "ngo_tuan_kiet", Email: "hank@example.com", Status: "active"},
	{DisplayName: "Phan Thùy Dương", Username: "phan_thuy_duong", Email: "ivy@example.com", Status: "suspended"},
	{DisplayName: "Lý Quốc Bảo", Username: "ly_quoc_bao", Email: "jack@example.com", Status: "active"},
	{DisplayName: "Tô Minh Tâm", Username: "to_minh_tam", Email: "kate@example.com", Status: "active"},
	{DisplayName: "Đinh Văn Hùng", Username: "dinh_van_hung", Email: "leo@example.com", Status: "active"},
	{DisplayName: "Hồ Cẩm Tú", Username: "ho_cam_tu", Email: "mila@example.com", Status: "active"},
	{DisplayName: "Mai Thanh Hà", Username: "mai_thanh_ha", Email: "neo@example.com", Status: "active"},
	{DisplayName: "Dương Khánh Vy", Username: "duong_khanh_vy", Email: "olivia@example.com", Status: "active"},
	{DisplayName: "Lưu Bích Ngọc", Username: "luu_bich_ngoc", Email: "peter@example.com", Status: "banned"},
	{DisplayName: "Chu Đức Minh", Username: "chu_duc_minh", Email: "quincy@example.com", Status: "active"},
	{DisplayName: "Trịnh Hoài Nam", Username: "trinh_hoai_nam", Email: "rachel@example.com", Status: "active"},
}
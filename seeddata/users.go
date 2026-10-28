package seeddata

// User là một tài khoản seed kèm hồ sơ đầy đủ. Key ổn định để các bước sau
// tham chiếu (bài viết, bình luận, chat, báo cáo) thay vì dựa vào index mảng.
type User struct {
	Key        string
	DisplayName string
	Username   string
	Email      string
	Status     string // active | suspended | banned
	IsSystem   bool   // superadmin / admin: không xuất hiện trong feed

	Avatar  string
	Cover   string
	Bio     string
	Phone   string
	DOB     string // YYYY-MM-DD
	Gender  string // male | female | other

	Location        string // địa chỉ hiển thị tự do
	HometownProv    string // mã tỉnh quê quán
	CurrentProv     string // mã tỉnh đang sống
	CurrentWard     string // mã phường/xã đang sống
	Work            string // enum work trong validations.ProfileValidation
	Education       string // enum education
	Website         string
	IsPrivate       bool
}

// Users là nguồn dữ liệu duy nhất cho tài khoản seed.
// Index 0, 1 là tài khoản hệ thống; index 2 là đối tác quảng cáo (role PARTNER);
// index 3 là tài khoản dùng để chụp ảnh báo cáo.
var Users = []User{
	{
		Key: "superadmin", DisplayName: "Quản Trị Viên Tối Cao", Username: "superadmin",
		Email: "superadmin@linkup.vn", Status: "active", IsSystem: true,
		Avatar: AvatarSuperAdmin, Cover: CoverHoanKiem,
		Bio: "Tài khoản vận hành hệ thống LinkUp.", Phone: "0901000001",
		DOB: "1985-01-15", Gender: "male", Location: "Hà Nội",
		HometownProv: "01", CurrentProv: "01", CurrentWard: "00004", Work: "it", Education: "master",
	},
	{
		Key: "admin", DisplayName: "Quản Trị Viên", Username: "admin",
		Email: "admin@linkup.vn", Status: "active", IsSystem: true,
		Avatar: AvatarAdmin, Cover: CoverPhoCo,
		Bio: "Đội ngũ kiểm duyệt nội dung LinkUp.", Phone: "0901000002",
		DOB: "1992-06-08", Gender: "female", Location: "Hà Nội",
		HometownProv: "01", CurrentProv: "01", CurrentWard: "00082", Work: "other", Education: "university",
	},
	{
		Key: "quocbao.pham", DisplayName: "Phạm Quốc Bảo", Username: "quocbao.pham",
		Email: "quocbao.pham@gmail.com", Status: "active",
		Avatar: AvatarQuocBao, Cover: CoverHoiAn,
		Bio: "Chủ quán Cà phê Nhà Gỗ ở Hội An. Mê cà phê rang mộc và những buổi chiều ít khách.",
		Phone: "0905112233", DOB: "1990-04-12", Gender: "male", Location: "Hội An, Đà Nẵng",
		HometownProv: "48", CurrentProv: "48", CurrentWard: "20242", Work: "business", Education: "university",
		Website: "nhagoc.coffee",
	},
	{
		Key: "minhanh.nguyen", DisplayName: "Nguyễn Minh Anh", Username: "minhanh.nguyen",
		Email: "minhanh.nguyen@gmail.com", Status: "active",
		Avatar: AvatarMinhAnh, Cover: CoverHoanKiem,
		Bio: "Marketing ở một agency Hà Nội. Thích cà phê muối, chạy bộ ven hồ và viết linh tinh.",
		Phone: "0987654321", DOB: "1998-09-21", Gender: "female", Location: "Ba Đình, Hà Nội",
		HometownProv: "01", CurrentProv: "01", CurrentWard: "00004", Work: "marketing", Education: "university",
	},
	{
		Key: "hoangnam.dev", DisplayName: "Trần Hoàng Nam", Username: "hoangnam.dev",
		Email: "hoangnam.dev@gmail.com", Status: "active",
		Avatar: AvatarHoangNam, Cover: CoverSaiGon,
		Bio: "Backend engineer, viết Go và đọc log lúc 2h sáng. Sài Gòn.",
		Phone: "0912345678", DOB: "1996-02-17", Gender: "male", Location: "Bến Thành, TP. Hồ Chí Minh",
		HometownProv: "40", CurrentProv: "79", CurrentWard: "26743", Work: "it", Education: "university",
		Website: "hoangnam.dev",
	},
	{
		Key: "thutrang_le", DisplayName: "Lê Thu Trang", Username: "thutrang_le",
		Email: "thutrang_le@gmail.com", Status: "active",
		Avatar: AvatarThuTrang, Cover: CoverDaNang,
		Bio: "Giáo viên IELTS 7 năm. Dạy ở Đà Nẵng, cuối tuần đi Đà Lạt trốn nắng.",
		Phone: "0934567890", DOB: "1994-11-03", Gender: "female", Location: "Hải Châu, Đà Nẵng",
		HometownProv: "48", CurrentProv: "48", CurrentWard: "20242", Work: "education", Education: "master",
	},
	{
		Key: "haiyen.vu", DisplayName: "Vũ Hải Yến", Username: "haiyen.vu",
		Email: "haiyen.vu@gmail.com", Status: "active",
		Avatar: AvatarHaiYen, Cover: CoverTrangAn,
		Bio: "Sinh viên năm 4 Kinh tế đối ngoại. Đang luyện IELTS để apply thạc sĩ.",
		Phone: "0978912345", DOB: "2003-06-25", Gender: "female", Location: "Cửa Nam, Hà Nội",
		HometownProv: "37", CurrentProv: "01", CurrentWard: "00082", Work: "student", Education: "university",
	},
	{
		Key: "tuankiet.dang", DisplayName: "Đặng Tuấn Kiệt", Username: "tuankiet.dang",
		Email: "tuankiet.dang@gmail.com", Status: "active",
		Avatar: AvatarTuanKiet, Cover: CoverDaNang,
		Bio: "Kỹ sư cầu đường, công trình ở Hải Phòng. Rảnh là đi phượt.",
		Phone: "0967891234", DOB: "1992-01-09", Gender: "male", Location: "Hồng Gai, Hải Phòng",
		HometownProv: "31", CurrentProv: "31", CurrentWard: "06685", Work: "engineering", Education: "university",
	},
	{
		Key: "ngochan_bui", DisplayName: "Bùi Ngọc Hân", Username: "ngochan_bui",
		Email: "ngochan_bui@gmail.com", Status: "active",
		Avatar: AvatarNgocHan, Cover: CoverCanTho,
		Bio: "Mẹ của hai đứa nhỏ ở Cần Thơ. Nấu ăn là cách mình xả stress.",
		Phone: "0945678901", DOB: "1988-08-14", Gender: "female", Location: "Ninh Kiều, Cần Thơ",
		HometownProv: "92", CurrentProv: "92", CurrentWard: "31135", Work: "housewife", Education: "college",
	},
	{
		Key: "nhatminh.ho", DisplayName: "Hồ Nhật Minh", Username: "nhatminh.ho",
		Email: "nhatminh.ho@gmail.com", Status: "active",
		Avatar: AvatarNhatMinh, Cover: CoverDaLat,
		Bio: "Nhiếp ảnh gia phong cảnh. Sống ở Đà Lạt, chụp sương sớm và cà phê chiều muộn.",
		Phone: "0923456789", DOB: "1995-05-30", Gender: "male", Location: "Xuân Hương, Đà Lạt, Lâm Đồng",
		HometownProv: "68", CurrentProv: "68", CurrentWard: "24781", Work: "arts", Education: "university",
		IsPrivate: true,
	},
	{
		Key: "thanhtung.ngo", DisplayName: "Ngô Thanh Tùng", Username: "thanhtung.ngo",
		Email: "thanhtung.ngo@gmail.com", Status: "active",
		Avatar: AvatarThanhTung, Cover: CoverSaiGon,
		Bio: "HLV chạy bộ. 42km đầu tiên năm 2019, đến giờ vẫn chưa dừng.",
		Phone: "0956789012", DOB: "1991-12-11", Gender: "male", Location: "Tân Định, TP. Hồ Chí Minh",
		HometownProv: "46", CurrentProv: "79", CurrentWard: "26737", Work: "service", Education: "university",
	},
	{
		Key: "khanhlinh.do", DisplayName: "Đỗ Khánh Linh", Username: "khanhlinh.do",
		Email: "khanhlinh.do@gmail.com", Status: "active",
		Avatar: AvatarKhanhLinh, Cover: CoverHoiAn,
		Bio: "Du học sinh Hà Nội, đang học ở Melbourne. Nhớ phở hơn nhớ nhà.",
		Phone: "0891234567", DOB: "2002-03-08", Gender: "female", Location: "Ba Đình, Hà Nội",
		HometownProv: "01", CurrentProv: "01", CurrentWard: "00004", Work: "student", Education: "university",
	},
	{
		Key: "phuongthao.mai", DisplayName: "Mai Phương Thảo", Username: "phuongthao.mai",
		Email: "phuongthao.mai@gmail.com", Status: "active",
		Avatar: AvatarPhuongThao, Cover: CoverDaLat,
		Bio: "Ngân hàng, làm ở quận 1. Thích chạy bộ và ăn uống cuối tuần.",
		Phone: "0887654321", DOB: "1993-07-19", Gender: "female", Location: "Bến Thành, TP. Hồ Chí Minh",
		HometownProv: "01", CurrentProv: "79", CurrentWard: "26743", Work: "finance", Education: "university",
	},
	{
		Key: "giahuy.ly", DisplayName: "Lý Gia Huy", Username: "giahuy.ly",
		Email: "giahuy.ly@gmail.com", Status: "active",
		Avatar: AvatarGiaHuy, Cover: CoverDaNang,
		Bio: "Founder một startup SaaS ở Đà Nẵng. Đang học cách nói không với họp.",
		Phone: "0876543210", DOB: "1990-10-02", Gender: "male", Location: "Hải Châu, Đà Nẵng",
		HometownProv: "79", CurrentProv: "48", CurrentWard: "20242", Work: "it", Education: "master",
		Website: "giahuy.ly",
	},
	{
		Key: "camtu.trinh", DisplayName: "Trịnh Cẩm Tú", Username: "camtu.trinh",
		Email: "camtu.trinh@gmail.com", Status: "active",
		Avatar: AvatarCamTu, Cover: CoverPhoCo,
		Bio: "Makeup artist tự do. Nhận makeup cô dâu, đi quay, chụp lookbook ở Hà Nội.",
		Phone: "0867891234", DOB: "1997-04-27", Gender: "female", Location: "Hoàn Kiếm, Hà Nội",
		HometownProv: "22", CurrentProv: "01", CurrentWard: "00070", Work: "arts", Education: "college",
	},
	{
		Key: "duclong.phan", DisplayName: "Phan Đức Long", Username: "duclong.phan",
		Email: "duclong.phan@gmail.com", Status: "active",
		Avatar: AvatarDucLong, Cover: CoverSaiGon,
		Bio: "Chạy xe công nghệ ở Sài Gòn. 6 năm rồi, thuộc từng con hẻm quận 3.",
		Phone: "0854321098", DOB: "1987-09-05", Gender: "male", Location: "An Khánh, TP. Hồ Chí Minh",
		HometownProv: "92", CurrentProv: "79", CurrentWard: "27094", Work: "service", Education: "highschool",
	},
	{
		Key: "bichha_to", DisplayName: "Tô Bích Hà", Username: "bichha_to",
		Email: "bichha_to@gmail.com", Status: "active",
		Avatar: AvatarBichHa, Cover: CoverHue,
		Bio: "Bác sĩ nội trú ở Huế. Trực đêm nhiều hơn ngủ.",
		Phone: "0832109876", DOB: "1989-02-23", Gender: "female", Location: "Thuận Hóa, Huế",
		HometownProv: "46", CurrentProv: "46", CurrentWard: "19789", Work: "healthcare", Education: "doctorate",
	},
	{
		Key: "anhdung.chu", DisplayName: "Chu Anh Dũng", Username: "anhdung.chu",
		Email: "anhdung.chu@gmail.com", Status: "suspended",
		Avatar: AvatarAnhDung, Cover: CoverCoffee,
		Bio: "Bán đồ gia dụng online. Ship toàn quốc.",
		Phone: "0823456789", DOB: "1994-06-16", Gender: "male", Location: "Thành Vinh, Nghệ An",
		HometownProv: "40", CurrentProv: "40", CurrentWard: "16681", Work: "trading", Education: "highschool",
	},
	{
		Key: "myduyen.duong", DisplayName: "Dương Mỹ Duyên", Username: "myduyen.duong",
		Email: "myduyen.duong@gmail.com", Status: "active",
		Avatar: AvatarMyDuyen, Cover: CoverCanTho,
		Bio: "Sinh viên năm 2 Đại học Cần Thơ. Mê chợ nổi và đồ ăn miền Tây.",
		Phone: "0812345678", DOB: "2004-01-31", Gender: "female", Location: "Ninh Kiều, Cần Thơ",
		HometownProv: "92", CurrentProv: "92", CurrentWard: "31135", Work: "student", Education: "university",
	},
	{
		Key: "quocthinh.lam", DisplayName: "Lâm Quốc Thịnh", Username: "quocthinh.lam",
		Email: "quocthinh.lam@gmail.com", Status: "banned",
		Avatar: AvatarQuocThinh, Cover: CoverCoffee,
		Bio: "Trồng cà phê ở Đắk Lắk. Ba đời làm nông.",
		Phone: "0798765432", DOB: "1985-11-20", Gender: "male", Location: "Buôn Ma Thuột, Đắk Lắk",
		HometownProv: "66", CurrentProv: "66", CurrentWard: "24133", Work: "agriculture", Education: "intermediate",
	},
}

// UserByKey tra cứu user theo key (dùng cho seed data tham chiếu chéo).
func UserByKey(key string) (User, bool) {
	for _, u := range Users {
		if u.Key == key {
			return u, true
		}
	}
	return User{}, false
}

// DemoUser là tài khoản đăng nhập dùng để chụp ảnh báo cáo.
const DemoUser = "minhanh.nguyen"

// PartnerUser là tài khoản sở hữu các chiến dịch quảng cáo seed.
const PartnerUser = "quocbao.pham"

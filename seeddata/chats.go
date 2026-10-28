package seeddata

// Message là một tin nhắn seed. AgeMin là số phút trước thời điểm seed.
type Message struct {
	From   string
	Text   string
	AgeMin int
}

// Chat là một hội thoại seed.
// Với chat direct, Members phải có đúng 2 người và MemberRoles để trống.
// Với chat group, Members[0] là CHAT_ADMIN.
type Chat struct {
	Key         string
	Type        string // direct | group
	Name        string
	Avatar      string
	Members     []string
	Messages    []Message
	UnreadFor   []string // thành viên chưa đọc tin cuối
	AllowAdd    bool     // group_chat_settings.allow_member_add
}

// Chats là nguồn dữ liệu duy nhất cho hội thoại seed.
var Chats = []Chat{
	{
		Key: "minhanh-thutrang", Type: "direct",
		Members: []string{"minhanh.nguyen", "thutrang_le"},
		Messages: []Message{
			{From: "minhanh.nguyen", Text: "Chị Trang ơi, em đang cân nhắc học IELTS lại. Mất gốc Writing quá.", AgeMin: 3200},
			{From: "thutrang_le", Text: "Em gửi chị một bài em viết gần nhất đi, chị xem lỗi ở đâu rồi mới nói được.", AgeMin: 3180},
			{From: "minhanh.nguyen", Text: "Dạ để em tìm lại. Mà học 1-1 thì học phí thế nào ạ?", AgeMin: 3160},
			{From: "thutrang_le", Text: "Chị dạy online 400k/buổi, 90 phút. Tuần 2 buổi là đủ nếu em tự học thêm.", AgeMin: 3120},
			{From: "minhanh.nguyen", Text: "Ok chị. Em sắp xếp lịch rồi nhắn chị tuần sau nhé.", AgeMin: 3100},
			{From: "thutrang_le", Text: "Ừ, cứ từ từ. Đừng học vì áp lực, học vì mình muốn thì mới bền 😊", AgeMin: 180},
			{From: "minhanh.nguyen", Text: "Nghe câu này nhẹ cả người. Cảm ơn chị!", AgeMin: 90},
		},
		UnreadFor: []string{"thutrang_le"},
	},
	{
		Key: "minhanh-hoangnam", Type: "direct",
		Members: []string{"minhanh.nguyen", "hoangnam.dev"},
		Messages: []Message{
			{From: "hoangnam.dev", Text: "Em ơi, bên anh đang tìm người làm content cho landing page. Em có nhận freelance không?", AgeMin: 1500},
			{From: "minhanh.nguyen", Text: "Dạ được anh, nhưng em chỉ nhận ngoài giờ. Deadline khoảng bao lâu ạ?", AgeMin: 1480},
			{From: "hoangnam.dev", Text: "Không gấp, 2 tuần. Budget 8 triệu cho 6 trang.", AgeMin: 1460},
			{From: "minhanh.nguyen", Text: "Ok anh, em gửi proposal tối nay.", AgeMin: 1440},
			{From: "hoangnam.dev", Text: "Anh gửi brief qua email rồi nhé. Có gì cứ hỏi.", AgeMin: 60},
		},
		UnreadFor: []string{"minhanh.nguyen"},
	},
	{
		Key: "ngochan-myduyen", Type: "direct",
		Members: []string{"ngochan_bui", "myduyen.duong"},
		Messages: []Message{
			{From: "myduyen.duong", Text: "Cô ơi, chủ nhật này cô có ở nhà không? Con qua gửi ít bánh xèo mẹ con làm.", AgeMin: 900},
			{From: "ngochan_bui", Text: "Có con, qua buổi sáng nhé. Cô làm sẵn nồi nước chấm.", AgeMin: 880},
			{From: "myduyen.duong", Text: "Dạ. Mà cô chỉ con cách làm bì với, con làm cứ bị khô.", AgeMin: 860},
			{From: "ngochan_bui", Text: "Thịt luộc xong phải để nguội hẳn rồi mới thái, thêm ít mỡ heo cho mềm. Cô chỉ tại bếp luôn.", AgeMin: 840},
			{From: "myduyen.duong", Text: "Dạ con ghi lại rồi ạ 🥰", AgeMin: 40},
		},
		UnreadFor: []string{"ngochan_bui"},
	},
	{
		Key: "thanhtung-phuongthao", Type: "direct",
		Members: []string{"thanhtung.ngo", "phuongthao.mai"},
		Messages: []Message{
			{From: "phuongthao.mai", Text: "Anh Tùng ơi, sáng mai nhóm mình chạy mấy giờ?", AgeMin: 700},
			{From: "thanhtung.ngo", Text: "5h30 em, ở Tao Đàn cổng Nguyễn Du. Chạy 8km nhẹ thôi.", AgeMin: 680},
			{From: "phuongthao.mai", Text: "Ok anh. Em mới mua giày mới, mai test luôn.", AgeMin: 660},
			{From: "thanhtung.ngo", Text: "Giày mới thì chạy 5km thôi, đừng 8. Về nhớ giãn cơ.", AgeMin: 640},
			{From: "phuongthao.mai", Text: "Dạ 😅", AgeMin: 120},
		},
	},
	{
		Key: "nhom-chay-530", Type: "group", Name: "Chạy Bộ 5h30",
		Avatar: ImgMarathon1,
		Members: []string{"thanhtung.ngo", "phuongthao.mai", "minhanh.nguyen", "hoangnam.dev"},
		AllowAdd: true,
		Messages: []Message{
			{From: "thanhtung.ngo", Text: "Sáng mai 5h30 Tao Đàn nhé cả nhà. Ai tới muộn thì chạy theo hướng Nguyễn Du.", AgeMin: 1300},
			{From: "phuongthao.mai", Text: "Em có mặt 🙋‍♀️", AgeMin: 1280},
			{From: "minhanh.nguyen", Text: "Em xin phép vắng, mai phải đi họp sớm ở công ty.", AgeMin: 1260},
			{From: "hoangnam.dev", Text: "Mình cũng vắng, đang ốm. Tuần sau quay lại.", AgeMin: 1240},
			{From: "thanhtung.ngo", Text: "Ok, vậy sáng mai có 2 người. Chạy nhẹ 8km, pace 6:30.", AgeMin: 1220},
			{From: "phuongthao.mai", Text: "Anh ơi, chạy xong ghé quán cà phê góc Võ Văn Tần không?", AgeMin: 220},
			{From: "thanhtung.ngo", Text: "Được, mà chỉ cà phê đen thôi, đừng bánh ngọt 😄", AgeMin: 200},
			{From: "phuongthao.mai", Text: "Biết rồi, huấn luyện viên khó tính.", AgeMin: 60},
		},
		UnreadFor: []string{"hoangnam.dev"},
	},
	{
		Key: "team-content", Type: "group", Name: "Team Content — Agency",
		Avatar: ImgLaptop1,
		Members: []string{"minhanh.nguyen", "camtu.trinh", "haiyen.vu"},
		AllowAdd: true,
		Messages: []Message{
			{From: "minhanh.nguyen", Text: "Brief khách mới: 3 video ngắn cho chiến dịch cuối năm. Deadline thứ 6.", AgeMin: 2600},
			{From: "camtu.trinh", Text: "Mình lo phần hình ảnh, cần 1 buổi quay. Thứ 3 được không?", AgeMin: 2580},
			{From: "haiyen.vu", Text: "Em rảnh thứ 3 buổi chiều. Sáng em còn lớp.", AgeMin: 2560},
			{From: "minhanh.nguyen", Text: "Vậy chốt thứ 3, 14h, studio quận 3. Chị gửi địa chỉ sau.", AgeMin: 2540},
			{From: "camtu.trinh", Text: "Nhớ mang theo 2 bộ đồ, phòng khi đổi ý concept.", AgeMin: 2520},
			{From: "haiyen.vu", Text: "Em note lại rồi ạ. Mà khách có yêu cầu giọng điệu gì không chị?", AgeMin: 120},
			{From: "minhanh.nguyen", Text: "Trẻ trung, đừng quá bán hàng. Cứ kể chuyện như người thật.", AgeMin: 100},
		},
		UnreadFor: []string{"camtu.trinh", "haiyen.vu"},
	},
	{
		Key: "gia-dinh-mien-tay", Type: "group", Name: "Gia Đình Miền Tây",
		Avatar: ImgCanTho1,
		Members: []string{"ngochan_bui", "myduyen.duong", "duclong.phan"},
		AllowAdd: true,
		Messages: []Message{
			{From: "ngochan_bui", Text: "Chủ nhật này giỗ ông, cả nhà về đủ nhé. Cô nấu sẵn.", AgeMin: 4000},
			{From: "duclong.phan", Text: "Dạ, con xin nghỉ chạy xe một ngày. Sáng con về sớm phụ dọn.", AgeMin: 3980},
			{From: "myduyen.duong", Text: "Con cũng về. Có cần mua gì ở Cần Thơ không cô?", AgeMin: 3960},
			{From: "ngochan_bui", Text: "Mua giúp cô 2kg tôm tươi với ít rau thơm. Tiền cô gửi lại.", AgeMin: 3940},
			{From: "duclong.phan", Text: "Cô đừng gửi tiền, con mua. Lâu rồi con chưa về.", AgeMin: 3920},
			{From: "myduyen.duong", Text: "Vậy con mua rau thơm với bánh tét 🙌", AgeMin: 3900},
			{From: "ngochan_bui", Text: "Ừ, cả nhà về sớm một chút cho vui.", AgeMin: 3000},
		},
	},
	{
		Key: "it-viet-nam-admin", Type: "group", Name: "IT Việt Nam — Ban Quản Trị",
		Avatar: ImgLaptop2,
		Members: []string{"hoangnam.dev", "giahuy.ly", "tuankiet.dang"},
		AllowAdd: false,
		Messages: []Message{
			{From: "hoangnam.dev", Text: "Tuần này có 3 bài tuyển dụng ẩn danh, mình đã xoá hết.", AgeMin: 2100},
			{From: "giahuy.ly", Text: "Ok. Mình nghĩ nên ghim một bài quy định về tuyển dụng để đỡ phải xoá.", AgeMin: 2080},
			{From: "tuankiet.dang", Text: "Đồng ý. Mình viết nháp rồi gửi cả nhóm xem nhé.", AgeMin: 2060},
			{From: "hoangnam.dev", Text: "Cảm ơn Tuấn Kiệt. Thêm phần khoảng lương bắt buộc là được.", AgeMin: 2040},
			{From: "giahuy.ly", Text: "Có ai nhận được report về bài Go 1.26 không? Mình thấy nội dung ổn.", AgeMin: 400},
			{From: "hoangnam.dev", Text: "Mình nhận 2 report, nhưng cả hai đều không nêu lý do. Để nguyên.", AgeMin: 380},
		},
		UnreadFor: []string{"tuankiet.dang"},
	},
	{
		Key: "quocbao-nhatminh", Type: "direct",
		Members: []string{"quocbao.pham", "nhatminh.ho"},
		Messages: []Message{
			{From: "nhatminh.ho", Text: "Anh Bảo ơi, tuần sau em qua Hội An chụp cho một nhóm khách. Quán anh còn chỗ ngồi ngoài sân không?", AgeMin: 1100},
			{From: "quocbao.pham", Text: "Có em, sáng thì thoải mái. Em qua sớm chụp nắng sớm là đẹp nhất.", AgeMin: 1080},
			{From: "nhatminh.ho", Text: "Dạ. Em gửi anh mấy ảnh đợt trước, anh dùng được thì lấy làm ảnh quán.", AgeMin: 1060},
			{From: "quocbao.pham", Text: "Cảm ơn em nhiều, ảnh đẹp quá. Hôm nào qua anh mời cà phê.", AgeMin: 1040},
		},
	},
	{
		Key: "khanhlinh-haiyen", Type: "direct",
		Members: []string{"khanhlinh.do", "haiyen.vu"},
		Messages: []Message{
			{From: "haiyen.vu", Text: "Chị Linh ơi, bên đó giờ mấy giờ rồi ạ?", AgeMin: 600},
			{From: "khanhlinh.do", Text: "Chị hơn em 4 tiếng, đang 22h. Em chưa ngủ à?", AgeMin: 580},
			{From: "haiyen.vu", Text: "Em đang ôn Writing, mà viết mãi không ra ý.", AgeMin: 560},
			{From: "khanhlinh.do", Text: "Em gửi đề đây chị gợi ý hướng làm cho. Đừng cố viết hoa mỹ, viết rõ là được.", AgeMin: 540},
			{From: "haiyen.vu", Text: "Dạ em gửi ngay. Cảm ơn chị 🥹", AgeMin: 520},
			{From: "khanhlinh.do", Text: "Ừ, mà ngủ đi. 1h sáng rồi, thức khuya viết càng dở.", AgeMin: 30},
		},
		UnreadFor: []string{"haiyen.vu"},
	},
}

// ChatByKey tra cứu hội thoại theo key.
func ChatByKey(key string) (Chat, bool) {
	for _, c := range Chats {
		if c.Key == key {
			return c, true
		}
	}
	return Chat{}, false
}

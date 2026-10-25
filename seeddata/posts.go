package seeddata

// Reaction là một lượt thả cảm xúc. Emoji là code không có dấu hai chấm
// (like, love, haha, wow, sad, angry, clap, fire, heart, rocket).
type Reaction struct {
	User  string
	Emoji string
}

// Comment là bình luận seed. AgeMin là số phút trước thời điểm seed.
type Comment struct {
	Author  string
	Text    string
	AgeMin  int
	Replies []Comment
}

// Post là một bài viết seed đầy đủ. Title rỗng nghĩa là bài không có tiêu đề
// (đúng như luồng đăng bài thật: content-only hoặc media-only).
type Post struct {
	Key       string
	Author    string
	Title     string
	Content   string
	Status    string // "" = public
	Community string // key cộng đồng, "" = bài cá nhân
	AgeHours  float64
	Views     int
	Pinned    bool
	NoComment bool
	Media     []Media
	ShareOf   string // key bài gốc nếu là repost
	ShareText string
	Reactions []Reaction
	Comments  []Comment
}

// Posts là nguồn dữ liệu duy nhất cho bài viết seed.
//
// 10 bài mới nhất (AgeHours nhỏ nhất) được thiết kế để trang feed đầu tiên
// đã đủ các dạng bài: video, album 4 ảnh, media-only, bài dài có "Xem thêm",
// title+content, repost, bài cộng đồng, hashtag + mention, ảnh đơn, 2 ảnh.
var Posts = []Post{
	{
		Key: "nhatminh-con-dao", Author: "nhatminh.ho", AgeHours: 0.5, Views: 4820,
		Content: "Ba ngày ở Côn Đảo, sáng nào cũng ra bãi Đầm Trầu từ 5h. Nước trong tới mức đứng trên bờ vẫn thấy rùa biển bơi qua. Mình quay được đúng một đoạn ngắn, còn lại toàn ngồi im nhìn.\n\n#ConDao #LanBien #BienDaoVietNam",
		Media:   []Media{Vid(VideoSeaTurtle, 18, 3120000)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "thutrang_le", Emoji: "wow"},
			{User: "haiyen.vu", Emoji: "heart"}, {User: "khanhlinh.do", Emoji: "love"},
			{User: "thanhtung.ngo", Emoji: "clap"}, {User: "phuongthao.mai", Emoji: "love"},
			{User: "myduyen.duong", Emoji: "wow"}, {User: "bichha_to", Emoji: "heart"},
			{User: "giahuy.ly", Emoji: "fire"}, {User: "tuankiet.dang", Emoji: "love"},
		},
		Comments: []Comment{
			{Author: "khanhlinh.do", AgeMin: 22, Text: "Trời ơi đẹp quá. Cho mình xin lịch trình với chi phí được không ạ?", Replies: []Comment{
				{Author: "nhatminh.ho", AgeMin: 18, Text: "3 ngày 2 đêm tầm 5 triệu/người, bay từ Sài Gòn. Mình gửi bạn lịch trình trong tin nhắn nhé."},
			}},
			{Author: "bichha_to", AgeMin: 20, Text: "Nước xanh kiểu này ngoài Côn Đảo hơi khó tìm. Đi tháng mấy là ít mưa nhất vậy bạn?"},
			{Author: "myduyen.duong", AgeMin: 15, Text: "Rùa biển 🥹 Thêm Côn Đảo vào danh sách ngay."},
			{Author: "thanhtung.ngo", AgeMin: 9, Text: "Chạy bộ ven biển Côn Đảo sáng sớm là một trong những cung đẹp nhất mình từng chạy."},
		},
	},
	{
		Key: "thutrang-dalat", Author: "thutrang_le", AgeHours: 1.1, Views: 3260,
		Title:   "Đà Lạt 3 ngày 2 đêm: lịch trình mình vừa đi về",
		Content: "Mình đi Đà Lạt đúng đợt se lạnh, sáng 14 độ, trưa lên 24. Lịch trình không quá dày vì đi để nghỉ chứ không phải để check-in.\n\nNgày 1: ga Đà Lạt buổi sáng, ăn cơm tấm, chiều dạo hồ Xuân Hương. Ngày 2: dậy 5h đi Cầu Đất săn mây, về ăn bánh mì xíu mại, chiều ghé nhà thờ Con Gà. Ngày 3: cà phê sáng rồi ra sân bay.\n\nChi phí cả chuyến khoảng 3,2 triệu/người, chưa tính vé máy bay. #DaLat #SanMay #DuLichVietNam",
		Media: []Media{Img(ImgDaLat1), Img(ImgDaLat4), Img(ImgDaLat2), Img(ImgDaLat3)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "haiyen.vu", Emoji: "heart"},
			{User: "khanhlinh.do", Emoji: "love"}, {User: "myduyen.duong", Emoji: "wow"},
			{User: "tuankiet.dang", Emoji: "like"}, {User: "ngochan_bui", Emoji: "love"},
			{User: "nhatminh.ho", Emoji: "like"}, {User: "camtu.trinh", Emoji: "heart"},
			{User: "phuongthao.mai", Emoji: "love"}, {User: "giahuy.ly", Emoji: "like"},
			{User: "bichha_to", Emoji: "heart"}, {User: "duclong.phan", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "minhanh.nguyen", AgeMin: 48, Text: "Cầu Đất sáng nào cũng có mây không chị? Em đi 2 lần rồi mà lần nào cũng trắng trời sương."},
			{Author: "nhatminh.ho", AgeMin: 44, Text: "Ảnh ga Đà Lạt chụp khung này đẹp nè. Sáng sớm tầm 6h30 nắng xiên là chuẩn nhất.", Replies: []Comment{
				{Author: "thutrang_le", AgeMin: 40, Text: "Cảm ơn em, chị canh đúng 6h40 luôn 😄"},
			}},
			{Author: "myduyen.duong", AgeMin: 35, Text: "3,2 triệu là tiết kiệm quá, em đi 2 đêm mà hết gần 5 triệu."},
			{Author: "tuankiet.dang", AgeMin: 28, Text: "Lưu lại, tháng sau chạy xe máy từ Hải Phòng vào Đà Lạt luôn."},
		},
	},
	{
		Key: "myduyen-cho-noi", Author: "myduyen.duong", AgeHours: 1.7, Views: 1980,
		Media: []Media{Img(ImgCanTho1), Img(ImgCanTho2), Img(ImgCanTho3)},
		Reactions: []Reaction{
			{User: "ngochan_bui", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "wow"},
			{User: "thutrang_le", Emoji: "heart"}, {User: "haiyen.vu", Emoji: "love"},
			{User: "duclong.phan", Emoji: "like"}, {User: "bichha_to", Emoji: "like"},
			{User: "khanhlinh.do", Emoji: "wow"},
		},
		Comments: []Comment{
			{Author: "ngochan_bui", AgeMin: 55, Text: "Chợ nổi Cái Răng sáng sớm đúng là phải đi thuyền nhỏ mới thấy hết. Con ăn hủ tiếu trên ghe chưa?"},
			{Author: "myduyen.duong", AgeMin: 50, Text: "Dạ rồi cô, 25 nghìn một tô, ngon hơn ăn trong quán luôn."},
			{Author: "haiyen.vu", AgeMin: 30, Text: "Nhìn mấy ghe trái cây mà thèm. Miền Tây chưa đi lần nào 😢"},
		},
	},
	{
		Key: "minhanh-ot", Author: "minhanh.nguyen", AgeHours: 2.3, Views: 5140,
		Content: "Tháng này mình OT 11 ngày. Không phải vì deadline gấp, mà vì khách đổi brief liên tục và mọi thứ đều được coi là \"gấp\".\n\nTối qua 9h rưỡi mới ra khỏi toà nhà, đứng đợi grab ở Nguyễn Trãi nhìn dòng xe mà tự nhiên thấy mệt. Về đến nhà, mở laptop định trả lời nốt cái email thì ngồi im 20 phút không gõ được chữ nào.\n\nSáng nay mình thử một việc rất nhỏ: đặt giờ kết thúc làm việc, 6h30 tắt máy, đi bộ một vòng hồ. Nghe thì đơn giản nhưng hoá ra khó nhất là bước ra khỏi cái cảm giác \"mình chưa làm đủ\".\n\nKhông biết có ai đang giống mình không. Mọi người đang cân bằng chuyện này thế nào? #CongSo #WorkLifeBalance",
		Reactions: []Reaction{
			{User: "hoangnam.dev", Emoji: "like"}, {User: "thutrang_le", Emoji: "love"},
			{User: "phuongthao.mai", Emoji: "like"}, {User: "giahuy.ly", Emoji: "like"},
			{User: "bichha_to", Emoji: "love"}, {User: "camtu.trinh", Emoji: "like"},
			{User: "haiyen.vu", Emoji: "sad"}, {User: "khanhlinh.do", Emoji: "like"},
			{User: "tuankiet.dang", Emoji: "like"}, {User: "duclong.phan", Emoji: "like"},
			{User: "myduyen.duong", Emoji: "sad"}, {User: "ngochan_bui", Emoji: "love"},
			{User: "thanhtung.ngo", Emoji: "like"}, {User: "nhatminh.ho", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "phuongthao.mai", AgeMin: 95, Text: "Mình cũng bank, tháng nào cũng có tuần như vậy. Cái câu \"mình chưa làm đủ\" đúng là bệnh chung."},
			{Author: "hoangnam.dev", AgeMin: 88, Text: "Đặt giờ tắt máy là bước đầu tiên. Mình dùng cách: 6h chiều đổi trạng thái Slack sang \"off\", ai cần gấp thì gọi.", Replies: []Comment{
				{Author: "minhanh.nguyen", AgeMin: 80, Text: "Chuẩn anh, mà bên em khách nhắn qua Zalo cơ, khó hơn 😅"},
			}},
			{Author: "bichha_to", AgeMin: 70, Text: "Trực đêm 3 ca/tuần nên mình hiểu cảm giác đó. Cố lên, sức khoẻ vẫn là trên hết."},
			{Author: "camtu.trinh", AgeMin: 62, Text: "Đi bộ một vòng hồ nghe đơn giản mà hiệu quả thật. Mình thì hay ngồi im nghe nhạc 15 phút trước khi ngủ."},
			{Author: "giahuy.ly", AgeMin: 44, Text: "Bên mình áp dụng \"no meeting Friday\" được 6 tháng, đỡ hẳn. Nhưng OT thì vẫn còn."},
		},
	},
	{
		Key: "giahuy-khoi-nghiep", Author: "giahuy.ly", AgeHours: 2.9, Views: 3980,
		Title:   "Ba năm khởi nghiệp, mình học được 4 điều không có trong sách",
		Content: "Hôm nay công ty tròn 3 tuổi. Không phải câu chuyện gọi vốn triệu đô, chỉ là một team 12 người ở Đà Nẵng, doanh thu đủ sống và bắt đầu có lãi.\n\n1. Khách hàng đầu tiên quan trọng hơn ý tưởng đẹp. Mình mất 8 tháng build một sản phẩm không ai cần.\n\n2. Đừng thuê người giỏi nhất, thuê người phù hợp nhất với giai đoạn hiện tại.\n\n3. Dòng tiền là sự thật duy nhất. Mọi thứ khác chỉ là cảm giác.\n\n4. Nói \"không\" với khách hàng sai giúp công ty sống lâu hơn.\n\nCảm ơn 12 người đã ở lại. #KhoiNghiep #Startup",
		Reactions: []Reaction{
			{User: "hoangnam.dev", Emoji: "clap"}, {User: "minhanh.nguyen", Emoji: "love"},
			{User: "quocbao.pham", Emoji: "like"}, {User: "thutrang_le", Emoji: "clap"},
			{User: "phuongthao.mai", Emoji: "like"}, {User: "tuankiet.dang", Emoji: "like"},
			{User: "bichha_to", Emoji: "heart"}, {User: "nhatminh.ho", Emoji: "fire"},
			{User: "thanhtung.ngo", Emoji: "clap"},
		},
		Comments: []Comment{
			{Author: "hoangnam.dev", AgeMin: 120, Text: "Điều 3 đúng kinh khủng. Bên em từng có tháng doanh thu đẹp nhưng tiền về chậm 45 ngày, suýt không trả lương."},
			{Author: "quocbao.pham", AgeMin: 100, Text: "Chúc mừng em. Quán cà phê của anh cũng 5 năm rồi, điều 1 và 4 là hai điều anh ước biết sớm hơn."},
			{Author: "minhanh.nguyen", AgeMin: 76, Text: "\"Dòng tiền là sự thật duy nhất\" — mình sẽ dán câu này lên màn hình."},
		},
	},
	{
		Key: "phuongthao-repost", Author: "phuongthao.mai", AgeHours: 3.5, Views: 1420,
		ShareOf: "thanhtung-chay-bo", ShareText: "Đọc xong tự nhiên muốn xỏ giày đi chạy luôn 🏃‍♀️",
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "like"}, {User: "thanhtung.ngo", Emoji: "love"},
			{User: "haiyen.vu", Emoji: "like"}, {User: "hoangnam.dev", Emoji: "like"},
			{User: "khanhlinh.do", Emoji: "heart"},
		},
		Comments: []Comment{
			{Author: "thanhtung.ngo", AgeMin: 40, Text: "Chạy đi, sáng mai 5h30 mình có nhóm ở Tao Đàn, vào chạy chung cho vui."},
		},
	},
	{
		Key: "ngochan-bun-cha", Author: "ngochan_bui", AgeHours: 4.1, Views: 2760,
		Community: "an-gi-hom-nay",
		Title:     "Bún chả tự làm ở nhà, lần đầu mà ăn được",
		Content:   "Nhà mình ở Cần Thơ nên bún chả Hà Nội chỉ ăn được khi ra Bắc công tác. Cuối tuần rồi mình thử làm: thịt ba chỉ ướp nước mắm, đường, hành tím, nướng bằng nồi chiên không dầu 200 độ trong 18 phút, trở mặt giữa chừng.\n\nNước chấm thì mình pha theo kiểu miền Nam hơi ngọt hơn một chút. Cả nhà ăn hết sạch, đứa lớn đòi mai ăn tiếp.\n\nAi có mẹo nướng thịt cho thơm kiểu than hoa không ạ? #AnGiHomNay #BunCha",
		Media: []Media{Img(ImgBunCha1), Img(ImgBunCha2), Img(ImgBunCha3)},
		Reactions: []Reaction{
			{User: "myduyen.duong", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "love"},
			{User: "thutrang_le", Emoji: "like"}, {User: "haiyen.vu", Emoji: "wow"},
			{User: "bichha_to", Emoji: "like"}, {User: "camtu.trinh", Emoji: "love"},
		},
		Comments: []Comment{
			{Author: "thutrang_le", AgeMin: 70, Text: "Nướng nồi chiên không dầu bị khô, chị bọc giấy bạc 10 phút đầu rồi mở ra nướng thêm 8 phút là mềm hơn nhiều."},
			{Author: "myduyen.duong", AgeMin: 60, Text: "Cô ơi nhìn ngon quá, cháu qua ăn ké được không ạ 😆"},
			{Author: "ngochan_bui", AgeMin: 52, Text: "Qua đi con, cô còn một hộp đây."},
		},
	},
	{
		Key: "haiyen-ielts", Author: "haiyen.vu", AgeHours: 6.5, Views: 1640,
		Content: "Nay lớp @thutrang_le chữa đề Writing Task 2, cả lớp viết \"partly agree\" xong cô bảo: nếu không dám khẳng định thì đừng mở bài bằng \"I completely agree\". Nghe xong thấy đúng cho cả ngoài đời.\n\nMục tiêu tháng 12 của mình là 7.0, hiện tại đang 6.0. Ai đang ôn thì vào comment mục tiêu nhé, mình lập nhóm học cho vui. #IELTS #HocTiengAnh",
		Reactions: []Reaction{
			{User: "thutrang_le", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "like"},
			{User: "myduyen.duong", Emoji: "like"}, {User: "khanhlinh.do", Emoji: "clap"},
			{User: "bichha_to", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "thutrang_le", AgeMin: 45, Text: "6.0 lên 7.0 trong 2 tháng là hoàn toàn làm được nếu giữ Writing đều. Cô tin em."},
			{Author: "myduyen.duong", AgeMin: 38, Text: "Em cũng đang 5.5, mục tiêu 6.5 tháng 3 năm sau. Cho em vào nhóm với ạ!"},
			{Author: "khanhlinh.do", AgeMin: 30, Text: "Mình 7.5 rồi, nếu cần review essay thì gửi mình xem cho."},
		},
	},
	{
		Key: "bichha-truc-dem", Author: "bichha_to", AgeHours: 5.3, Views: 2210,
		Title:   "Trực đêm thứ 3 trong tuần và một chuyện nhỏ mình mới hiểu",
		Content: "2h sáng, hành lang bệnh viện vắng tới mức nghe được tiếng máy thở ở phòng cuối. Mình ngồi uống ly sữa nóng, tự nhiên nghĩ về câu hỏi của một bệnh nhân chiều nay: \"Bác sĩ có sợ không?\".\n\nCó chứ. Sợ chứ. Nhưng cái sợ đó phải để sau khi mình rửa tay, mặc áo, và bước vào phòng.\n\nSáng nay tan ca, mình ra ngoài, thấy trời Huế mưa nhỏ. Đi ăn bún bò rồi về ngủ. Một đêm bình thường. #BacSi #TrueStory",
		Media: []Media{Img(ImgHospital)},
		Reactions: []Reaction{
			{User: "thutrang_le", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "heart"},
			{User: "ngochan_bui", Emoji: "love"}, {User: "haiyen.vu", Emoji: "heart"},
			{User: "khanhlinh.do", Emoji: "clap"}, {User: "duclong.phan", Emoji: "love"},
			{User: "thanhtung.ngo", Emoji: "clap"},
		},
		Comments: []Comment{
			{Author: "ngochan_bui", AgeMin: 40, Text: "Cảm ơn bác sĩ. Mẹ mình nằm viện tháng trước, mấy đêm trực của các bạn là thứ mình nhớ nhất."},
			{Author: "minhanh.nguyen", AgeMin: 28, Text: "\"Cái sợ đó phải để sau khi mình rửa tay\" — câu này hay quá."},
		},
	},
	{
		Key: "tuankiet-cau", Author: "tuankiet.dang", AgeHours: 6.1, Views: 1130,
		Content: "Công trường tuần này mưa 4 ngày, tiến độ chậm gần 2 tuần. Cả đội phải chia ca đổ bê tông ban đêm để kịp mốc.\n\nĐứng trên cầu lúc 11h đêm, gió sông thổi lạnh, nhìn về phía thành phố thấy đèn còn sáng. Làm cầu nhiều năm rồi mà mỗi lần thấy xe chạy qua đoạn mình thi công vẫn thấy vui.\n\n#KySu #XayDung",
		Media: []Media{Img(ImgBridge), Img(ImgHanoiRain1)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "clap"}, {User: "thanhtung.ngo", Emoji: "like"},
			{User: "hoangnam.dev", Emoji: "like"}, {User: "bichha_to", Emoji: "clap"},
		},
		Comments: []Comment{
			{Author: "hoangnam.dev", AgeMin: 50, Text: "Đổ bê tông đêm mà mưa thì căng thật. Bên anh làm phần mềm, deadline trượt còn đỡ hơn."},
		},
	},
	{
		Key: "haiyen-friend", Author: "haiyen.vu", AgeHours: 7.1, Views: 640, Status: "friend",
		Content: "Chỉ bạn bè mới thấy bài này: lớp mình vừa bảo vệ đồ án xong, cả nhóm kéo ra quán ăn tới 11h đêm. Bốn năm đại học kết thúc bằng một bữa lẩu và một đứa khóc.\n\n#SinhVien #TotNghiep",
		Media:     []Media{Img(ImgStudents1)},
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "love"}, {User: "camtu.trinh", Emoji: "heart"}},
		Comments:  []Comment{{Author: "minhanh.nguyen", AgeMin: 60, Text: "Chúc mừng em! Nhớ hồi chị bảo vệ xong cũng ngồi khóc ở quán trà sữa."}},
	},
	{
		Key: "minhanh-bigbang", Author: "minhanh.nguyen", AgeHours: 8, Views: 7320,
		Content: "Tối qua Mỹ Đình đông nghẹt. Mình không phải fan cứng nhưng đứng giữa đám đông hát theo đoạn cuối thì hiểu vì sao người ta bay từ Sài Gòn ra chỉ để nghe một bài.\n\nVề đến nhà 12h, sáng nay họp 8h30 mà đầu vẫn còn tiếng bass. Đáng.\n\n#BigBang #MyDinh #Concert",
		Media: []Media{Img(ImgConcert), Img(ImgFireworks2)},
		Reactions: []Reaction{
			{User: "haiyen.vu", Emoji: "love"}, {User: "khanhlinh.do", Emoji: "fire"},
			{User: "camtu.trinh", Emoji: "love"}, {User: "thutrang_le", Emoji: "like"},
			{User: "myduyen.duong", Emoji: "wow"}, {User: "duclong.phan", Emoji: "like"},
			{User: "hoangnam.dev", Emoji: "like"}, {User: "phuongthao.mai", Emoji: "love"},
			{User: "tuankiet.dang", Emoji: "like"}, {User: "nhatminh.ho", Emoji: "wow"},
			{User: "ngochan_bui", Emoji: "like"}, {User: "bichha_to", Emoji: "heart"},
		},
		Comments: []Comment{
			{Author: "khanhlinh.do", AgeMin: 180, Text: "Mình săn vé 3 lần không được, ngồi nhà xem livestream mà cay 😭"},
			{Author: "haiyen.vu", AgeMin: 160, Text: "Vé đứng khu B có 1,2 triệu thôi mà vào phải chen, đứng 4 tiếng. Vẫn xứng đáng."},
			{Author: "camtu.trinh", AgeMin: 140, Text: "Makeup cho khách đi concert tối qua, bạn nào cũng mắt lấp lánh. Nhìn ảnh này nhớ không khí quá."},
			{Author: "duclong.phan", AgeMin: 120, Text: "Tối qua chở khách từ Mỹ Đình về Hà Đông, đường kẹt từ 10h tới 11h rưỡi 🥲"},
		},
	},
	{
		Key: "hoangnam-go", Author: "hoangnam.dev", AgeHours: 10, Views: 4460, Community: "it-viet-nam",
		Title:   "Go 1.26 có gì đáng để nâng cấp? Góc nhìn từ người chạy production",
		Content: "Team mình vừa lên 1.26 cho hai service chính. Không có gì quá đột phá nhưng có mấy điểm đáng chú ý:\n\n- Compiler nhanh hơn rõ rệt ở codebase lớn, build CI giảm từ 2m10 xuống 1m35.\n- Một số cải thiện ở runtime giúp giảm allocation trong hot path.\n- Tooling ổn định hơn, `go vet` bắt thêm vài case.\n\nĐiều mình thích nhất không phải tính năng mới, mà là việc nâng cấp không làm vỡ gì cả. Với một team nhỏ thì đó là điều quan trọng nhất.\n\nMọi người đã lên 1.26 chưa? #Golang #Backend",
		Media: []Media{Img(ImgLaptop1)},
		Reactions: []Reaction{
			{User: "giahuy.ly", Emoji: "like"}, {User: "minhanh.nguyen", Emoji: "like"},
			{User: "tuankiet.dang", Emoji: "wow"}, {User: "thanhtung.ngo", Emoji: "like"},
			{User: "bichha_to", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "giahuy.ly", AgeMin: 130, Text: "Bên mình vẫn 1.24 vì có dependency cũ. Build chậm mà chưa dám lên."},
			{Author: "hoangnam.dev", AgeMin: 120, Text: "Bác thử tách module ra trước, bên em cũng mất 2 sprint mới dọn xong."},
		},
	},
	{
		Key: "camtu-makeup", Author: "camtu.trinh", AgeHours: 12, Views: 3050,
		Title:   "Look makeup đi tiệc cuối năm: da mỏng, mắt nhấn, môi trầm",
		Content: "Cuối năm là mùa makeup dày nhất của mình. Năm nay khách chuộng kiểu \"da như không trang điểm\" hơn là lớp nền dày.\n\nBa thứ mình luôn dùng: kem lót kiềm dầu, phấn phủ bột mịn, và một cây cọ nhỏ để tán kem che khuyết điểm đúng chỗ. Mắt thì nhấn bằng eyeliner nâu thay vì đen, môi chọn tone trầm cho hợp ánh đèn tiệc.\n\nẢnh dưới là look mình làm cho một bạn đi dự cưới ở Hà Nội tuần trước. #Makeup #LamDep",
		Media: []Media{Img(ImgMakeup1), Img(ImgMakeup2), Img(ImgAodai2)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "haiyen.vu", Emoji: "love"},
			{User: "khanhlinh.do", Emoji: "heart"}, {User: "phuongthao.mai", Emoji: "like"},
			{User: "thutrang_le", Emoji: "like"}, {User: "myduyen.duong", Emoji: "love"},
		},
		Comments: []Comment{
			{Author: "minhanh.nguyen", AgeMin: 200, Text: "Cây cọ nhỏ là bí quyết thật. Mình dùng cọ to tán xong lúc nào cũng bị loang."},
			{Author: "myduyen.duong", AgeMin: 180, Text: "Chị ơi makeup cô dâu ở Cần Thơ có nhận không ạ? Em cưới tháng 4."},
			{Author: "camtu.trinh", AgeMin: 170, Text: "Chị chỉ nhận ở Hà Nội thôi em ạ, nhưng em tìm mấy bạn makeup miền Tây cũng nhiều bạn giỏi lắm."},
		},
	},
	{
		Key: "thanhtung-chay-bo", Author: "thanhtung.ngo", AgeHours: 14, Views: 6120, Community: "chay-bo-moi-ngay",
		Content: "Sáng nay chạy 10km ở Tao Đàn, gặp một chú khoảng 60 tuổi chạy cùng tốc độ mình suốt 4km. Cuối buổi chú bảo: \"Chú chạy 22 năm rồi, mỗi ngày 5km, không nghỉ\".\n\nMình hỏi bí quyết, chú cười: \"Đừng nghĩ tới thành tích, nghĩ tới việc mai còn được chạy\".\n\nMình tập 6 năm, đổi giày 4 đôi, chạy 3 cái full marathon, mà câu đó vẫn là câu hay nhất mình nghe về chạy bộ.\n\n#ChayBo #Running #5AM",
		Media: []Media{Img(ImgMarathon1)},
		Reactions: []Reaction{
			{User: "phuongthao.mai", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "love"},
			{User: "hoangnam.dev", Emoji: "clap"}, {User: "tuankiet.dang", Emoji: "like"},
			{User: "bichha_to", Emoji: "clap"}, {User: "duclong.phan", Emoji: "like"},
			{User: "khanhlinh.do", Emoji: "heart"}, {User: "giahuy.ly", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "phuongthao.mai", AgeMin: 160, Text: "Câu này mình phải lưu lại. Chạy vì thành tích dễ bỏ lắm, chạy vì thấy vui thì bền."},
			{Author: "duclong.phan", AgeMin: 140, Text: "Chú 60 tuổi chạy 22 năm, tính ra hơn 40.000km. Kinh khủng."},
		},
	},
	{
		Key: "quocbao-quan", Author: "quocbao.pham", AgeHours: 16, Views: 2870,
		Title:   "Quán mình mở 5 năm, đây là 3 thứ mình đã làm sai",
		Content: "Nhân dịp quán tròn 5 tuổi, mình viết lại mấy sai lầm để ai đang tính mở quán thì tránh.\n\n1. Thuê mặt bằng quá rộng. 60m2 ở Hội An là quá nhiều cho 20 chỗ ngồi, tiền thuê ăn hết 30% doanh thu hai năm đầu.\n\n2. Menu 40 món. Không món nào làm thật tốt. Giờ còn 12 món, khách quay lại nhiều hơn.\n\n3. Không tính chi phí nguyên liệu hao hụt. Sữa, đá, cà phê rơi vãi mỗi ngày cộng lại một tháng bằng tiền lương một nhân viên.\n\nGiờ quán ổn rồi. Cảm ơn mọi người đã ghé. #QuanCaPhe #HoiAn",
		Media: []Media{Img(ImgCoffeeShop2), Img(ImgCaPhe2)},
		Reactions: []Reaction{
			{User: "giahuy.ly", Emoji: "like"}, {User: "minhanh.nguyen", Emoji: "clap"},
			{User: "thutrang_le", Emoji: "like"}, {User: "nhatminh.ho", Emoji: "love"},
			{User: "myduyen.duong", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "nhatminh.ho", AgeMin: 210, Text: "Quán anh view đẹp, mình chụp được mấy tấm ưng nhất ở đây. Hôm nào ghé lại."},
			{Author: "minhanh.nguyen", AgeMin: 190, Text: "Menu 12 món nghe hợp lý. Nhiều quán cứ nghĩ menu dài mới sang."},
		},
	},
	{
		Key: "khanhlinh-title-only", Author: "khanhlinh.do", AgeHours: 5.9, Views: 980,
		Title: "Ai ở Melbourne tháng 11 này không? Hẹn cà phê ở CBD",
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "like"}, {User: "haiyen.vu", Emoji: "like"},
			{User: "myduyen.duong", Emoji: "heart"},
		},
		Comments: []Comment{
			{Author: "haiyen.vu", AgeMin: 100, Text: "Ước gì em được đi 😭 Chị ở đó có nhớ đồ ăn Việt không?"},
			{Author: "khanhlinh.do", AgeMin: 90, Text: "Nhớ phở kinh khủng. Ở đây có quán Việt nhưng 18 đô một tô, ăn mà xót."},
		},
	},
	{
		Key: "ngochan-com-tam", Author: "ngochan_bui", AgeHours: 20, Views: 3350,
		Title:   "Cơm tấm sườn bì chả — bữa sáng của người Sài Gòn, mình nấu ở Cần Thơ",
		Content: "Sườn ướp qua đêm với sả, mật ong, nước mắm, tiêu. Bì trộn thính. Chả hấp. Nước mắm chua ngọt pha loãng hơn bình thường một chút.\n\nMình thấy cơm tấm ngon nhất là ăn sáng, tầm 7h, ngồi ở quán vỉa hè nhìn xe chạy. Ở nhà làm được 80% vị đó, còn 20% là không khí.\n\n#ComTam #AmThucVietNam",
		Media: []Media{Img(ImgComTam1), Img(ImgComTam2), Img(ImgComTam3)},
		Reactions: []Reaction{
			{User: "myduyen.duong", Emoji: "love"}, {User: "duclong.phan", Emoji: "love"},
			{User: "minhanh.nguyen", Emoji: "like"}, {User: "bichha_to", Emoji: "like"},
			{User: "tuankiet.dang", Emoji: "wow"},
		},
		Comments: []Comment{
			{Author: "duclong.phan", AgeMin: 300, Text: "Cơm tấm ăn sáng là chuẩn. Mình chạy xe cả ngày, sáng ăn dĩa cơm tấm là tới trưa không đói."},
			{Author: "myduyen.duong", AgeMin: 280, Text: "Cô làm bì bằng thịt heo luộc hay thịt nguội vậy cô?"},
		},
	},
	{
		Key: "minhanh-phoco-album", Author: "minhanh.nguyen", AgeHours: 24, Views: 4210,
		Title:   "Một chiều phố cổ, không kế hoạch gì cả",
		Content: "Chiều thứ bảy, mình không đặt lịch hẹn nào. Đi bộ từ Hàng Bạc ra hồ, ghé một quán cà phê trên gác hai nhìn xuống phố, ngồi tới lúc trời tối.\n\nPhố cổ đông, ồn, nhưng cứ tầm 5h chiều khi nắng xiên qua mái ngói là mình lại thấy nó đẹp theo một kiểu rất khó chụp.\n\n#PhoCo #HaNoi #ChieuThuBay",
		Media: []Media{Img(ImgPhoCo1), Img(ImgPhoCo2), Img(ImgPhoCo3), Img(ImgHoanKiem1), Img(ImgAodai)},
		Reactions: []Reaction{
			{User: "thutrang_le", Emoji: "love"}, {User: "haiyen.vu", Emoji: "love"},
			{User: "camtu.trinh", Emoji: "heart"}, {User: "khanhlinh.do", Emoji: "heart"},
			{User: "nhatminh.ho", Emoji: "like"}, {User: "quocbao.pham", Emoji: "like"},
			{User: "bichha_to", Emoji: "love"},
		},
		Comments: []Comment{
			{Author: "thutrang_le", AgeMin: 330, Text: "Ảnh thứ 3 là phố nào vậy em? Nhìn mấy ô cửa gỗ đẹp quá."},
			{Author: "minhanh.nguyen", AgeMin: 320, Text: "Hàng Bạc đó chị, em chụp lúc 5h20 nắng xiên."},
			{Author: "camtu.trinh", AgeMin: 300, Text: "Gái Hà Nội mặc áo dài chụp phố cổ là combo không bao giờ lỗi."},
		},
	},
	{
		Key: "nhatminh-media-only", Author: "nhatminh.ho", AgeHours: 30, Views: 1450,
		Media:     []Media{Img(ImgSapa3)},
		Reactions: []Reaction{{User: "thutrang_le", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "wow"}, {User: "khanhlinh.do", Emoji: "heart"}},
		Comments:  []Comment{{Author: "khanhlinh.do", AgeMin: 200, Text: "Ảnh này chụp lúc mấy giờ vậy anh? Ánh sáng đẹp quá."}},
	},
	{
		Key: "camtu-dog-video", Author: "camtu.trinh", AgeHours: 4.7, Views: 2380,
		Content:   "Cún nhà mình học được trò mới sau 3 ngày. Ăn hết một hộp hạt là làm được hết 😅\n\n#ThuCung #CunCon",
		Media:     []Media{Vid(VideoDog, 12, 1640000)},
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "haha"}, {User: "haiyen.vu", Emoji: "love"}, {User: "myduyen.duong", Emoji: "haha"}, {User: "ngochan_bui", Emoji: "love"}},
		Comments:  []Comment{{Author: "ngochan_bui", AgeMin: 240, Text: "Trộm vía khôn quá. Nhà cô có con mèo, dạy gì cũng giả vờ không nghe."}, {Author: "camtu.trinh", AgeMin: 210, Text: "@ngochan_bui cô dạy mèo giả vờ nghe thì con chịu rồi 😅 Cún nhà con thấy đồ ăn là quên hết."}},
	},
	{
		Key: "tuankiet-phongnha", Author: "tuankiet.dang", AgeHours: 40, Views: 2110,
		Title:   "Phong Nha: đi thuyền vào động, tối về ngủ homestay 200 nghìn",
		Content: "Mình đi Phong Nha 2 ngày, thuê xe máy từ Đồng Hới chạy vào 45km. Động Thiên Đường đi bộ nhiều nhưng xứng đáng, động Phong Nha thì phải đi thuyền, nước xanh đen, lạnh.\n\nTối ngủ homestay gần sông Son, 200 nghìn/đêm, có bữa sáng. Ăn tối ở quán ven đường, cá sông nướng và rau rừng.\n\n#PhongNha #QuangBinh #DuLichBui",
		Media: []Media{Img(ImgPhongNha1), Img(ImgPhongNha2)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "wow"}, {User: "thutrang_le", Emoji: "like"},
			{User: "nhatminh.ho", Emoji: "like"}, {User: "giahuy.ly", Emoji: "like"},
		},
		Comments: []Comment{
			{Author: "nhatminh.ho", AgeMin: 500, Text: "Phong Nha chụp sáng sớm sương trên sông Son đẹp lắm. Lần sau anh thử ở thêm một đêm."},
		},
	},
	{
		Key: "giahuy-trangan", Author: "giahuy.ly", AgeHours: 44, Views: 1890, Community: "du-lich-bui-viet-nam",
		Content: "Cuối tuần rồi cả team đi Tràng An. 8 người, 2 xe, đi từ 6h sáng, về tới Đà Nẵng 11h đêm. Mệt nhưng đáng.\n\nĐi thuyền 2 tiếng, nước trong, hai bên là núi đá vôi. Ngồi trên thuyền không ai nói gì, chỉ nghe tiếng mái chèo.\n\nĂn trưa ở nhà hàng gần bến, dê núi và cơm cháy. Tổng chi phí 700 nghìn/người.\n\n#TrangAn #NinhBinh #TeamBuilding",
		Media: []Media{Img(ImgTrangAn1), Img(ImgTrangAn2), Img(ImgTrangAn3)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "thanhtung.ngo", Emoji: "like"},
			{User: "tuankiet.dang", Emoji: "like"}, {User: "bichha_to", Emoji: "wow"},
			{User: "haiyen.vu", Emoji: "love"},
		},
		Comments: []Comment{
			{Author: "haiyen.vu", AgeMin: 620, Text: "Tràng An đi thuyền 2 tiếng mà không chán, mình đi hồi tháng 3 vẫn nhớ."},
		},
	},
	{
		Key: "myduyen-mientay", Author: "myduyen.duong", AgeHours: 48, Views: 2760,
		Title:   "Một ngày ở miền Tây: chợ nổi, bánh xèo, và cơn mưa chiều",
		Content: "5h sáng đã ra bến, thuê ghe 150 nghìn đi chợ nổi Cái Răng. Trên sông đã đông từ lúc trời còn tối.\n\nTrưa ăn bánh xèo ở quán cô Ba, 20 nghìn một cái to đùng. Chiều mưa, ngồi ở hiên nhà nhìn mưa rơi xuống sông.\n\nỞ miền Tây, mọi thứ chậm hơn, và mình thích vậy.\n\n#MienTay #CanTho #DuLich",
		Media: []Media{Img(ImgCanTho2), Img(ImgCanTho1), Img(ImgCanTho3), Img(ImgMotorbike3)},
		Reactions: []Reaction{
			{User: "ngochan_bui", Emoji: "love"}, {User: "thutrang_le", Emoji: "like"},
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "khanhlinh.do", Emoji: "wow"},
		},
		Comments: []Comment{
			{Author: "ngochan_bui", AgeMin: 700, Text: "Bánh xèo miền Tây to gấp đôi miền Trung, ăn một cái là no tới chiều."},
		},
	},
	{
		Key: "thutrang-ngan", Author: "thutrang_le", AgeHours: 52, Views: 860,
		Content:   "Sáng nay mở cửa sổ, gió lạnh đầu mùa. Đà Nẵng hiếm khi lạnh, nên cứ có gió là cả thành phố mặc áo khoác đi cà phê ☕",
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "love"}, {User: "haiyen.vu", Emoji: "like"}, {User: "bichha_to", Emoji: "like"}},
	},
	{
		Key: "bichha-hue", Author: "bichha_to", AgeHours: 56, Views: 2260,
		Title:   "Huế mùa này: đi bộ trong Đại Nội lúc 7h sáng",
		Content: "Bệnh viện cách Đại Nội 4km. Tan ca sớm, mình hay chạy xe vào đó rồi đi bộ một vòng.\n\nSáng sớm chưa có khách du lịch, chỉ có người tập thể dục và mấy cô bán vé. Nắng vàng rất nhẹ, tường gạch rêu.\n\nCó những ngày mình nghĩ, nếu không làm bác sĩ chắc mình đi làm hướng dẫn viên.\n\n#Hue #DaiNoi",
		Media: []Media{Img(ImgHue1), Img(ImgHue2)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "thutrang_le", Emoji: "heart"},
			{User: "nhatminh.ho", Emoji: "like"}, {User: "myduyen.duong", Emoji: "wow"},
		},
		Comments: []Comment{
			{Author: "thutrang_le", AgeMin: 800, Text: "Huế sáng sớm đúng là khác hẳn. Mình đi tháng 9, mưa mà vẫn đẹp."},
		},
	},
	{
		Key: "quocthinh-cafe", Author: "quocthinh.lam", AgeHours: 60, Views: 1240,
		Content:   "Vụ này cà phê chín muộn hơn mọi năm gần 3 tuần. Mưa nhiều quá. Nhà mình 2ha, năm nay chắc giảm 15% sản lượng.\n\n#CaPhe #DakLak",
		Media:     []Media{Img(ImgCoffeeFarm)},
		Reactions: []Reaction{{User: "duclong.phan", Emoji: "like"}, {User: "ngochan_bui", Emoji: "like"}},
	},
	{
		Key: "anhdung-ban-hang", Author: "anhdung.chu", AgeHours: 64, Views: 720,
		Title:   "Xả kho nồi chiên không dầu giá chỉ 299k — số lượng có hạn",
		Content: "Hàng chính hãng, bảo hành 12 tháng, ship toàn quốc. Inbox để nhận giá sỉ. Số lượng có hạn, ai nhanh thì được. Liên hệ ngay để nhận ưu đãi.",
		Reactions: []Reaction{{User: "duclong.phan", Emoji: "like"}},
	},
	{
		Key: "phuongthao-phaohoa", Author: "phuongthao.mai", AgeHours: 70, Views: 3540,
		Content: "Đứng ở bờ sông Sài Gòn xem pháo hoa, chen chúc mà vui. Năm nay mình nhìn đồng hồ đếm ngược với cả nghìn người xung quanh, tự nhiên thấy bình yên.\n\nMong năm nay mọi người đều có một năm nhẹ nhàng hơn. #PháoHoa #SaiGon #NamMoi",
		Media: []Media{Img(ImgFireworks1), Img(ImgFireworks2)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "thutrang_le", Emoji: "heart"},
			{User: "tuankiet.dang", Emoji: "like"}, {User: "camtu.trinh", Emoji: "love"},
			{User: "hoangnam.dev", Emoji: "like"},
		},
	},
	{
		Key: "minhanh-pinned", Author: "minhanh.nguyen", AgeHours: 72, Views: 1860, Pinned: true,
		Title:   "Mình là Minh Anh — marketing, Hà Nội, và đang học cách viết ít hơn",
		Content: "Bài ghim để ai mới vào profile thì biết mình là ai.\n\nMình làm marketing 5 năm, hiện ở một agency nhỏ ở Hà Nội. Viết nhiều cho khách, viết ít cho mình, đây là chỗ duy nhất mình viết linh tinh.\n\nThích: cà phê muối, chạy bộ ven hồ, phim của Lâm Hùng.\nKhông thích: họp mà không có agenda.\n\nRất vui được làm quen 👋",
		Media:     []Media{Img(ImgCaPhe1)},
		Reactions: []Reaction{{User: "thutrang_le", Emoji: "love"}, {User: "haiyen.vu", Emoji: "heart"}, {User: "camtu.trinh", Emoji: "love"}, {User: "myduyen.duong", Emoji: "like"}},
		Comments:  []Comment{{Author: "camtu.trinh", AgeMin: 900, Text: "Ghim bài giới thiệu là ý hay, mình cũng phải làm một bài."}},
	},
	{
		Key: "hoangnam-laptop", Author: "hoangnam.dev", AgeHours: 80, Views: 5240,
		Title:   "Chọn laptop cho người mới học lập trình: đừng mua cái đắt nhất",
		Content: "Mỗi tuần mình nhận 2-3 câu hỏi kiểu này nên viết luôn.\n\nRAM 16GB là mức tối thiểu nếu chạy Docker hoặc mở 20 tab. 8GB vẫn dùng được nhưng sẽ khó chịu sau 6 tháng.\n\nSSD 512GB. 256GB chật rất nhanh, và nâng cấp sau thường đắt hơn mua sẵn.\n\nĐừng quan tâm card đồ họa nếu không học game/AI. Đừng mua màn 4K trên máy 13 inch. Và đừng mua máy chỉ vì nó mỏng.\n\n#Laptop #LapTrinh #SinhVienIT",
		Media:     []Media{Img(ImgLaptop2)},
		Reactions: []Reaction{{User: "haiyen.vu", Emoji: "like"}, {User: "myduyen.duong", Emoji: "like"}, {User: "giahuy.ly", Emoji: "clap"}, {User: "tuankiet.dang", Emoji: "like"}},
		Comments:  []Comment{{Author: "haiyen.vu", AgeMin: 1000, Text: "Em đang dùng 8GB và đúng là mở Docker xong máy lag thật. Chắc phải lên 16GB."}},
	},
	{
		Key: "nhatminh-sapa", Author: "nhatminh.ho", AgeHours: 90, Views: 3120, Community: "du-lich-bui-viet-nam",
		Title:   "Sa Pa tháng 9: lúa chín vàng, nhưng đi sai ngày là mưa cả tuần",
		Content: "Mình ở Sa Pa 4 ngày, đi Tả Van, Lao Chải, và lên Fansipan bằng cáp treo vì trời mù.\n\nLúa chín đẹp nhất là khoảng 10-20/9. Mưa thì đi bộ trơn, nên mang giày bám tốt.\n\nỞ homestay Tả Van 250 nghìn/đêm, ăn tối cùng gia đình, có rượu ngô.\n\n#SaPa #LaoCai #RuongBacThang",
		Media: []Media{Img(ImgSapa1), Img(ImgSapa2), Img(ImgSapa3)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "wow"}, {User: "thutrang_le", Emoji: "love"},
			{User: "giahuy.ly", Emoji: "like"}, {User: "khanhlinh.do", Emoji: "heart"},
		},
	},
	{
		Key: "thanhtung-marathon", Author: "thanhtung.ngo", AgeHours: 100, Views: 4480,
		Title:   "Full marathon đầu tiên: 42km, 4h12, và một bài học về việc đi chậm",
		Content: "Mình xuất phát quá nhanh ở 10km đầu vì thấy ai cũng chạy. Đến km 30 thì đau đầu gối, phải đi bộ gần 2km.\n\nBài học: pace của mình, không phải pace của người bên cạnh. Chạy chậm mà về đích vẫn hơn bỏ cuộc ở km 35.\n\nGiờ mình đang tập cho cái thứ hai, mục tiêu dưới 4h. #Marathon #ChayBo",
		Media:     []Media{Img(ImgMarathon2)},
		Reactions: []Reaction{{User: "phuongthao.mai", Emoji: "clap"}, {User: "minhanh.nguyen", Emoji: "fire"}, {User: "hoangnam.dev", Emoji: "clap"}, {User: "bichha_to", Emoji: "like"}},
	},
	{
		Key: "haiyen-hoc", Author: "haiyen.vu", AgeHours: 120, Views: 1560,
		Content: "Mình vừa nhận kết quả học kỳ, GPA 3.4/4.0. Không phải cao nhất lớp nhưng là học kỳ mình học đều nhất.\n\nBa tháng trước mình lập một bảng: mỗi ngày 2 tiếng học, không điện thoại. Tuần đầu làm được 2 ngày. Tuần thứ tư làm được 5 ngày. Bây giờ thì thành thói quen.\n\nMình không tin vào động lực, mình tin vào việc sắp xếp lại cái bàn học.\n\n#SinhVien #HocTap",
		Reactions: []Reaction{{User: "thutrang_le", Emoji: "clap"}, {User: "minhanh.nguyen", Emoji: "like"}, {User: "khanhlinh.do", Emoji: "love"}, {User: "myduyen.duong", Emoji: "like"}},
		Comments:  []Comment{{Author: "thutrang_le", AgeMin: 1400, Text: "\"Mình không tin vào động lực, mình tin vào việc sắp xếp lại cái bàn học\" — cô sẽ dùng câu này với lớp."}},
	},
	{
		Key: "thutrang-sach", Author: "thutrang_le", AgeHours: 150, Views: 2380, Community: "review-phim-sach",
		Title:   "Ba cuốn sách mình đọc lại lần hai trong năm nay",
		Content: "1. Nhà giả kim — đọc lần đầu năm 18 tuổi thấy hay, đọc lại năm 30 thấy đúng.\n2. Tội ác và hình phạt — dày, nhưng đoạn đối thoại với Porfiry đáng cả cuốn sách.\n3. Một cuốn về kỹ năng viết của William Zinsser — mình dùng khi dạy Writing cho học sinh.\n\nĐọc lại một cuốn sách cũ giống như gặp lại một người bạn ở một giai đoạn khác của mình.\n\n#Sach #DocSach",
		Media:     []Media{Img(ImgBooks1)},
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "love"}, {User: "haiyen.vu", Emoji: "like"}, {User: "bichha_to", Emoji: "like"}},
		Comments:  []Comment{{Author: "haiyen.vu", AgeMin: 1600, Text: "Nhà giả kim em đọc hồi cấp 3, giờ đọc lại chắc hiểu khác hẳn."}, {Author: "bichha_to", AgeMin: 1200, Text: "Chị @thutrang_le ơi, cuốn số 3 em đang tìm, chị cho em xin tên đầy đủ với."}},
	},
	{
		Key: "giahuy-yoga", Author: "giahuy.ly", AgeHours: 170, Views: 980,
		Content:   "Sáng nào cũng 20 phút yoga trước khi mở laptop. Nghe thì sến nhưng nó là thứ duy nhất giữ mình không mở điện thoại ngay khi thức dậy.\n\n#Yoga #MorningRoutine",
		Media:     []Media{Img(ImgYoga)},
		Reactions: []Reaction{{User: "bichha_to", Emoji: "like"}, {User: "minhanh.nguyen", Emoji: "like"}},
	},
	{
		Key: "quocbao-hoian", Author: "quocbao.pham", AgeHours: 200, Views: 4120,
		Title:   "Hội An tháng này: phố đông trở lại, và mình mừng",
		Content: "Ba năm trước quán mình vắng tới mức một buổi chiều chỉ bán được 4 ly. Tháng này cuối tuần nào cũng kín chỗ.\n\nĐêm Hội An đẹp nhất là tầm 8h, khi đèn lồng lên hết và dòng người đi bộ dọc bờ sông. Mình đứng ở cửa quán nhìn ra, thấy thành phố của mình sống lại.\n\n#HoiAn #DuLich #QuangNam",
		Media: []Media{Img(ImgHoiAn1), Img(ImgHoiAn2), Img(ImgHoiAn3)},
		Reactions: []Reaction{
			{User: "minhanh.nguyen", Emoji: "love"}, {User: "thutrang_le", Emoji: "love"},
			{User: "nhatminh.ho", Emoji: "like"}, {User: "myduyen.duong", Emoji: "heart"},
			{User: "giahuy.ly", Emoji: "like"},
		},
	},
	{
		Key: "ngochan-me", Author: "ngochan_bui", AgeHours: 240, Views: 2860,
		Content: "Hôm nay đứa nhỏ hỏi: \"Mẹ ơi mẹ có ước gì không?\". Mình nói có, mẹ ước được ngủ một giấc tới trưa.\n\nNó suy nghĩ một lúc rồi bảo: \"Vậy mai con dậy sớm pha cà phê cho mẹ\". Sáng nay nó dậy thật, pha một ly cà phê sữa ngọt gắt, và làm đổ ra bàn.\n\nMình uống hết. #MeVaCon #GiaDinh",
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "love"}, {User: "thutrang_le", Emoji: "heart"}, {User: "myduyen.duong", Emoji: "love"}, {User: "bichha_to", Emoji: "love"}},
		Comments:  []Comment{{Author: "bichha_to", AgeMin: 2000, Text: "Ly cà phê ngọt gắt đó chắc là ly ngon nhất mẹ từng uống."}},
	},
	{
		Key: "myduyen-cafe", Author: "myduyen.duong", AgeHours: 300, Views: 720,
		Title:   "Quán cà phê nhỏ ở Cần Thơ, 25 nghìn một ly, ngồi cả buổi không ai đuổi",
		Content:   "Mình hay ngồi ở đây làm bài. Cà phê sữa đá pha phin, đá nhiều, ngọt vừa. Chủ quán là cô lớn tuổi, nhớ tên từng khách quen.\n\n#CaPhe #CanTho",
		Media:     []Media{Img(ImgCaPhe3)},
		Reactions: []Reaction{{User: "ngochan_bui", Emoji: "love"}, {User: "minhanh.nguyen", Emoji: "like"}},
	},
	{
		Key: "tuankiet-mua", Author: "tuankiet.dang", AgeHours: 340, Views: 640,
		Content:   "Hà Nội mưa cả ngày. Ngồi trong quán nhìn ra phố, thấy ai cũng đi nhanh hơn bình thường.\n\n#HaNoi #Mua",
		Media:     []Media{Img(ImgHanoiRain2)},
		Reactions: []Reaction{{User: "minhanh.nguyen", Emoji: "like"}, {User: "thutrang_le", Emoji: "like"}},
	},
	{
		Key: "minhanh-caphe", Author: "minhanh.nguyen", AgeHours: 400, Views: 1120,
		Content:   "Cà phê muối ở Hà Nội giờ quán nào cũng có, nhưng mình vẫn quay lại một quán nhỏ trên Hàng Bạc. Vì cô chủ nhớ mình uống ít đường.\n\n#CaPheMuoi #HaNoi",
		Media:     []Media{Img(ImgCaPhe2)},
		Reactions: []Reaction{{User: "haiyen.vu", Emoji: "love"}, {User: "camtu.trinh", Emoji: "like"}},
	},
	{
		Key: "hoangnam-pickleball", Author: "hoangnam.dev", AgeHours: 500, Views: 1340,
		Content: "Bị rủ đi pickleball 3 tuần trước, giờ tuần nào cũng chơi 2 buổi. Môn này dễ chơi hơn tennis mà vẫn ra mồ hôi.\n\nCó điều sân cuối tuần ở Sài Gòn khó đặt thật. #Pickleball #TheThao",
		Media:     []Media{Img(ImgPickleball)},
		Reactions: []Reaction{{User: "thanhtung.ngo", Emoji: "like"}, {User: "tuankiet.dang", Emoji: "like"}},
	},
	{
		Key: "anhdung-hidden", Author: "anhdung.chu", AgeHours: 26, Views: 340, Status: "hidden",
		Title:   "Bán tài khoản game giá rẻ — uy tín số 1",
		Content: "Nhận cày thuê, bán acc giá rẻ, bao check. Ai cần inbox. Cam kết uy tín, không lừa đảo.",
		Reactions: []Reaction{{User: "duclong.phan", Emoji: "angry"}},
	},
	{
		Key: "minhanh-private", Author: "minhanh.nguyen", AgeHours: 28, Views: 210, Status: "private",
		Content: "Ghi chú riêng: tháng sau phải xin nghỉ 3 ngày để về quê ăn giỗ. Nhớ đặt vé trước 2 tuần.",
	},
}

// PostsByKey tra cứu bài viết theo key.
func PostsByKey(key string) (Post, bool) {
	for _, p := range Posts {
		if p.Key == key {
			return p, true
		}
	}
	return Post{}, false
}

package seeddata

// AdPackage là một gói quảng cáo dựng sẵn cho đối tác.
type AdPackage struct {
	Key                  string
	Name                 string
	Description          string
	PriceMonthly         float64
	MaxSlots             int
	MaxDurationDays      int
	SupportsVideo        bool
	SupportsCarousel     bool
	HasAdvancedAnalytics bool
	SortOrder            int
}

// AdMedia là một tệp đính kèm của quảng cáo.
type AdMedia struct {
	URL   string
	Type  string // image | video
	Order int
}

// Ad là một chiến dịch quảng cáo seed.
// TargetLocations để trống nghĩa là chạy toàn quốc — bước seed ghi '[]'.
type Ad struct {
	Key             string
	Partner         string
	Package         string
	Title           string
	Content         string
	Format          string // image | video | carousel
	TargetURL       string
	Status          string // active | paused | completed | pending | rejected
	Budget          float64
	DailyBudget     float64
	TotalSpent      float64
	CPMPrice        float64
	CPCPrice        float64
	MaxImpressions  int
	TargetGender    string
	TargetAgeMin    int
	TargetAgeMax    int
	TargetLocations []string
	AgeDays         int // số ngày trước thời điểm seed
	DurationDays    int // số ngày chạy; 0 = không đặt hạn
	Media           []AdMedia
	// Impressions, Clicks dùng để sinh ad_analytics tổng hợp.
	Impressions int
	Clicks      int
}

// AdPackages là nguồn dữ liệu duy nhất cho gói quảng cáo.
var AdPackages = []AdPackage{
	{
		Key: "starter", Name: "Starter",
		Description:  "Dành cho cửa hàng nhỏ muốn thử quảng cáo. 1 chiến dịch, ảnh đơn, thống kê cơ bản.",
		PriceMonthly: 990000, MaxSlots: 1, MaxDurationDays: 15,
		SortOrder: 1,
	},
	{
		Key: "growth", Name: "Growth",
		Description:  "Cho quán cà phê, shop online đang mở rộng. 3 chiến dịch, hỗ trợ video và carousel.",
		PriceMonthly: 2490000, MaxSlots: 3, MaxDurationDays: 30,
		SupportsVideo: true, SupportsCarousel: true,
		SortOrder: 2,
	},
	{
		Key: "business", Name: "Business",
		Description:  "Cho thương hiệu cần báo cáo chi tiết. 8 chiến dịch, video, carousel và phân tích nâng cao.",
		PriceMonthly: 4990000, MaxSlots: 8, MaxDurationDays: 90,
		SupportsVideo: true, SupportsCarousel: true, HasAdvancedAnalytics: true,
		SortOrder: 3,
	},
}

// Ads là nguồn dữ liệu duy nhất cho chiến dịch quảng cáo seed.
// Tất cả thuộc tài khoản đối tác Phạm Quốc Bảo (quán Cà phê Nhà Gỗ, Hội An).
var Ads = []Ad{
	{
		Key: "nhagoc-ca-phe-sang", Partner: PartnerUser, Package: "growth",
		Title:       "Cà phê rang mộc Nhà Gỗ — 25k/ly, sân vườn Hội An",
		Content:     "Hạt robusta rang mộc mỗi sáng, pha phin, đá nhiều. Sân vườn 40 chỗ, mở từ 6h30. Giảm 20% cho khách ghé trước 8h.",
		Format:      "image",
		TargetURL:   "https://nhagoc.coffee/khung-gio-sang",
		Status:      "active",
		Budget:      6000000, DailyBudget: 250000, TotalSpent: 2840000,
		CPMPrice: 25000, CPCPrice: 1200, MaxImpressions: 240000,
		TargetGender: "all", TargetAgeMin: 0, TargetAgeMax: 100,
		AgeDays: 12, DurationDays: 30,
		Media: []AdMedia{
			{URL: ImgCoffeeShop2, Type: "image", Order: 0},
			{URL: ImgCaPhe2, Type: "image", Order: 1},
		},
		Impressions: 113600, Clicks: 3410,
	},
	{
		Key: "nhagoc-lop-rang", Partner: PartnerUser, Package: "business",
		Title:       "Lớp học rang cà phê cuối tuần tại Nhà Gỗ",
		Content:     "Một buổi sáng học cách chọn hạt, rang và pha phin. Lớp tối đa 8 người, có mang hạt về. Phù hợp người mới hoàn toàn.",
		Format:      "video",
		TargetURL:   "https://nhagoc.coffee/lop-rang",
		Status:      "active",
		Budget:      9000000, DailyBudget: 400000, TotalSpent: 4120000,
		CPMPrice: 30000, CPCPrice: 1500, MaxImpressions: 300000,
		TargetGender: "all", TargetAgeMin: 22, TargetAgeMax: 45,
		AgeDays: 8, DurationDays: 45,
		Media: []AdMedia{
			{URL: VideoSeaTurtle, Type: "video", Order: 0},
			{URL: ImgCoffeeFarm, Type: "image", Order: 1},
		},
		Impressions: 137300, Clicks: 4210,
	},
	{
		Key: "nhagoc-qua-tang", Partner: PartnerUser, Package: "growth",
		Title:       "Hộp quà cà phê Tết — gói tay, gửi toàn quốc",
		Content:     "Hộp 500g robusta rang mộc kèm phin nhôm và thiệp viết tay. Đặt trước 20/12 để kịp giao Tết. Miễn phí ship đơn từ 2 hộp.",
		Format:      "carousel",
		TargetURL:   "https://nhagoc.coffee/qua-tet",
		Status:      "active",
		Budget:      5000000, DailyBudget: 200000, TotalSpent: 970000,
		CPMPrice: 22000, CPCPrice: 1100, MaxImpressions: 200000,
		TargetGender: "all", TargetAgeMin: 25, TargetAgeMax: 55,
		AgeDays: 5, DurationDays: 40,
		Media: []AdMedia{
			{URL: ImgCaPhe1, Type: "image", Order: 0},
			{URL: ImgCaPhe3, Type: "image", Order: 1},
			{URL: ImgCoffeeShop1, Type: "image", Order: 2},
		},
		Impressions: 44100, Clicks: 1280,
	},
	{
		Key: "nhagoc-cuoi-tuan", Partner: PartnerUser, Package: "starter",
		Title:       "Cuối tuần này Nhà Gỗ có nhạc sống 19h",
		Content:     "Hai bạn trẻ chơi acoustic, không gian ngoài vườn, tối đa 50 khách. Đặt bàn trước để giữ chỗ gần sân khấu.",
		Format:      "image",
		TargetURL:   "https://nhagoc.coffee/nhac-song",
		Status:      "paused",
		Budget:      1500000, DailyBudget: 150000, TotalSpent: 1180000,
		CPMPrice: 20000, CPCPrice: 900, MaxImpressions: 60000,
		TargetGender: "all", TargetAgeMin: 18, TargetAgeMax: 40,
		AgeDays: 20, DurationDays: 14,
		Media: []AdMedia{
			{URL: ImgCoffeeShop3, Type: "image", Order: 0},
		},
		Impressions: 59000, Clicks: 1740,
	},
	{
		Key: "nhagoc-khai-truong", Partner: PartnerUser, Package: "business",
		Title:       "Nhà Gỗ mở chi nhánh thứ hai ở An Hội",
		Content:     "Chi nhánh mới ngay bờ sông, view đèn lồng. Khai trương 3 ngày đầu giảm 30% toàn menu.",
		Format:      "image",
		TargetURL:   "https://nhagoc.coffee/an-hoi",
		Status:      "completed",
		Budget:      7000000, DailyBudget: 300000, TotalSpent: 7000000,
		CPMPrice: 28000, CPCPrice: 1300, MaxImpressions: 250000,
		TargetGender: "all", TargetAgeMin: 0, TargetAgeMax: 100,
		AgeDays: 70, DurationDays: 30,
		Media: []AdMedia{
			{URL: ImgHoiAn1, Type: "image", Order: 0},
			{URL: ImgHoiAn3, Type: "image", Order: 1},
		},
		Impressions: 250000, Clicks: 8120,
	},
	{
		Key: "nhagoc-voucher", Partner: PartnerUser, Package: "starter",
		Title:       "Voucher 50k cho khách lần đầu",
		Content:     "Nhập mã NHAGOC50 khi đặt bàn online. Áp dụng cho hoá đơn từ 150k, không áp dụng ngày lễ.",
		Format:      "image",
		TargetURL:   "https://nhagoc.coffee/voucher",
		Status:      "pending",
		Budget:      2000000, DailyBudget: 150000, TotalSpent: 0,
		CPMPrice: 20000, CPCPrice: 1000, MaxImpressions: 80000,
		TargetGender: "all", TargetAgeMin: 18, TargetAgeMax: 50,
		AgeDays: 1, DurationDays: 21,
		Media: []AdMedia{
			{URL: ImgCaPhe2, Type: "image", Order: 0},
		},
	},
}

// AdByKey tra cứu chiến dịch quảng cáo theo key.
func AdByKey(key string) (Ad, bool) {
	for _, a := range Ads {
		if a.Key == key {
			return a, true
		}
	}
	return Ad{}, false
}

// AdPackageByKey tra cứu gói quảng cáo theo key.
func AdPackageByKey(key string) (AdPackage, bool) {
	for _, p := range AdPackages {
		if p.Key == key {
			return p, true
		}
	}
	return AdPackage{}, false
}

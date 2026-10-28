package seeddata

// Media là một tệp đính kèm trong seed. Type dùng đúng MIME mà server lưu
// (`image/jpeg`, `video/mp4`); Size tính bằng byte; Duration chỉ dùng cho video.
type Media struct {
	URI      string
	Type     string
	Size     float64
	Duration int
}

// Img tạo media ảnh JPEG từ URI đã kiểm duyệt.
func Img(uri string) Media {
	return Media{URI: uri, Type: "image/jpeg", Size: 268435}
}

// Vid tạo media video MP4.
func Vid(uri string, duration int, size float64) Media {
	return Media{URI: uri, Type: "video/mp4", Size: size, Duration: duration}
}

const (
	// Ảnh chân dung — Unsplash License (dùng thương mại, không cần ghi nguồn).
	AvatarSuperAdmin = "https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarAdmin      = "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarQuocBao    = "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarMinhAnh    = "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarHoangNam   = "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarThuTrang   = "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarHaiYen     = "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarTuanKiet   = "https://images.unsplash.com/photo-1527980965255-d3b416303d12?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarNgocHan    = "https://images.unsplash.com/photo-1531123897727-8f129e1688ce?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarNhatMinh   = "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarThanhTung  = "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarKhanhLinh  = "https://images.unsplash.com/photo-1517841905240-472988babdf9?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarPhuongThao = "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarGiaHuy     = "https://images.unsplash.com/photo-1560250097-0b93528c311a?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarCamTu      = "https://images.unsplash.com/photo-1524504388940-b1c1722653e1?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarDucLong    = "https://images.unsplash.com/photo-1552058544-f2b08422138a?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarBichHa     = "https://images.unsplash.com/photo-1573497019940-1c28c88b4f3e?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarAnhDung    = "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarMyDuyen    = "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=320&h=320&fit=crop&crop=faces&q=80"
	AvatarQuocThinh  = "https://images.unsplash.com/photo-1489424731084-a5d8b219a5bb?w=320&h=320&fit=crop&crop=faces&q=80"
)

const (
	// Ảnh bìa hồ sơ — ảnh phong cảnh thật của Việt Nam (Wikimedia Commons).
	CoverHoanKiem  = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f3/H%E1%BB%93_Ho%C3%A0n_Ki%E1%BA%BFm.jpg/1280px-H%E1%BB%93_Ho%C3%A0n_Ki%E1%BA%BFm.jpg"
	CoverHoiAn     = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/35/%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2828%29.jpg/1280px-%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2828%29.jpg"
	CoverDaLat     = "https://upload.wikimedia.org/wikipedia/commons/thumb/a/aa/Da_Lat_-_Xuan_Huong_Lake.jpg/1280px-Da_Lat_-_Xuan_Huong_Lake.jpg"
	CoverSapa      = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cd/Sa_Pa_Rice_Terrace_III.jpg/1280px-Sa_Pa_Rice_Terrace_III.jpg"
	CoverCanTho    = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c6/Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Sale.jpg/1280px-Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Sale.jpg"
	CoverSaiGon    = "https://upload.wikimedia.org/wikipedia/commons/thumb/0/08/Ho_Chi_Minh_City_Skyline_at_Night.jpg/1280px-Ho_Chi_Minh_City_Skyline_at_Night.jpg"
	CoverDaNang    = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/38/Marine_da_nang.jpg/1280px-Marine_da_nang.jpg"
	CoverHue       = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/12/Hue_Vietnam_Citadel-of-Hu%E1%BA%BF-21.jpg/1280px-Hue_Vietnam_Citadel-of-Hu%E1%BA%BF-21.jpg"
	CoverHaGiang   = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2f/Ha_Giang_rizieres_bis.jpg/1280px-Ha_Giang_rizieres_bis.jpg"
	CoverTrangAn   = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/68/Vietnam%2C_Ninh_Binh%2C_Trang_An_Area.jpg/1280px-Vietnam%2C_Ninh_Binh%2C_Trang_An_Area.jpg"
	CoverCoffee    = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/12/Vietnam_-_coffee_plantation.jpg/1280px-Vietnam_-_coffee_plantation.jpg"
	CoverPhoCo     = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/68/Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_-_NKS.jpg/1280px-Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_-_NKS.jpg"
	CoverPhongNha  = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/bc/Phong_Nha-Ke_Bang_cave3.jpg/1280px-Phong_Nha-Ke_Bang_cave3.jpg"
)

const (
	// Ẩm thực Việt.
	ImgPho1      = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3d/Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_02.jpg/960px-Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_02.jpg"
	ImgPho2      = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/42/Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_01.jpg/960px-Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_01.jpg"
	ImgPho3      = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/44/Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_03.jpg/960px-Ph%E1%BB%9F_b%C3%B2_in_a_German_restaurant_03.jpg"
	ImgBunCha1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f9/B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%C3%A1ng_7_n%C4%83m_2018%29_%281%29.jpg/960px-B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%C3%A1ng_7_n%C4%83m_2018%29_%281%29.jpg"
	ImgBunCha2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/da/B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%C3%A1ng_7_n%C4%83m_2018%29_%282%29.jpg/960px-B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%C3%A1ng_7_n%C4%83m_2018%29_%282%29.jpg"
	ImgBunCha3   = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/87/B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng%29_t%E1%BA%A1i_Nguy%E1%BB%85n_S%C6%A1n%2C_T%C3%A2n_Ph%C3%BA_%281%29.jpg/960px-B%C3%BAn_ch%E1%BA%A3_H%C3%A0_N%E1%BB%99i_%28th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng%29_t%E1%BA%A1i_Nguy%E1%BB%85n_S%C6%A1n%2C_T%C3%A2n_Ph%C3%BA_%281%29.jpg"
	ImgComTam1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/C%C6%A1m_T%E1%BA%A5m%2C_Da_Nang%2C_Vietnam.jpg/960px-C%C6%A1m_T%E1%BA%A5m%2C_Da_Nang%2C_Vietnam.jpg"
	ImgComTam2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/8e/Com-Tam-2008.jpg/960px-Com-Tam-2008.jpg"
	ImgComTam3   = "https://upload.wikimedia.org/wikipedia/commons/thumb/a/ab/C%C6%A1m_t%E1%BA%A5m_SG_%28c%C6%A1m_s%C6%B0%E1%BB%9Dn%29_ng27th9n2020_%281%29.jpg/960px-C%C6%A1m_t%E1%BA%A5m_SG_%28c%C6%A1m_s%C6%B0%E1%BB%9Dn%29_ng27th9n2020_%281%29.jpg"
	ImgBanhMi1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a8/B%C3%A1nh_m%C3%AC_than_tre_Vi%E1%BB%87t_Nam_20201205.jpg/960px-B%C3%A1nh_m%C3%AC_than_tre_Vi%E1%BB%87t_Nam_20201205.jpg"
	ImgBanhMi2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e5/B%C3%A1nh_m%C3%AC_k%E1%BA%B9p_b%C3%A1nh_l%E1%BB%8Dc_%E1%BB%9F_%C4%90%C3%B4ng_H%C3%A0_n%C4%83m_2018.jpg/960px-B%C3%A1nh_m%C3%AC_k%E1%BA%B9p_b%C3%A1nh_l%E1%BB%8Dc_%E1%BB%9F_%C4%90%C3%B4ng_H%C3%A0_n%C4%83m_2018.jpg"
	ImgBanhMi3   = "https://upload.wikimedia.org/wikipedia/commons/thumb/0/0e/Cu%E1%BB%99c_thi_n%E1%BA%A5u_n%C6%B0%E1%BB%9Bng_%E1%BB%9F_Vi%E1%BB%87t_Nam_n%C4%83m_2010-B%C3%A1nh_m%C3%AC_Vi%E1%BB%87t_Nam_c%E1%BA%AFt_mi%E1%BA%BFng_%286%29.jpg/960px-Cu%E1%BB%99c_thi_n%E1%BA%A5u_n%C6%B0%E1%BB%9Bng_%E1%BB%9F_Vi%E1%BB%87t_Nam_n%C4%83m_2010-B%C3%A1nh_m%C3%AC_Vi%E1%BB%87t_Nam_c%E1%BA%AFt_mi%E1%BA%BFng_%286%29.jpg"
	ImgCaPhe1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/e/ec/Ca_Phe_Sua_Da.jpg/960px-Ca_Phe_Sua_Da.jpg"
	ImgCaPhe2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/fc/C%C3%80_PH%C3%8A_S%E1%BB%AEA_PHA_PHIN.jpg/960px-C%C3%80_PH%C3%8A_S%E1%BB%AEA_PHA_PHIN.jpg"
	ImgCaPhe3    = "https://upload.wikimedia.org/wikipedia/commons/2/28/C%C3%A0_ph%C3%AA_s%E1%BB%AFa_%C4%91%C3%A1_preparation.jpg"
	ImgCoffeeFarm = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/12/Vietnam_-_coffee_plantation.jpg/1280px-Vietnam_-_coffee_plantation.jpg"
)

const (
	// Hà Nội.
	ImgHoanKiem1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/7/70/Hanoi_Vietnam_Shrine-at-Ho%C3%A0n-Ki%E1%BA%BFm-Lake-01.jpg/960px-Hanoi_Vietnam_Shrine-at-Ho%C3%A0n-Ki%E1%BA%BFm-Lake-01.jpg"
	ImgHoanKiem2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f3/H%E1%BB%93_Ho%C3%A0n_Ki%E1%BA%BFm.jpg/960px-H%E1%BB%93_Ho%C3%A0n_Ki%E1%BA%BFm.jpg"
	ImgHoanKiem3   = "https://upload.wikimedia.org/wikipedia/commons/1/1c/Hoan_Kiem.jpg"
	ImgPhoCo1      = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/68/Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_-_NKS.jpg/960px-Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_-_NKS.jpg"
	ImgPhoCo2      = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d0/%C4%90i%C3%AAu_kh%E1%BA%AFc_h%E1%BB%95_tr%C3%AAn_ng%C3%B4i_%C4%91%C3%ACnh_%E1%BB%9F_Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_n%C4%83m_2017_%288%29.jpg/960px-%C4%90i%C3%AAu_kh%E1%BA%AFc_h%E1%BB%95_tr%C3%AAn_ng%C3%B4i_%C4%91%C3%ACnh_%E1%BB%9F_Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_n%C4%83m_2017_%288%29.jpg"
	ImgPhoCo3      = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/%C4%90i%C3%AAu_kh%E1%BA%AFc_h%E1%BB%95_tr%C3%AAn_ng%C3%B4i_%C4%91%C3%ACnh_%E1%BB%9F_Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_n%C4%83m_2017_%287%29.jpg/960px-%C4%90i%C3%AAu_kh%E1%BA%AFc_h%E1%BB%95_tr%C3%AAn_ng%C3%B4i_%C4%91%C3%ACnh_%E1%BB%9F_Ph%E1%BB%91_c%E1%BB%95_H%C3%A0_N%E1%BB%99i_n%C4%83m_2017_%287%29.jpg"
	ImgHanoiTraffic1 = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f9/Hanoi_traffic_sentinel_Flickr.jpg/960px-Hanoi_traffic_sentinel_Flickr.jpg"
	ImgHanoiTraffic2 = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cc/Hanoi_Traffic.jpg/960px-Hanoi_Traffic.jpg"
	ImgHanoiRain1  = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/f1/Hanoi_rain.JPG/960px-Hanoi_rain.JPG"
	ImgHanoiRain2  = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/46/Streets_of_Hanoi_-_Hanoi1615.jpg/960px-Streets_of_Hanoi_-_Hanoi1615.jpg"
	ImgOnePillar   = "https://upload.wikimedia.org/wikipedia/commons/thumb/f/fd/One_Pillar_Pagoda%2C_Hanoi_%2831083369920%29.jpg/960px-One_Pillar_Pagoda%2C_Hanoi_%2831083369920%29.jpg"
)

const (
	// Miền Trung & Tây Nguyên.
	ImgDaNang1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/38/Marine_da_nang.jpg/1280px-Marine_da_nang.jpg"
	ImgDaNang2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/12/Da-Nang_Vietnam_Coracles-01.jpg/960px-Da-Nang_Vietnam_Coracles-01.jpg"
	ImgHoiAn1     = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/60/%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%285%29.jpg/960px-%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%285%29.jpg"
	ImgHoiAn2     = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/35/%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2828%29.jpg/960px-%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2828%29.jpg"
	ImgHoiAn3     = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/1c/%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2820%29.jpg/960px-%C4%90%C3%A8n_l%E1%BB%93ng_H%E1%BB%99i_An_lung_linh%2C_th%C3%A1ng_3_n%C4%83m_2021_%2820%29.jpg"
	ImgHue1       = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/12/Hue_Vietnam_Citadel-of-Hu%E1%BA%BF-21.jpg/960px-Hue_Vietnam_Citadel-of-Hu%E1%BA%BF-21.jpg"
	ImgHue2       = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/88/Hue_Vietnam_Thien-Mu-Temple-and-Pagoda-03.jpg/960px-Hue_Vietnam_Thien-Mu-Temple-and-Pagoda-03.jpg"
	ImgPhongNha1  = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/bc/Phong_Nha-Ke_Bang_cave3.jpg/960px-Phong_Nha-Ke_Bang_cave3.jpg"
	ImgPhongNha2  = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4c/Dong_Hoi_Phong_Nha_grotten_%2868776%29.jpg/960px-Dong_Hoi_Phong_Nha_grotten_%2868776%29.jpg"
	ImgPhongNha3  = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/1f/Dong_Hoi_Phong_Nha_grotten_%2859500%29.jpg/960px-Dong_Hoi_Phong_Nha_grotten_%2859500%29.jpg"
	ImgDaLat1     = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/67/Da_Lat_train_station_02.JPG/960px-Da_Lat_train_station_02.JPG"
	ImgDaLat2     = "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Da_Lat_-_Viet_Nam.jpg/960px-Da_Lat_-_Viet_Nam.jpg"
	ImgDaLat3     = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/66/Cathedral_of_Da_Lat_05.jpg/960px-Cathedral_of_Da_Lat_05.jpg"
	ImgDaLat4     = "https://upload.wikimedia.org/wikipedia/commons/thumb/a/aa/Da_Lat_-_Xuan_Huong_Lake.jpg/960px-Da_Lat_-_Xuan_Huong_Lake.jpg"
	ImgSapa1      = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2c/Road_and_paddy_fields_in_Sa_Pa%2C_Vietnam%2C_20240126_1203_3590.jpg/960px-Road_and_paddy_fields_in_Sa_Pa%2C_Vietnam%2C_20240126_1203_3590.jpg"
	ImgSapa2      = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/38/Road_and_paddy_fields_in_Sa_Pa%2C_Vietnam%2C_20240126_1202_3586.jpg/960px-Road_and_paddy_fields_in_Sa_Pa%2C_Vietnam%2C_20240126_1202_3586.jpg"
	ImgSapa3      = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cd/Sa_Pa_Rice_Terrace_III.jpg/960px-Sa_Pa_Rice_Terrace_III.jpg"
	ImgHaGiang1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/37/Ru%E1%BB%99ng_b%E1%BA%ADc_thang_%E1%BB%9F_Ho%C3%A0ng_Su_Ph%C3%AC.jpg/960px-Ru%E1%BB%99ng_b%E1%BA%ADc_thang_%E1%BB%9F_Ho%C3%A0ng_Su_Ph%C3%AC.jpg"
	ImgHaGiang2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2f/Ha_Giang_rizieres_bis.jpg/960px-Ha_Giang_rizieres_bis.jpg"
	ImgTrangAn1   = "https://upload.wikimedia.org/wikipedia/commons/thumb/0/0b/Trang_An_Landscape_Complex%2C_Ninh_Binh_Province%2C_Vietnam%2C_20240202_1456_5313.jpg/960px-Trang_An_Landscape_Complex%2C_Ninh_Binh_Province%2C_Vietnam%2C_20240202_1456_5313.jpg"
	ImgTrangAn2   = "https://upload.wikimedia.org/wikipedia/commons/thumb/9/9c/Trang_An_Landscape_Complex%2C_Ninh_Binh_Province%2C_Vietnam%2C_20240202_1447_5310.jpg/960px-Trang_An_Landscape_Complex%2C_Ninh_Binh_Province%2C_Vietnam%2C_20240202_1447_5310.jpg"
	ImgTrangAn3   = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/68/Vietnam%2C_Ninh_Binh%2C_Trang_An_Area.jpg/960px-Vietnam%2C_Ninh_Binh%2C_Trang_An_Area.jpg"
	ImgAodai      = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cb/Vietnamese_girl_wearing_ao_dai_3.jpg/960px-Vietnamese_girl_wearing_ao_dai_3.jpg"
	ImgAodai2     = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3b/%C3%81o_d%C3%A0i_hoa_ph%C6%B0%E1%BB%A3ng_-_Ph%C3%A1t_Tr%C6%B0%C6%A1ng_-_51231007393_%28extend_cropped%29.jpg/960px-%C3%81o_d%C3%A0i_hoa_ph%C6%B0%E1%BB%A3ng_-_Ph%C3%A1t_Tr%C6%B0%C6%A1ng_-_51231007393_%28extend_cropped%29.jpg"
)

const (
	// Miền Nam & đời sống.
	ImgCanTho1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/2b/Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Rawing.jpg/960px-Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Rawing.jpg"
	ImgCanTho2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/26/Vietnam%2C_Phong_Dien%2C_Floating_market%2C_Boats.jpg/960px-Vietnam%2C_Phong_Dien%2C_Floating_market%2C_Boats.jpg"
	ImgCanTho3    = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c6/Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Sale.jpg/960px-Can_Tho%2C_Vietnam%2C_Floating_Market%2C_Sale.jpg"
	ImgSaiGon1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/0/08/Ho_Chi_Minh_City_Skyline_at_Night.jpg/1280px-Ho_Chi_Minh_City_Skyline_at_Night.jpg"
	ImgSaiGon2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b1/R%C3%ADo_Saig%C3%B3n%2C_Ciudad_Ho_Chi_Minh%2C_Vietnam%2C_2013-08-14%2C_DD_29.JPG/960px-R%C3%ADo_Saig%C3%B3n%2C_Ciudad_Ho_Chi_Minh%2C_Vietnam%2C_2013-08-14%2C_DD_29.JPG"
	ImgMotorbike1 = "https://upload.wikimedia.org/wikipedia/commons/thumb/9/90/Ho_Chi_Minh_City%2C_Vietnam%2C_Motorbike_riders.jpg/960px-Ho_Chi_Minh_City%2C_Vietnam%2C_Motorbike_riders.jpg"
	ImgMotorbike2 = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/62/Man_on_Motorbike_-_Lang_Son_-_Vietnam_%2848142923111%29.jpg/960px-Man_on_Motorbike_-_Lang_Son_-_Vietnam_%2848142923111%29.jpg"
	ImgMotorbike3 = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/86/Motorbike_river_ferry_Vietnam.jpg/960px-Motorbike_river_ferry_Vietnam.jpg"
	ImgCat1       = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4d/Cat_November_2010-1a.jpg/960px-Cat_November_2010-1a.jpg"
	ImgCat2       = "https://upload.wikimedia.org/wikipedia/commons/thumb/7/72/Cat_playing_with_a_lizard.jpg/960px-Cat_playing_with_a_lizard.jpg"
	ImgMarathon1  = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/16/Berlin-Marathon_2015_Runners_14.jpg/960px-Berlin-Marathon_2015_Runners_14.jpg"
	ImgMarathon2  = "https://upload.wikimedia.org/wikipedia/commons/thumb/4/49/Berlin-Marathon_2015_Runners_33.jpg/960px-Berlin-Marathon_2015_Runners_33.jpg"
	ImgPickleball = "https://upload.wikimedia.org/wikipedia/commons/thumb/5/5d/Pickleball_Paddle_2025.jpg/960px-Pickleball_Paddle_2025.jpg"
	ImgConcert    = "https://upload.wikimedia.org/wikipedia/commons/thumb/7/7e/Crowd_at_Noam_Rotem_concert.jpg/960px-Crowd_at_Noam_Rotem_concert.jpg"
	ImgLaptop1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/3/37/Schenker_VIA14_Laptop_asv2021-01.jpg/960px-Schenker_VIA14_Laptop_asv2021-01.jpg"
	ImgLaptop2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/8/88/Notebook_computer.jpg/960px-Notebook_computer.jpg"
	ImgBooks1     = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c5/Part_of_a_bookshelf_containing_books_by_Aristotle_%281%29.jpg/960px-Part_of_a_bookshelf_containing_books_by_Aristotle_%281%29.jpg"
	ImgBooks2     = "https://upload.wikimedia.org/wikipedia/commons/thumb/7/73/Part_of_a_bookshelf_containing_books_by_ancient_philosophers_%281.1%29.jpg/960px-Part_of_a_bookshelf_containing_books_by_ancient_philosophers_%281.1%29.jpg"
	ImgCoffeeShop1 = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b0/Starbucks_coffee_shop_interior_design.jpg/960px-Starbucks_coffee_shop_interior_design.jpg"
	ImgCoffeeShop2 = "https://upload.wikimedia.org/wikipedia/commons/thumb/1/1b/Interior_Johnie%27s_Coffee_Shop_2021.jpg/960px-Interior_Johnie%27s_Coffee_Shop_2021.jpg"
	ImgCoffeeShop3 = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/c8/Interior_Johnie%27s_Coffee_Shop_2021_2.jpg/960px-Interior_Johnie%27s_Coffee_Shop_2021_2.jpg"
	ImgFireworks1 = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b4/Fireworks_on_New_Year%27s_Eve_in_a_small_Swabian_village_%281%29%2C_brightened.jpg/960px-Fireworks_on_New_Year%27s_Eve_in_a_small_Swabian_village_%281%29%2C_brightened.jpg"
	ImgFireworks2 = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/6b/%D0%A4%D0%B5%D0%B9%D0%B5%D1%80%D0%B2%D0%B5%D1%80%D0%BA_%D0%B2_%D0%BC%D0%BE%D1%80%D1%81%D0%BA%D0%BE%D0%BC_%D0%BF%D0%BE%D1%80%D1%82%D1%83_%D0%A1%D0%BE%D1%87%D0%B8_9_%D0%BC%D0%B0%D1%8F_2015_%D0%B3%D0%BE%D0%B4%D0%B0.jpg/960px-%D0%A4%D0%B5%D0%B9%D0%B5%D1%80%D0%B2%D0%B5%D1%80%D0%BA_%D0%B2_%D0%BC%D0%BE%D1%80%D1%81%D0%BA%D0%BE%D0%BC_%D0%BF%D0%BE%D1%80%D1%82%D1%83_%D0%A1%D0%BE%D1%87%D0%B8_9_%D0%BC%D0%B0%D1%8F_2015_%D0%B3%D0%BE%D0%B4%D0%B0.jpg"
	ImgYoga       = "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a8/Yoga_%28IMO_9790921%29_Liverpool_2019.jpg/960px-Yoga_%28IMO_9790921%29_Liverpool_2019.jpg"
	ImgMakeup1    = "https://upload.wikimedia.org/wikipedia/commons/thumb/6/67/Makeup_weapons_Brushes_.jpg/960px-Makeup_weapons_Brushes_.jpg"
	ImgMakeup2    = "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b4/Blush_-_A_selection_of_makeup_brushes.jpg/960px-Blush_-_A_selection_of_makeup_brushes.jpg"
	ImgHospital   = "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cd/Hospital_corridor_of_Sir_Ganga_Ram_Hospital.jpg/960px-Hospital_corridor_of_Sir_Ganga_Ram_Hospital.jpg"
	ImgStudents1  = "https://upload.wikimedia.org/wikipedia/commons/thumb/0/05/Photographer_taking_a_group_photograph_of_smiling_students_in_front_of_the_Tokyo_station%2C_Marunouchi%2C_Japan.jpg/960px-Photographer_taking_a_group_photograph_of_smiling_students_in_front_of_the_Tokyo_station%2C_Marunouchi%2C_Japan.jpg"
	ImgStudents2  = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/d9/Chittagong_University_Central_Students%27_Union_%2813%29.jpg/960px-Chittagong_University_Central_Students%27_Union_%2813%29.jpg"
	ImgBridge     = "https://upload.wikimedia.org/wikipedia/commons/thumb/2/23/Dhubri-Phulbari_early_bridge_construction_%28Dhubri_end%29.jpg/960px-Dhubri-Phulbari_early_bridge_construction_%28Dhubri_end%29.jpg"
)

const (
	// Video MP4 host công khai (Cloudinary demo) — đã kiểm tra trả 206 video/mp4.
	VideoSeaTurtle = "https://res.cloudinary.com/demo/video/upload/sea_turtle.mp4"
	VideoDog       = "https://res.cloudinary.com/demo/video/upload/dog.mp4"
	VideoElephants = "https://res.cloudinary.com/demo/video/upload/elephants.mp4"
)

// MediaUpload là một tệp nằm trong hàng chờ kiểm duyệt (chưa gắn vào bài nào).
// Nhờ đó trang kiểm duyệt media của admin có dữ liệu thật để xử lý.
type MediaUpload struct {
	User     string
	URI      string
	Type     string
	Status   string // pending | flagged | rejected
	Reason   string
	AgeHours int
}

// UploadQueue là các tệp đang chờ duyệt / bị đánh dấu / bị từ chối.
var UploadQueue = []MediaUpload{
	{User: "minhanh.nguyen", URI: ImgMotorbike1, Type: "image/jpeg", Status: "pending", AgeHours: 3},
	{User: "hoangnam.dev", URI: ImgCat1, Type: "image/jpeg", Status: "pending", AgeHours: 6},
	{User: "camtu.trinh", URI: ImgStudents2, Type: "image/jpeg", Status: "flagged",
		Reason: "Bị báo cáo: ảnh có thể không phải của người đăng", AgeHours: 14},
	{User: "quocthinh.lam", URI: ImgBanhMi1, Type: "image/jpeg", Status: "rejected",
		Reason: "Ảnh mờ, không đạt yêu cầu chất lượng", AgeHours: 30},
}

// AllMediaURIs gom toàn bộ URI ảnh/video dùng trong seed để HEAD-check
// trước khi chụp ảnh báo cáo (cờ -check-media).
func AllMediaURIs() []string {
	return []string{
		AvatarSuperAdmin, AvatarAdmin, AvatarQuocBao, AvatarMinhAnh, AvatarHoangNam,
		AvatarThuTrang, AvatarHaiYen, AvatarTuanKiet, AvatarNgocHan, AvatarNhatMinh,
		AvatarThanhTung, AvatarKhanhLinh, AvatarPhuongThao, AvatarGiaHuy, AvatarCamTu,
		AvatarDucLong, AvatarBichHa, AvatarAnhDung, AvatarMyDuyen, AvatarQuocThinh,
		CoverHoanKiem, CoverHoiAn, CoverDaLat, CoverSapa, CoverCanTho, CoverSaiGon,
		CoverDaNang, CoverHue, CoverHaGiang, CoverTrangAn, CoverCoffee, CoverPhoCo, CoverPhongNha,
		ImgPho1, ImgPho2, ImgPho3, ImgBunCha1, ImgBunCha2, ImgBunCha3,
		ImgComTam1, ImgComTam2, ImgComTam3, ImgBanhMi1, ImgBanhMi2, ImgBanhMi3,
		ImgCaPhe1, ImgCaPhe2, ImgCaPhe3, ImgCoffeeFarm,
		ImgHoanKiem1, ImgHoanKiem2, ImgHoanKiem3, ImgPhoCo1, ImgPhoCo2, ImgPhoCo3,
		ImgHanoiTraffic1, ImgHanoiTraffic2, ImgHanoiRain1, ImgHanoiRain2, ImgOnePillar,
		ImgDaNang1, ImgDaNang2, ImgHoiAn1, ImgHoiAn2, ImgHoiAn3, ImgHue1, ImgHue2,
		ImgPhongNha1, ImgPhongNha2, ImgPhongNha3, ImgDaLat1, ImgDaLat2, ImgDaLat3, ImgDaLat4,
		ImgSapa1, ImgSapa2, ImgSapa3, ImgHaGiang1, ImgHaGiang2,
		ImgTrangAn1, ImgTrangAn2, ImgTrangAn3, ImgAodai, ImgAodai2,
		ImgCanTho1, ImgCanTho2, ImgCanTho3, ImgSaiGon1, ImgSaiGon2,
		ImgMotorbike1, ImgMotorbike2, ImgMotorbike3, ImgCat1, ImgCat2,
		ImgMarathon1, ImgMarathon2, ImgPickleball, ImgConcert, ImgLaptop1, ImgLaptop2,
		ImgBooks1, ImgBooks2, ImgCoffeeShop1, ImgCoffeeShop2, ImgCoffeeShop3,
		ImgFireworks1, ImgFireworks2, ImgYoga, ImgMakeup1, ImgMakeup2,
		ImgHospital, ImgStudents1, ImgStudents2, ImgBridge,
		VideoSeaTurtle, VideoDog, VideoElephants,
	}
}

package seeddata

import (
	"fmt"
	"strings"
)

// reactionCodes là 10 emoji cảm xúc legacy dùng cho post_reactions.
var reactionCodes = map[string]bool{
	"like": true, "love": true, "haha": true, "wow": true, "sad": true,
	"angry": true, "clap": true, "fire": true, "heart": true, "rocket": true,
}

// Validate kiểm tra chéo toàn bộ seeddata trước khi ghi vào DB: mọi tham chiếu
// theo key (user, bài viết, cộng đồng, quy tắc vi phạm) phải tồn tại. Nhờ vậy
// lỗi dữ liệu lộ ra ngay thay vì đổ vỡ giữa chuỗi seed.
func Validate() error {
	// ── Users ──
	users := make(map[string]bool, len(Users))
	usernames := make(map[string]bool, len(Users))
	emails := make(map[string]bool, len(Users))

	requireUser := func(where, key string) error {
		if key == "" {
			return fmt.Errorf("seeddata: %s thiếu user", where)
		}
		if !users[key] {
			return fmt.Errorf("seeddata: %s tham chiếu user %q không tồn tại", where, key)
		}
		return nil
	}

	for _, u := range Users {
		if u.Key == "" || u.Username == "" || u.Email == "" {
			return fmt.Errorf("seeddata: user %q thiếu key/username/email", u.Key)
		}
		if users[u.Key] {
			return fmt.Errorf("seeddata: user key trùng %q", u.Key)
		}
		if usernames[u.Username] {
			return fmt.Errorf("seeddata: username trùng %q", u.Username)
		}
		if emails[u.Email] {
			return fmt.Errorf("seeddata: email trùng %q", u.Email)
		}
		if !strings.HasPrefix(u.Email, u.Username+"@") {
			return fmt.Errorf("seeddata: email %q không khớp username %q", u.Email, u.Username)
		}
		users[u.Key] = true
		usernames[u.Username] = true
		emails[u.Email] = true
	}

	if !users[DemoUser] {
		return fmt.Errorf("seeddata: DemoUser %q không có trong Users", DemoUser)
	}
	if !users[PartnerUser] {
		return fmt.Errorf("seeddata: PartnerUser %q không có trong Users", PartnerUser)
	}

	// ── Communities ──
	communities := make(map[string]bool, len(Communities))
	for _, c := range Communities {
		if c.Key == "" {
			return fmt.Errorf("seeddata: cộng đồng %q thiếu key", c.Name)
		}
		if communities[c.Key] {
			return fmt.Errorf("seeddata: community key trùng %q", c.Key)
		}
		communities[c.Key] = true

		if err := requireUser("cộng đồng "+c.Key+" creator", c.Creator); err != nil {
			return err
		}
		if len(c.Members) == 0 {
			return fmt.Errorf("seeddata: cộng đồng %q không có thành viên", c.Key)
		}
		members := make(map[string]bool, len(c.Members))
		for _, m := range c.Members {
			if err := requireUser("cộng đồng "+c.Key+" member", m.User); err != nil {
				return err
			}
			if members[m.User] {
				return fmt.Errorf("seeddata: cộng đồng %q trùng thành viên %q", c.Key, m.User)
			}
			members[m.User] = true
		}
		if !members[c.Creator] {
			return fmt.Errorf("seeddata: cộng đồng %q: creator %q không nằm trong Members", c.Key, c.Creator)
		}
		if c.ChatName == "" && c.CreateChat {
			return fmt.Errorf("seeddata: cộng đồng %q bật chat nhưng thiếu ChatName", c.Key)
		}
		if c.Policy.ModThreshold <= c.Policy.TopThreshold {
			return fmt.Errorf("seeddata: cộng đồng %q có ngưỡng Moderator không lớn hơn Top Contributor", c.Key)
		}
	}

	for _, r := range JoinRequests {
		if !communities[r.Community] {
			return fmt.Errorf("seeddata: yêu cầu tham gia trỏ tới cộng đồng lạ %q", r.Community)
		}
		if err := requireUser("yêu cầu tham gia", r.User); err != nil {
			return err
		}
	}

	for _, c := range Challenges {
		if !communities[c.Community] {
			return fmt.Errorf("seeddata: thử thách %q trỏ tới cộng đồng lạ %q", c.Title, c.Community)
		}
		if c.StartHours >= c.EndHours {
			return fmt.Errorf("seeddata: thử thách %q có thời gian không hợp lệ", c.Title)
		}
	}

	// ── Posts ──
	posts := make(map[string]bool, len(Posts))
	for _, p := range Posts {
		if p.Key == "" {
			return fmt.Errorf("seeddata: bài viết của %q thiếu key", p.Author)
		}
		if posts[p.Key] {
			return fmt.Errorf("seeddata: post key trùng %q", p.Key)
		}
		posts[p.Key] = true

		if err := requireUser("bài viết "+p.Key, p.Author); err != nil {
			return err
		}
		if p.Community != "" && !communities[p.Community] {
			return fmt.Errorf("seeddata: bài viết %q trỏ tới cộng đồng lạ %q", p.Key, p.Community)
		}
		if p.AgeHours < 0 {
			return fmt.Errorf("seeddata: bài viết %q có AgeHours âm", p.Key)
		}
		for _, m := range p.Media {
			if m.URI == "" || m.Type == "" {
				return fmt.Errorf("seeddata: bài viết %q có media thiếu URI/Type", p.Key)
			}
		}
		for _, r := range p.Reactions {
			if err := requireUser("cảm xúc bài "+p.Key, r.User); err != nil {
				return err
			}
			if !reactionCodes[r.Emoji] {
				return fmt.Errorf("seeddata: bài viết %q dùng emoji cảm xúc lạ %q", p.Key, r.Emoji)
			}
		}
	}

	// Repost chỉ được tham chiếu bài gốc đã khai báo.
	for _, p := range Posts {
		if p.ShareOf == "" {
			continue
		}
		if !posts[p.ShareOf] {
			return fmt.Errorf("seeddata: repost %q trỏ tới bài gốc lạ %q", p.Key, p.ShareOf)
		}
		if p.ShareOf == p.Key {
			return fmt.Errorf("seeddata: repost %q tự tham chiếu chính nó", p.Key)
		}
	}

	// Bình luận lồng nhau.
	var checkComments func(postKey string, comments []Comment) error
	checkComments = func(postKey string, comments []Comment) error {
		for _, c := range comments {
			if err := requireUser("bình luận bài "+postKey, c.Author); err != nil {
				return err
			}
			if strings.TrimSpace(c.Text) == "" {
				return fmt.Errorf("seeddata: bình luận bài %q của %q rỗng", postKey, c.Author)
			}
			if c.AgeMin <= 0 {
				return fmt.Errorf("seeddata: bình luận bài %q của %q thiếu AgeMin", postKey, c.Author)
			}
			if err := checkComments(postKey, c.Replies); err != nil {
				return err
			}
		}
		return nil
	}
	for _, p := range Posts {
		if err := checkComments(p.Key, p.Comments); err != nil {
			return err
		}
	}

	// ── Quan hệ ──
	for _, f := range Follows {
		if err := requireUser("follow", f.Follower); err != nil {
			return err
		}
		if err := requireUser("follow", f.Following); err != nil {
			return err
		}
		if f.Follower == f.Following {
			return fmt.Errorf("seeddata: %q tự theo dõi chính mình", f.Follower)
		}
	}

	friendPairs := make(map[string]bool, len(Friends))
	for _, f := range Friends {
		if err := requireUser("friend", f.Sender); err != nil {
			return err
		}
		if err := requireUser("friend", f.Receiver); err != nil {
			return err
		}
		if f.Status != "accepted" && f.Status != "pending" {
			return fmt.Errorf("seeddata: friend %s→%s có status lạ %q", f.Sender, f.Receiver, f.Status)
		}
		pair := f.Sender + "|" + f.Receiver
		reverse := f.Receiver + "|" + f.Sender
		if friendPairs[pair] || friendPairs[reverse] {
			return fmt.Errorf("seeddata: cặp bạn bè trùng %s→%s", f.Sender, f.Receiver)
		}
		friendPairs[pair] = true
	}

	for _, b := range Blocks {
		if err := requireUser("block", b.User); err != nil {
			return err
		}
		if err := requireUser("block", b.Blocked); err != nil {
			return err
		}
	}

	for _, b := range Bookmarks {
		if err := requireUser("bookmark", b.User); err != nil {
			return err
		}
		if !posts[b.Post] {
			return fmt.Errorf("seeddata: bookmark trỏ tới bài lạ %q", b.Post)
		}
	}

	sharePairs := make(map[string]bool, len(Shares))
	for _, s := range Shares {
		if err := requireUser("share", s.User); err != nil {
			return err
		}
		if !posts[s.Post] {
			return fmt.Errorf("seeddata: share trỏ tới bài lạ %q", s.Post)
		}
		pair := s.User + "|" + s.Post
		if sharePairs[pair] {
			return fmt.Errorf("seeddata: share trùng %s→%s", s.User, s.Post)
		}
		sharePairs[pair] = true
	}

	// ── Stories ──
	storyKeys := make(map[string]bool, len(Stories))
	for _, s := range Stories {
		if s.Key == "" || storyKeys[s.Key] {
			return fmt.Errorf("seeddata: story key trống hoặc trùng %q", s.Key)
		}
		storyKeys[s.Key] = true
		if err := requireUser("story "+s.Key, s.User); err != nil {
			return err
		}
		if s.MediaURI == "" {
			return fmt.Errorf("seeddata: story %q thiếu MediaURI", s.Key)
		}
		if s.Type != "image" && s.Type != "video" {
			return fmt.Errorf("seeddata: story %q có Type lạ %q", s.Key, s.Type)
		}
		if s.AgeHours < 0 || s.AgeHours >= 24 {
			return fmt.Errorf("seeddata: story %q phải còn hạn (0 ≤ AgeHours < 24)", s.Key)
		}
		for _, v := range s.Viewers {
			if err := requireUser("story "+s.Key+" viewer", v); err != nil {
				return err
			}
		}
		for _, r := range s.Reacts {
			if err := requireUser("story "+s.Key+" react", r.User); err != nil {
				return err
			}
			if !reactionCodes[r.Emoji] {
				return fmt.Errorf("seeddata: story %q dùng emoji lạ %q", s.Key, r.Emoji)
			}
			if r.Click < 1 || r.Click > 5 {
				return fmt.Errorf("seeddata: story %q có Click ngoài 1..5", s.Key)
			}
		}
		for _, r := range s.Replies {
			if err := requireUser("story "+s.Key+" reply", r.User); err != nil {
				return err
			}
		}
	}

	for _, m := range StoryMutes {
		if err := requireUser("story mute", m.User); err != nil {
			return err
		}
		if err := requireUser("story mute", m.Muted); err != nil {
			return err
		}
	}

	// ── Chats ──
	chatKeys := make(map[string]bool, len(Chats))
	for _, c := range Chats {
		if c.Key == "" || chatKeys[c.Key] {
			return fmt.Errorf("seeddata: chat key trống hoặc trùng %q", c.Key)
		}
		chatKeys[c.Key] = true

		switch c.Type {
		case "direct":
			if len(c.Members) != 2 {
				return fmt.Errorf("seeddata: chat trực tiếp %q phải có đúng 2 thành viên", c.Key)
			}
		case "group":
			if len(c.Members) < 3 {
				return fmt.Errorf("seeddata: chat nhóm %q phải có ít nhất 3 thành viên", c.Key)
			}
		default:
			return fmt.Errorf("seeddata: chat %q có Type lạ %q", c.Key, c.Type)
		}

		members := make(map[string]bool, len(c.Members))
		for _, m := range c.Members {
			if err := requireUser("chat "+c.Key, m); err != nil {
				return err
			}
			if members[m] {
				return fmt.Errorf("seeddata: chat %q trùng thành viên %q", c.Key, m)
			}
			members[m] = true
		}
		for _, m := range c.UnreadFor {
			if !members[m] {
				return fmt.Errorf("seeddata: chat %q đánh dấu chưa đọc cho người ngoài %q", c.Key, m)
			}
		}
		for _, msg := range c.Messages {
			if !members[msg.From] {
				return fmt.Errorf("seeddata: chat %q có tin nhắn từ người ngoài %q", c.Key, msg.From)
			}
		}
	}

	// ── Quảng cáo ──
	packages := make(map[string]bool, len(AdPackages))
	for _, p := range AdPackages {
		if p.Key == "" || packages[p.Key] {
			return fmt.Errorf("seeddata: gói quảng cáo key trống hoặc trùng %q", p.Key)
		}
		packages[p.Key] = true
	}

	adKeys := make(map[string]bool, len(Ads))
	for _, a := range Ads {
		if a.Key == "" || adKeys[a.Key] {
			return fmt.Errorf("seeddata: quảng cáo key trống hoặc trùng %q", a.Key)
		}
		adKeys[a.Key] = true

		if err := requireUser("quảng cáo "+a.Key, a.Partner); err != nil {
			return err
		}
		if a.Package != "" && !packages[a.Package] {
			return fmt.Errorf("seeddata: quảng cáo %q trỏ tới gói lạ %q", a.Key, a.Package)
		}
		if len(a.Media) == 0 {
			return fmt.Errorf("seeddata: quảng cáo %q không có media", a.Key)
		}
		for _, m := range a.Media {
			if m.URL == "" {
				return fmt.Errorf("seeddata: quảng cáo %q có media thiếu URL", a.Key)
			}
			if m.Type != "image" && m.Type != "video" {
				return fmt.Errorf("seeddata: quảng cáo %q có media Type lạ %q", a.Key, m.Type)
			}
		}
		if a.TargetAgeMin < 0 || a.TargetAgeMax > 100 || a.TargetAgeMin > a.TargetAgeMax {
			return fmt.Errorf("seeddata: quảng cáo %q có khoảng tuổi không hợp lệ", a.Key)
		}
	}

	// ── Kiểm duyệt ──
	rules := make(map[string]bool, len(ViolationRules))
	for _, r := range ViolationRules {
		if r.Title == "" || rules[r.Title] {
			return fmt.Errorf("seeddata: quy tắc vi phạm tiêu đề trống hoặc trùng %q", r.Title)
		}
		rules[r.Title] = true
	}

	for _, r := range Reports {
		if err := requireUser("report", r.Reporter); err != nil {
			return err
		}
		targets := 0
		if r.TargetUser != "" {
			targets++
			if err := requireUser("report "+r.Reporter, r.TargetUser); err != nil {
				return err
			}
		}
		if r.TargetPost != "" {
			targets++
			if !posts[r.TargetPost] {
				return fmt.Errorf("seeddata: report trỏ tới bài lạ %q", r.TargetPost)
			}
		}
		if r.TargetCommunity != "" {
			targets++
			if !communities[r.TargetCommunity] {
				return fmt.Errorf("seeddata: report trỏ tới cộng đồng lạ %q", r.TargetCommunity)
			}
		}
		if targets != 1 {
			return fmt.Errorf("seeddata: report của %q phải có đúng một đối tượng", r.Reporter)
		}
		if r.Type != "violation" && r.Type != "other" {
			return fmt.Errorf("seeddata: report của %q có Type lạ %q", r.Reporter, r.Type)
		}
		if r.RuleTitle != "" && !rules[r.RuleTitle] {
			return fmt.Errorf("seeddata: report của %q trỏ tới quy tắc lạ %q", r.Reporter, r.RuleTitle)
		}
	}

	for _, b := range Bans {
		if err := requireUser("ban", b.User); err != nil {
			return err
		}
		if err := requireUser("ban admin", b.Admin); err != nil {
			return err
		}
	}

	for _, l := range ModLogs {
		if err := requireUser("mod log", l.Moderator); err != nil {
			return err
		}
		switch l.TargetType {
		case "post":
			if !posts[l.TargetKey] {
				return fmt.Errorf("seeddata: mod log trỏ tới bài lạ %q", l.TargetKey)
			}
		case "user":
			if !users[l.TargetKey] {
				return fmt.Errorf("seeddata: mod log trỏ tới user lạ %q", l.TargetKey)
			}
		case "community":
			if !communities[l.TargetKey] {
				return fmt.Errorf("seeddata: mod log trỏ tới cộng đồng lạ %q", l.TargetKey)
			}
		default:
			return fmt.Errorf("seeddata: mod log có TargetType lạ %q", l.TargetType)
		}
	}

	// ── Thông báo & cuộc gọi ──
	for _, n := range Notifications {
		if err := requireUser("notification receiver", n.Receiver); err != nil {
			return err
		}
		if err := requireUser("notification sender", n.Sender); err != nil {
			return err
		}
		if n.PostKey != "" && !posts[n.PostKey] {
			return fmt.Errorf("seeddata: notification trỏ tới bài lạ %q", n.PostKey)
		}
		if n.UserKey != "" {
			if err := requireUser("notification target", n.UserKey); err != nil {
				return err
			}
		}
	}

	for _, c := range Calls {
		if err := requireUser("call caller", c.Caller); err != nil {
			return err
		}
		if err := requireUser("call callee", c.Callee); err != nil {
			return err
		}
		if c.Caller == c.Callee {
			return fmt.Errorf("seeddata: cuộc gọi tự gọi chính mình (%q)", c.Caller)
		}
	}

	for _, u := range UploadQueue {
		if err := requireUser("media upload", u.User); err != nil {
			return err
		}
		switch u.Status {
		case "pending", "flagged", "rejected":
		default:
			return fmt.Errorf("seeddata: media upload của %q có Status lạ %q", u.User, u.Status)
		}
	}

	return nil
}

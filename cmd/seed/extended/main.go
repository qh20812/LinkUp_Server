package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"linkup/config"
	"linkup/db"
)

type userRow struct {
	ID    int
	Email string
}

type postRow struct {
	ID int
}

type emojiRow struct {
	ID int
}

type mediaRow struct {
	ID int
}

type commentRow struct {
	ID int
}

type roleRow struct {
	ID   int
	Name string
}

type chatRow struct {
	ID   int
	Type string
}

type communityRow struct {
	ID int
}

type violationRuleRow struct {
	ID int
}

type adRow struct {
	ID int
}

type seedPost struct {
	UserID int
	Title  string
	Body   string
	Views  int
	Status string
}

type seedStory struct {
	UserID    int
	MediaURI  string
	MediaType string
	Caption   string
}

type seedMedia struct {
	UserID   int
	PostID   sql.NullInt64
	FileURI  string
	FileType string
	FileSize float64
	Status   string
}

type seedReaction struct {
	UserID  int
	PostID  int
	EmojiID int
}

type seedComment struct {
	UserID   int
	PostID   int
	ParentID sql.NullInt64
	Content  string
}

type seedBookmark struct {
	UserID int
	PostID int
}

type seedTag struct {
	PostID       int
	CommentID    sql.NullInt64
	TagType      string
	TargetUserID sql.NullInt64
	Name         string
}

type seedFollow struct {
	FollowerID  int
	FollowingID int
}

type seedBlock struct {
	UserID        int
	BlockedUserID int
}

type seedFriend struct {
	SenderID   int
	ReceiverID int
	Status     string
}

type seedGroupMember struct {
	CommunityID int
	UserID      int
	Role        string
	Points      int
}

type seedParticipant struct {
	ChatID int
	UserID int
	Role   string
}

type seedMessage struct {
	ChatID   int
	SenderID int
	Content  string
	MediaID  sql.NullInt64
	EmojiID  sql.NullInt64
}

type seedCall struct {
	ChatID   sql.NullInt64
	CallerID int
	CallType string
	IsGroup  bool
	Status   string
}

type seedNotification struct {
	ReceiverID        int
	SenderID          sql.NullInt64
	Type              string
	RedirectPostID    sql.NullInt64
	RedirectUserID    sql.NullInt64
	RedirectCommentID sql.NullInt64
	Content           string
	IsRead            bool
}

type seedReport struct {
	ReporterID      int
	ReportType      string
	TargetUserID    sql.NullInt64
	TargetPostID    sql.NullInt64
	TargetCommentID sql.NullInt64
	ViolationRuleID int
	ReasonDetail    string
	Status          string
}

type seedBan struct {
	UserID    int
	AdminID   int
	Reason    string
	ExpiresAt sql.NullString
}

type seedModerationLog struct {
	ModeratorID int
	Action      string
	TargetType  string
	TargetID    int
	Reason      string
}

type seedAd struct {
	AdminID   int
	Title     string
	Content   string
	MediaID   sql.NullInt64
	TargetURL string
	Status    string
	Budget    float64
	StartedAt sql.NullString
	ExpiresAt sql.NullString
}

type seedAdAnalytics struct {
	AdID       int
	UserID     sql.NullInt64
	ActionType string
	IPAddress  string
}

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	conn, err := db.ConnectDb(config.GetEnv())
	if err != nil {
		log.Fatalf("DB connection: failed (%v)", err)
	}
	defer conn.Close()

	users, err := fetchUsers(conn)
	if err != nil {
		log.Fatalf("fetch users failed: %v", err)
	}
	if len(users) == 0 {
		log.Fatalf("no users found for phase7 seeding")
	}

	roles, err := fetchRoles(conn)
	if err != nil {
		log.Fatalf("fetch roles failed: %v", err)
	}
	if len(roles) == 0 {
		log.Fatalf("no roles found for phase7 seeding")
	}

	posts, err := fetchPosts(conn)
	if err != nil {
		log.Fatalf("fetch posts failed: %v", err)
	}
	if len(posts) == 0 {
		log.Fatalf("no posts found for phase7 seeding")
	}

	emojis, err := fetchEmojis(conn)
	if err != nil {
		log.Fatalf("fetch emojis failed: %v", err)
	}
	if len(emojis) == 0 {
		log.Fatalf("no emojis found for phase7 seeding")
	}

	media, err := fetchMedia(conn)
	if err != nil {
		log.Fatalf("fetch media failed: %v", err)
	}

	comments, err := fetchComments(conn)
	if err != nil {
		log.Fatalf("fetch comments failed: %v", err)
	}

	chats, err := fetchChats(conn)
	if err != nil {
		log.Fatalf("fetch chats failed: %v", err)
	}

	communities, err := fetchCommunities(conn)
	if err != nil {
		log.Fatalf("fetch communities failed: %v", err)
	}

	violationRules, err := fetchViolationRules(conn)
	if err != nil {
		log.Fatalf("fetch violation rules failed: %v", err)
	}
	if len(violationRules) == 0 {
		log.Fatalf("no violation rules found for phase7 seeding")
	}

	adIDs, err := fetchAdIDs(conn, 20)
	if err != nil {
		log.Fatalf("fetch ad ids failed: %v", err)
	}

	postsInserted, err := seedPosts(conn, buildPosts(users, 40))
	if err != nil {
		log.Fatalf("seed additional posts failed: %v", err)
	}

	storiesInserted, err := seedStories(conn, buildStories(users, 25))
	if err != nil {
		log.Fatalf("seed additional stories failed: %v", err)
	}

	mediaInserted, err := seedMediaItems(conn, buildMedia(users, posts))
	if err != nil {
		log.Fatalf("seed additional media failed: %v", err)
	}

	reactionsInserted, err := seedPostReactions(conn, buildReactions(users, posts, emojis))
	if err != nil {
		log.Fatalf("seed additional reactions failed: %v", err)
	}

	commentsInserted, err := seedComments(conn, buildComments(users, posts, comments))
	if err != nil {
		log.Fatalf("seed additional comments failed: %v", err)
	}

	bookmarksInserted, err := seedBookmarks(conn, buildBookmarks(users, posts))
	if err != nil {
		log.Fatalf("seed additional bookmarks failed: %v", err)
	}

	tagsInserted, err := seedTags(conn, buildTags(users, posts, comments))
	if err != nil {
		log.Fatalf("seed additional tags failed: %v", err)
	}

	followsInserted, err := seedFollows(conn, buildFollows(users))
	if err != nil {
		log.Fatalf("seed additional follows failed: %v", err)
	}

	blocksInserted, err := seedBlocks(conn, buildBlocks(users))
	if err != nil {
		log.Fatalf("seed additional blocks failed: %v", err)
	}

	friendsInserted, err := seedFriends(conn, buildFriends(users))
	if err != nil {
		log.Fatalf("seed additional friends failed: %v", err)
	}

	groupMembersInserted, err := seedGroupMembers(conn, buildGroupMembers(users, communities))
	if err != nil {
		log.Fatalf("seed additional group members failed: %v", err)
	}

	participantsInserted, err := seedChatParticipants(conn, buildChatParticipants(users, chats))
	if err != nil {
		log.Fatalf("seed additional chat participants failed: %v", err)
	}

	messagesInserted, err := seedMessages(conn, buildMessages(users, chats, emojis))
	if err != nil {
		log.Fatalf("seed additional messages failed: %v", err)
	}

	callsInserted, err := seedCalls(conn, buildCalls(users, chats))
	if err != nil {
		log.Fatalf("seed additional calls failed: %v", err)
	}

	notificationsInserted, err := seedNotifications(conn, buildNotifications(users, posts, comments))
	if err != nil {
		log.Fatalf("seed additional notifications failed: %v", err)
	}

	reportsInserted, err := seedReports(conn, buildReports(users, posts, comments, violationRules))
	if err != nil {
		log.Fatalf("seed additional reports failed: %v", err)
	}

	bansInserted, err := seedBans(conn, buildBans(users))
	if err != nil {
		log.Fatalf("seed additional bans failed: %v", err)
	}

	moderationInserted, err := seedModerationLogs(conn, buildModerationLogs(users, posts))
	if err != nil {
		log.Fatalf("seed additional moderation logs failed: %v", err)
	}

	adsInserted, err := seedAds(conn, buildAds(users, media))
	if err != nil {
		log.Fatalf("seed additional ads failed: %v", err)
	}

	adAnalyticsInserted, err := seedAdAnalyticsItems(conn, buildAdAnalytics(users, adIDs))
	if err != nil {
		log.Fatalf("seed additional ad analytics failed: %v", err)
	}

	fmt.Printf("Seed phase7: success (posts=%d, stories=%d, media=%d, reactions=%d, comments=%d, bookmarks=%d, tags=%d, follows=%d, blocks=%d, friends=%d, group_members=%d, chat_participants=%d, messages=%d, calls=%d, notifications=%d, reports=%d, bans=%d, moderation_logs=%d, ads=%d, ad_analytics=%d)\n",
		postsInserted,
		storiesInserted,
		mediaInserted,
		reactionsInserted,
		commentsInserted,
		bookmarksInserted,
		tagsInserted,
		followsInserted,
		blocksInserted,
		friendsInserted,
		groupMembersInserted,
		participantsInserted,
		messagesInserted,
		callsInserted,
		notificationsInserted,
		reportsInserted,
		bansInserted,
		moderationInserted,
		adsInserted,
		adAnalyticsInserted,
	)
}

func fetchUsers(conn *sql.DB) ([]userRow, error) {
	rows, err := conn.Query("SELECT id, email FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	users := make([]userRow, 0)
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func fetchRoles(conn *sql.DB) ([]roleRow, error) {
	rows, err := conn.Query("SELECT id, name FROM roles ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]roleRow, 0)
	for rows.Next() {
		var r roleRow
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

func fetchPosts(conn *sql.DB) ([]postRow, error) {
	rows, err := conn.Query("SELECT id FROM posts ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query posts: %w", err)
	}
	defer rows.Close()

	posts := make([]postRow, 0)
	for rows.Next() {
		var p postRow
		if err := rows.Scan(&p.ID); err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func fetchEmojis(conn *sql.DB) ([]emojiRow, error) {
	rows, err := conn.Query("SELECT id FROM emojis ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query emojis: %w", err)
	}
	defer rows.Close()

	emojis := make([]emojiRow, 0)
	for rows.Next() {
		var e emojiRow
		if err := rows.Scan(&e.ID); err != nil {
			return nil, fmt.Errorf("scan emoji: %w", err)
		}
		emojis = append(emojis, e)
	}
	return emojis, rows.Err()
}

func fetchMedia(conn *sql.DB) ([]mediaRow, error) {
	rows, err := conn.Query("SELECT id FROM media ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query media: %w", err)
	}
	defer rows.Close()

	media := make([]mediaRow, 0)
	for rows.Next() {
		var m mediaRow
		if err := rows.Scan(&m.ID); err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}
		media = append(media, m)
	}
	return media, rows.Err()
}

func fetchComments(conn *sql.DB) ([]commentRow, error) {
	rows, err := conn.Query("SELECT id FROM comments ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query comments: %w", err)
	}
	defer rows.Close()

	comments := make([]commentRow, 0)
	for rows.Next() {
		var c commentRow
		if err := rows.Scan(&c.ID); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func fetchChats(conn *sql.DB) ([]chatRow, error) {
	rows, err := conn.Query("SELECT id, type FROM chats ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query chats: %w", err)
	}
	defer rows.Close()

	chats := make([]chatRow, 0)
	for rows.Next() {
		var c chatRow
		if err := rows.Scan(&c.ID, &c.Type); err != nil {
			return nil, fmt.Errorf("scan chat: %w", err)
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func fetchCommunities(conn *sql.DB) ([]communityRow, error) {
	rows, err := conn.Query("SELECT id FROM communities ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query communities: %w", err)
	}
	defer rows.Close()

	communities := make([]communityRow, 0)
	for rows.Next() {
		var c communityRow
		if err := rows.Scan(&c.ID); err != nil {
			return nil, fmt.Errorf("scan community: %w", err)
		}
		communities = append(communities, c)
	}
	return communities, rows.Err()
}

func fetchViolationRules(conn *sql.DB) ([]violationRuleRow, error) {
	rows, err := conn.Query("SELECT id FROM violation_rules ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query violation rules: %w", err)
	}
	defer rows.Close()

	rules := make([]violationRuleRow, 0)
	for rows.Next() {
		var r violationRuleRow
		if err := rows.Scan(&r.ID); err != nil {
			return nil, fmt.Errorf("scan violation rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func fetchAdIDs(conn *sql.DB, limit int) ([]int, error) {
	rows, err := conn.Query("SELECT id FROM ads ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("query ad ids: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan ad id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func buildPosts(users []userRow, total int) []seedPost {
	posts := make([]seedPost, 0, total)
	for i := 1; i <= total; i++ {
		user := users[i%len(users)]
		status := "active"
		if i%10 == 0 {
			status = "hidden"
		}
		posts = append(posts, seedPost{
			UserID: user.ID,
			Title:  fmt.Sprintf("Phase7 bài viết %03d", i),
			Body:   fmt.Sprintf("Nội dung mở rộng cho bài viết phase7 #%d từ tài khoản %s.", i, user.Email),
			Views:  i * 13,
			Status: status,
		})
	}
	return posts
}

func buildStories(users []userRow, total int) []seedStory {
	stories := make([]seedStory, 0, total)
	for i := 1; i <= total; i++ {
		user := users[i%len(users)]
		mediaType := "image"
		if i%4 == 0 {
			mediaType = "video"
		}
		stories = append(stories, seedStory{
			UserID:    user.ID,
			MediaURI:  fmt.Sprintf("/seeds/phase7/story-%03d.%s", i, map[string]string{"image": "jpg", "video": "mp4"}[mediaType]),
			MediaType: mediaType,
			Caption:   fmt.Sprintf("Story phase7 cho %s.", user.Email),
		})
	}
	return stories
}

func buildMedia(users []userRow, posts []postRow) []seedMedia {
	items := make([]seedMedia, 0)
	for i, post := range posts {
		owner := users[i%len(users)]
		fileType := "image"
		if i%3 == 0 {
			fileType = "video"
		}
		items = append(items, seedMedia{
			UserID:   owner.ID,
			PostID:   sql.NullInt64{Int64: int64(post.ID), Valid: true},
			FileURI:  fmt.Sprintf("/seeds/phase7/media-%03d.%s", i+1, map[string]string{"image": "jpg", "video": "mp4"}[fileType]),
			FileType: fileType,
			FileSize: 1.0 + float64(i%6)*0.5,
			Status:   "approved",
		})
	}
	return items
}

func buildReactions(users []userRow, posts []postRow, emojis []emojiRow) []seedReaction {
	items := make([]seedReaction, 0)
	for i, user := range users {
		for j := 0; j < 2 && j < len(posts); j++ {
			items = append(items, seedReaction{UserID: user.ID, PostID: posts[(i+j)%len(posts)].ID, EmojiID: emojis[(i+j)%len(emojis)].ID})
		}
	}
	return items
}

func buildComments(users []userRow, posts []postRow, existingComments []commentRow) []seedComment {
	items := make([]seedComment, 0)
	for i := 0; i < 30; i++ {
		user := users[i%len(users)]
		postID := posts[i%len(posts)].ID
		parent := sql.NullInt64{Valid: false}
		if len(existingComments) > 0 && i%5 == 0 {
			parent = sql.NullInt64{Int64: int64(existingComments[i%len(existingComments)].ID), Valid: true}
		}
		items = append(items, seedComment{UserID: user.ID, PostID: postID, ParentID: parent, Content: fmt.Sprintf("Phản hồi phase7 cho bài viết %d.", postID)})
	}
	return items
}

func buildBookmarks(users []userRow, posts []postRow) []seedBookmark {
	items := make([]seedBookmark, 0)
	for i := 0; i < 20; i++ {
		items = append(items, seedBookmark{UserID: users[i%len(users)].ID, PostID: posts[(i*2)%len(posts)].ID})
	}
	return items
}

func buildTags(users []userRow, posts []postRow, comments []commentRow) []seedTag {
	items := make([]seedTag, 0)
	for i := 0; i < 18; i++ {
		postID := posts[i%len(posts)].ID
		commentID := sql.NullInt64{Valid: false}
		if len(comments) > 0 {
			commentID = sql.NullInt64{Int64: int64(comments[i%len(comments)].ID), Valid: true}
		}
		targetUser := sql.NullInt64{Valid: false}
		if i%3 == 0 {
			targetUser = sql.NullInt64{Int64: int64(users[i%len(users)].ID), Valid: true}
		}
		items = append(items, seedTag{PostID: postID, CommentID: commentID, TagType: "mention", TargetUserID: targetUser, Name: fmt.Sprintf("@%s", strings.SplitN(users[i%len(users)].Email, "@", 2)[0])})
		items = append(items, seedTag{PostID: postID, CommentID: sql.NullInt64{Valid: false}, TagType: "hashtag", TargetUserID: sql.NullInt64{Valid: false}, Name: fmt.Sprintf("#phase7_%d", i+1)})
	}
	return items
}

func buildFollows(users []userRow) []seedFollow {
	items := make([]seedFollow, 0)
	seen := make(map[string]struct{})
	for i := 0; i < len(users); i++ {
		follower := users[i]
		for j := 1; j <= 3; j++ {
			following := users[(i+j)%len(users)]
			if follower.ID == following.ID {
				continue
			}
			key := fmt.Sprintf("%d_%d", follower.ID, following.ID)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, seedFollow{FollowerID: follower.ID, FollowingID: following.ID})
		}
	}
	return items
}

func buildBlocks(users []userRow) []seedBlock {
	items := make([]seedBlock, 0)
	for i := 0; i < 12; i++ {
		items = append(items, seedBlock{UserID: users[i%len(users)].ID, BlockedUserID: users[(i+4)%len(users)].ID})
	}
	return items
}

func buildFriends(users []userRow) []seedFriend {
	items := make([]seedFriend, 0)
	statuses := []string{"accepted", "pending", "rejected"}
	for i := 0; i < 18; i++ {
		items = append(items, seedFriend{SenderID: users[i%len(users)].ID, ReceiverID: users[(i+2)%len(users)].ID, Status: statuses[i%len(statuses)]})
	}
	return items
}

func buildGroupMembers(users []userRow, communities []communityRow) []seedGroupMember {
	items := make([]seedGroupMember, 0)
	for ci, community := range communities {
		for j := 0; j < 3; j++ {
			member := users[(ci+j)%len(users)]
			role := "member"
			if j == 0 {
				role = "admin"
			}
			items = append(items, seedGroupMember{CommunityID: community.ID, UserID: member.ID, Role: role, Points: 20 * j})
		}
	}
	return items
}

func buildChatParticipants(users []userRow, chats []chatRow) []seedParticipant {
	items := make([]seedParticipant, 0)
	for i, chat := range chats {
		for j := 0; j < 3; j++ {
			items = append(items, seedParticipant{ChatID: chat.ID, UserID: users[(i+j)%len(users)].ID, Role: "member"})
		}
	}
	return items
}

func buildMessages(users []userRow, chats []chatRow, emojis []emojiRow) []seedMessage {
	texts := []string{"Phase7 trực tiếp", "Tin nhắn mở rộng", "Cập nhật dữ liệu seed", "Kiểm tra chat", "Nội dung thử nghiệm"}
	items := make([]seedMessage, 0)
	for i, chat := range chats {
		for j := 0; j < 4; j++ {
			sender := users[(i+j)%len(users)]
			emoji := sql.NullInt64{Valid: false}
			if len(emojis) > 0 && j%2 == 0 {
				emoji = sql.NullInt64{Int64: int64(emojis[(i+j)%len(emojis)].ID), Valid: true}
			}
			items = append(items, seedMessage{ChatID: chat.ID, SenderID: sender.ID, Content: texts[(i+j)%len(texts)], MediaID: sql.NullInt64{Valid: false}, EmojiID: emoji})
		}
	}
	return items
}

func buildCalls(users []userRow, chats []chatRow) []seedCall {
	statuses := []string{"completed", "missed", "declined"}
	items := make([]seedCall, 0)
	for i, chat := range chats {
		items = append(items, seedCall{ChatID: sql.NullInt64{Int64: int64(chat.ID), Valid: true}, CallerID: users[i%len(users)].ID, CallType: "voice", IsGroup: strings.EqualFold(chat.Type, "group"), Status: statuses[i%len(statuses)]})
	}
	return items
}

func buildNotifications(users []userRow, posts []postRow, comments []commentRow) []seedNotification {
	items := make([]seedNotification, 0)
	for i := 0; i < 20; i++ {
		sender := users[(i+1)%len(users)]
		receiver := users[(i+2)%len(users)]
		items = append(items, seedNotification{ReceiverID: receiver.ID, SenderID: sql.NullInt64{Int64: int64(sender.ID), Valid: true}, Type: "message", RedirectPostID: sql.NullInt64{Valid: false}, RedirectUserID: sql.NullInt64{Int64: int64(sender.ID), Valid: true}, RedirectCommentID: sql.NullInt64{Valid: false}, Content: fmt.Sprintf("%s đã gửi tin nhắn mới.", sender.Email), IsRead: i%2 == 0})
		if len(posts) > 0 && i%3 == 0 {
			items = append(items, seedNotification{ReceiverID: receiver.ID, SenderID: sql.NullInt64{Int64: int64(sender.ID), Valid: true}, Type: "comment", RedirectPostID: sql.NullInt64{Int64: int64(posts[i%len(posts)].ID), Valid: true}, RedirectUserID: sql.NullInt64{Valid: false}, RedirectCommentID: sql.NullInt64{Valid: false}, Content: fmt.Sprintf("%s đã bình luận bài viết của bạn.", sender.Email), IsRead: i%3 == 0})
		}
	}
	return items
}

func buildReports(users []userRow, posts []postRow, comments []commentRow, rules []violationRuleRow) []seedReport {
	items := make([]seedReport, 0)
	for i := 0; i < 12; i++ {
		reporter := users[i%len(users)]
		items = append(items, seedReport{ReporterID: reporter.ID, ReportType: "post", TargetUserID: sql.NullInt64{Valid: false}, TargetPostID: sql.NullInt64{Int64: int64(posts[i%len(posts)].ID), Valid: true}, TargetCommentID: sql.NullInt64{Valid: false}, ViolationRuleID: rules[i%len(rules)].ID, ReasonDetail: "Báo cáo nội dung vi phạm.", Status: "pending"})
	}
	for i := 0; i < 8; i++ {
		reporter := users[(i+2)%len(users)]
		items = append(items, seedReport{ReporterID: reporter.ID, ReportType: "comment", TargetUserID: sql.NullInt64{Valid: false}, TargetPostID: sql.NullInt64{Valid: false}, TargetCommentID: sql.NullInt64{Int64: int64(comments[i%len(comments)].ID), Valid: true}, ViolationRuleID: rules[(i+3)%len(rules)].ID, ReasonDetail: "Bình luận không phù hợp.", Status: "reviewed"})
	}
	return items
}

func buildBans(users []userRow) []seedBan {
	items := make([]seedBan, 0)
	for i := 0; i < 5; i++ {
		items = append(items, seedBan{UserID: users[(i+3)%len(users)].ID, AdminID: users[0].ID, Reason: fmt.Sprintf("Vi phạm chính sách phase7 #%d.", i+1), ExpiresAt: sql.NullString{String: "2027-01-01 00:00:00", Valid: true}})
	}
	return items
}

func buildModerationLogs(users []userRow, posts []postRow) []seedModerationLog {
	actions := []string{"DELETE_POST", "REVIEW_REPORT", "BAN_USER"}
	items := make([]seedModerationLog, 0)
	for i := 0; i < 10; i++ {
		items = append(items, seedModerationLog{ModeratorID: users[0].ID, Action: actions[i%len(actions)], TargetType: "posts", TargetID: posts[i%len(posts)].ID, Reason: fmt.Sprintf("Ghi nhận kiểm duyệt phase7 #%d.", i+1)})
	}
	return items
}

func buildAds(users []userRow, media []mediaRow) []seedAd {
	items := make([]seedAd, 0)
	for i := 0; i < 8; i++ {
		mediaID := sql.NullInt64{Valid: false}
		if len(media) > 0 {
			mediaID = sql.NullInt64{Int64: int64(media[i%len(media)].ID), Valid: true}
		}
		items = append(items, seedAd{AdminID: users[i%len(users)].ID, Title: fmt.Sprintf("Quảng cáo phase7 #%d", i+1), Content: fmt.Sprintf("Nội dung quảng cáo phase7 số %d.", i+1), MediaID: mediaID, TargetURL: fmt.Sprintf("https://example.com/phase7/%d", i+1), Status: "active", Budget: 500.0 + float64(i)*100.0, StartedAt: sql.NullString{String: "2026-03-01 00:00:00", Valid: true}, ExpiresAt: sql.NullString{String: "2026-12-31 23:59:59", Valid: true}})
	}
	return items
}

func buildAdAnalytics(users []userRow, adIDs []int) []seedAdAnalytics {
	items := make([]seedAdAnalytics, 0)
	actions := []string{"view", "click"}
	for i, adID := range adIDs {
		items = append(items, seedAdAnalytics{AdID: adID, UserID: sql.NullInt64{Int64: int64(users[(i+1)%len(users)].ID), Valid: true}, ActionType: actions[i%len(actions)], IPAddress: fmt.Sprintf("10.0.0.%d", i+10)})
	}
	return items
}

func seedPosts(conn *sql.DB, rows []seedPost) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.Title, item.Body, item.Views, item.Status})
	}
	return bulkInsertIgnore(conn, "posts", []string{"user_id", "title", "content", "views_count", "status"}, values)
}

func seedStories(conn *sql.DB, rows []seedStory) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.MediaURI, item.MediaType, item.Caption, sql.NullString{Valid: false}})
	}
	return bulkInsertIgnore(conn, "stories", []string{"user_id", "media_uri", "media_type", "caption", "expires_at"}, values)
}

func seedMediaItems(conn *sql.DB, rows []seedMedia) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.PostID, item.FileURI, item.FileType, item.FileSize, item.Status})
	}
	return bulkInsertIgnore(conn, "media", []string{"user_id", "post_id", "file_uri", "file_type", "file_size", "status"}, values)
}

func seedPostReactions(conn *sql.DB, rows []seedReaction) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.PostID, item.EmojiID})
	}
	return bulkInsertIgnore(conn, "post_reactions", []string{"user_id", "post_id", "emoji_id"}, values)
}

func seedComments(conn *sql.DB, rows []seedComment) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.PostID, item.ParentID, item.Content})
	}
	return bulkInsertIgnore(conn, "comments", []string{"user_id", "post_id", "parent_id", "content"}, values)
}

func seedBookmarks(conn *sql.DB, rows []seedBookmark) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.PostID})
	}
	return bulkInsertIgnore(conn, "bookmarks", []string{"user_id", "post_id"}, values)
}

func seedTags(conn *sql.DB, rows []seedTag) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.PostID, item.CommentID, item.TagType, item.TargetUserID, item.Name})
	}
	return bulkInsertIgnore(conn, "tags", []string{"post_id", "comment_id", "tag_type", "target_user_id", "name"}, values)
}

func seedFollows(conn *sql.DB, rows []seedFollow) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.FollowerID, item.FollowingID})
	}
	return bulkInsertIgnore(conn, "follows", []string{"follower_id", "following_id"}, values)
}

func seedBlocks(conn *sql.DB, rows []seedBlock) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.BlockedUserID})
	}
	return bulkInsertIgnore(conn, "blocks", []string{"user_id", "blocked_user_id"}, values)
}

func seedFriends(conn *sql.DB, rows []seedFriend) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.SenderID, item.ReceiverID, item.Status})
	}
	return bulkInsertIgnore(conn, "friends", []string{"sender_id", "receiver_id", "status"}, values)
}

func seedGroupMembers(conn *sql.DB, rows []seedGroupMember) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.CommunityID, item.UserID, item.Role, item.Points})
	}
	return bulkInsertIgnore(conn, "group_members", []string{"community_id", "user_id", "role", "points"}, values)
}

func seedChatParticipants(conn *sql.DB, rows []seedParticipant) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ChatID, item.UserID, item.Role})
	}
	return bulkInsertIgnore(conn, "chat_participants", []string{"chat_id", "user_id", "role"}, values)
}

func seedMessages(conn *sql.DB, rows []seedMessage) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ChatID, item.SenderID, item.Content, item.MediaID, item.EmojiID})
	}
	return bulkInsertIgnore(conn, "messages", []string{"chat_id", "sender_id", "content", "media_id", "emoji_id"}, values)
}

func seedCalls(conn *sql.DB, rows []seedCall) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ChatID, item.CallerID, item.CallType, item.IsGroup, item.Status})
	}
	return bulkInsertIgnore(conn, "calls", []string{"chat_id", "caller_id", "call_type", "is_group", "status"}, values)
}

func seedNotifications(conn *sql.DB, rows []seedNotification) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ReceiverID, item.SenderID, item.Type, item.RedirectPostID, item.RedirectUserID, item.RedirectCommentID, item.Content, item.IsRead})
	}
	return bulkInsertIgnore(conn, "notifications", []string{"receiver_id", "sender_id", "type", "redirect_post_id", "redirect_user_id", "redirect_comment_id", "content", "is_read"}, values)
}

func seedReports(conn *sql.DB, rows []seedReport) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ReporterID, item.ReportType, item.TargetUserID, item.TargetPostID, item.TargetCommentID, item.ViolationRuleID, item.ReasonDetail, item.Status})
	}
	return bulkInsertIgnore(conn, "reports", []string{"reporter_id", "report_type", "target_user_id", "target_post_id", "target_comment_id", "violation_rule_id", "reason_detail", "status"}, values)
}

func seedBans(conn *sql.DB, rows []seedBan) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.UserID, item.AdminID, item.Reason, item.ExpiresAt})
	}
	return bulkInsertIgnore(conn, "bans", []string{"user_id", "admin_id", "reason", "expires_at"}, values)
}

func seedModerationLogs(conn *sql.DB, rows []seedModerationLog) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.ModeratorID, item.Action, item.TargetType, item.TargetID, item.Reason})
	}
	return bulkInsertIgnore(conn, "moderation_logs", []string{"moderator_id", "action", "target_type", "target_id", "reason"}, values)
}

func seedAds(conn *sql.DB, rows []seedAd) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.AdminID, item.Title, item.Content, item.MediaID, item.TargetURL, item.Status, item.Budget, item.StartedAt, item.ExpiresAt})
	}
	return bulkInsertIgnore(conn, "ads", []string{"admin_id", "title", "content", "media_id", "target_url", "status", "budget", "started_at", "expires_at"}, values)
}

func seedAdAnalyticsItems(conn *sql.DB, rows []seedAdAnalytics) (int64, error) {
	values := make([][]any, 0, len(rows))
	for _, item := range rows {
		values = append(values, []any{item.AdID, item.UserID, item.ActionType, item.IPAddress})
	}
	return bulkInsertIgnore(conn, "ad_analytics", []string{"ad_id", "user_id", "action_type", "ip_address"}, values)
}

func bulkInsertIgnore(conn *sql.DB, table string, columns []string, rows [][]any) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	var builder strings.Builder
	builder.WriteString("INSERT IGNORE INTO ")
	builder.WriteString(table)
	builder.WriteString(" (")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(") VALUES ")

	args := make([]any, 0, len(rows)*len(columns))
	for i, row := range rows {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("(")
		for j := range columns {
			if j > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString("?")
		}
		builder.WriteString(")")
		args = append(args, row...)
	}

	result, err := conn.Exec(builder.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("insert seed rows into %s: %w", table, err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected for %s: %w", table, err)
	}
	return inserted, nil
}

package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linkup/config"
	"linkup/models"

	"gorm.io/gorm"
)

// PersonalCursor là con trỏ phân trang cho rank v2 (3 stream độc lập).
// Format: v2_{snapshotNano}_{mainScore}_{mainID}_{smallNano}_{smallID}_{genScore}_{genID}
// Stream chưa bắt đầu dùng score/nano = -1 và id = "-". Stream đã cạn dùng
// id = "!" (cờ Done) để trang sau không query lại từ đầu — nếu không, stream
// cạn sẽ bị đọc lại và sinh bài trùng giữa các trang.
type PersonalCursor struct {
	MainScore float64
	MainID    string
	MainOK    bool
	MainDone  bool

	SmallNano int64
	SmallID   string
	SmallOK   bool
	SmallDone bool

	GenScore float64
	GenID    string
	GenOK    bool
	GenDone  bool
}

// cursorDoneID đánh dấu stream đã cạn trong cursor (không phải UUID nên an toàn).
const cursorDoneID = "!"

func (c PersonalCursor) Encode(snapshot time.Time) string {
	mScore, mID := -1.0, "-"
	if c.MainDone {
		mScore, mID = 0, cursorDoneID
	} else if c.MainOK {
		mScore, mID = c.MainScore, c.MainID
	}
	sNano, sID := int64(-1), "-"
	if c.SmallDone {
		sNano, sID = 0, cursorDoneID
	} else if c.SmallOK {
		sNano, sID = c.SmallNano, c.SmallID
	}
	gScore, gID := -1.0, "-"
	if c.GenDone {
		gScore, gID = 0, cursorDoneID
	} else if c.GenOK {
		gScore, gID = c.GenScore, c.GenID
	}
	return fmt.Sprintf("v2_%d_%f_%s_%d_%s_%f_%s",
		snapshot.UnixNano(), mScore, mID, sNano, sID, gScore, gID)
}

// ParsePersonalCursor parse cursor v2; trả về ok=false khi là cursor v1/cũ
// (caller sẽ query mới từ đầu).
func ParsePersonalCursor(raw string) (PersonalCursor, time.Time, bool) {
	var c PersonalCursor
	parts := strings.SplitN(raw, "_", 8)
	if len(parts) != 8 || parts[0] != "v2" {
		return c, time.Time{}, false
	}
	snapNano, err1 := strconv.ParseInt(parts[1], 10, 64)
	mScore, err2 := strconv.ParseFloat(parts[2], 64)
	sNano, err3 := strconv.ParseInt(parts[4], 10, 64)
	gScore, err4 := strconv.ParseFloat(parts[6], 64)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return c, time.Time{}, false
	}
	switch {
	case parts[3] == cursorDoneID:
		c.MainDone = true
	case parts[3] != "-" && mScore >= 0:
		c.MainScore, c.MainID, c.MainOK = mScore, parts[3], true
	}
	if parts[5] == cursorDoneID {
		c.SmallDone = true
	} else if parts[5] != "-" && sNano >= 0 {
		c.SmallNano, c.SmallID, c.SmallOK = sNano, parts[5], true
	}
	if parts[7] == cursorDoneID {
		c.GenDone = true
	} else if parts[7] != "-" && gScore >= 0 {
		c.GenScore, c.GenID, c.GenOK = gScore, parts[7], true
	}
	return c, time.Unix(0, snapNano), true
}

// feedAuthorSelect là cột chung cho mọi stream feed (author + is_following).
const feedAuthorSelect = `posts.*,
	users.username,
	COALESCE(profiles.display_name, users.username) AS display_name,
	COALESCE(profiles.avatar_uri, '') AS avatar_uri,
	CASE WHEN f.follower_id IS NOT NULL THEN true ELSE false END AS is_following`

// ageHoursExpr là biểu thức tuổi bài (giờ) dùng chung cho mọi stream.
// Chứa đúng một placeholder `?` (thời điểm snapshot) và KHÔNG có verb fmt:
// dùng thẳng trong fmt.Sprintf/Select, đừng bọc `fmt.Sprintf(ageHoursExpr, "?")`
// — làm vậy sẽ sinh `%!(EXTRA string=?)` và MySQL báo lỗi 1064.
const ageHoursExpr = "((UNIX_TIMESTAMP(?) - UNIX_TIMESTAMP(posts.created_at)) / 3600.0)"

// applyFeedVisibility lọc bài theo visibility + block 2 chiều.
// Khách vãng lai (userID rỗng) chỉ thấy public.
func applyFeedVisibility(q *gorm.DB, userID string, includeFriend bool) *gorm.DB {
	if userID == "" {
		return q.Where("posts.status = ?", models.PostStatusPublic)
	}
	if includeFriend {
		q = q.Where("(posts.status = ? OR (posts.status = ? AND EXISTS (SELECT 1 FROM friends WHERE ((sender_id = ? AND receiver_id = posts.user_id) OR (sender_id = posts.user_id AND receiver_id = ?)) AND status = 'accepted')))",
			models.PostStatusPublic, models.PostStatusFriend, userID, userID)
	} else {
		q = q.Where("posts.status = ?", models.PostStatusPublic)
	}
	return q.Where("NOT EXISTS (SELECT 1 FROM blocks WHERE user_id = ? AND blocked_user_id = posts.user_id)", userID).
		Where("NOT EXISTS (SELECT 1 FROM blocks WHERE user_id = posts.user_id AND blocked_user_id = ?)", userID).
		// Tín hiệu âm chủ động: user đã bấm "Không quan tâm" thì bài không
		// được quay lại feed của chính họ.
		Where("NOT EXISTS (SELECT 1 FROM post_hidden ph WHERE ph.post_id = posts.id AND ph.user_id = ?)", userID)
}

func feedBaseJoins(q *gorm.DB, userID *string) *gorm.DB {
	return q.Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Joins("LEFT JOIN follows f ON f.following_id = posts.user_id AND f.follower_id = ?", userID)
}

// boostCaseExpr trả biểu thức bonus giảm tuyến tính theo tuổi bài trong
// `hours` giờ. extraCond là điều kiện bổ sung (ví dụ loại repost) và đứng
// trước biểu thức tuổi trong SQL, nên placeholder của nó phải được truyền
// trước các placeholder của ageExpr.
func boostCaseExpr(extraCond, ageExpr string, hours, bonus float64) string {
	cond := ""
	if extraCond != "" {
		cond = extraCond + " AND "
	}
	return fmt.Sprintf(`CASE WHEN %s(%s) < %f THEN %f * (1 - (%s) / %f) ELSE 0 END`,
		cond, ageExpr, hours, bonus, ageExpr, hours)
}

// scoreTerm là một thành phần của công thức điểm feed v2. Expr đã bao gồm
// trọng số và dấu (thành phần phạt là số âm), nên điểm cuối = tổng các Expr.
type scoreTerm struct {
	Name string // tên cột khi trả về ?debug=1
	Expr string
}

func joinScoreTerms(terms []scoreTerm) string {
	exprs := make([]string, len(terms))
	for i, t := range terms {
		exprs[i] = t.Expr
	}
	return strings.Join(exprs, " + ")
}

// mainScoreTerms dựng công thức điểm main stream theo đúng thứ tự placeholder
// của mainScoreArgs. Tách khỏi query để test được số placeholder mà không cần DB.
func mainScoreTerms(w config.FeedWeights, v2 config.FeedRankV2Config) []scoreTerm {
	ageH := ageHoursExpr

	// Interest dùng snapshotTime (?) thay vì NOW() để trang 1 và trang N
	// tính trên cùng một mốc thời gian.
	interestSub := fmt.Sprintf(`(SELECT COALESCE(SUM(ui.score * EXP(-%f * ((UNIX_TIMESTAMP(?) - UNIX_TIMESTAMP(ui.updated_at)) / 86400.0))), 0)
		FROM user_interests ui
		WHERE ui.user_id = ?
		AND ui.tag IN (SELECT t.name FROM tags t WHERE t.post_id = posts.id AND t.tag_type = 'hashtag' AND t.comment_id IS NULL))`,
		v2.InterestDecayRate)

	affinityCase := fmt.Sprintf(`CASE WHEN f.follower_id IS NOT NULL THEN %f
		WHEN EXISTS (SELECT 1 FROM friends WHERE ((sender_id = ? AND receiver_id = posts.user_id) OR (sender_id = posts.user_id AND receiver_id = ?)) AND status = 'accepted') THEN %f
		WHEN EXISTS (SELECT 1 FROM post_reactions pr JOIN posts p2 ON p2.id = pr.post_id WHERE pr.user_id = ? AND p2.user_id = posts.user_id)
		  OR EXISTS (SELECT 1 FROM comments cm JOIN posts p3 ON p3.id = cm.post_id WHERE cm.user_id = ? AND p3.user_id = posts.user_id)
		  OR EXISTS (SELECT 1 FROM post_views pv JOIN posts p4 ON p4.id = pv.post_id WHERE pv.viewer_id = ? AND p4.user_id = posts.user_id) THEN %f
		ELSE 0.0 END`,
		v2.AffinityFollow, v2.AffinityFriend, v2.AffinityInteracted)

	seenSub := fmt.Sprintf(`(SELECT COUNT(*) FROM post_views pv WHERE pv.post_id = posts.id AND pv.viewer_id = ? AND pv.created_at > DATE_SUB(?, INTERVAL %d DAY))`,
		v2.SeenWindowDays)

	// post_views dedup 1 user/post/ngày nên bậc 7 ngày chỉ tăng 1 lần/ngày;
	// thêm phạt riêng cho "đã xem trong 24h" để bài vừa đọc rời top ngay.
	seenRecentSub := `(SELECT COUNT(*) FROM post_views pv WHERE pv.post_id = posts.id AND pv.viewer_id = ? AND pv.created_at > DATE_SUB(?, INTERVAL 1 DAY))`

	// Repost không được boost bài mới: nội dung không mới, chỉ là hành động
	// share của người khác (xem ownBoost ở FetchActive).
	newBoost := boostCaseExpr("posts.shared_from_post_id IS NULL",
		ageH, v2.NewPostBoostHours, v2.NewPostBoost)

	// Bonus tác giả nhỏ cũng không áp cho repost.
	smallBonus := fmt.Sprintf(`CASE WHEN posts.shared_from_post_id IS NULL AND (SELECT COUNT(*) FROM follows WHERE following_id = posts.user_id) < %d THEN %f ELSE 0 END`,
		v2.SmallAuthorFollowers, v2.SmallAuthorBonus)

	hiddenAuthorSub := `(SELECT COUNT(*) FROM post_hidden ph WHERE ph.user_id = ? AND EXISTS (SELECT 1 FROM posts px WHERE px.id = ph.post_id AND px.user_id = posts.user_id))`

	return []scoreTerm{
		{"recency", fmt.Sprintf("%f * EXP(-%f * %s)", w.Recency, w.DecayRate, ageH)},
		{"engagement", fmt.Sprintf("%f * (LOG(1 + COALESCE(lr.likes, 0)) * %f + LOG(1 + COALESCE(cr.comments, 0)) * %f + LOG(1 + COALESCE(sr.shares, 0)) * %f) / 20.0",
			w.Engagement, w.LikeWeight, w.CommentWeight, w.ShareWeight)},
		{"affinity", fmt.Sprintf("%f * (%s)", w.Affinity, affinityCase)},
		{"interest", fmt.Sprintf("%f * LEAST(%s, 10) / 10", v2.InterestWeight, interestSub)},
		{"new_boost", newBoost},
		{"small_bonus", smallBonus},
		{"seen_penalty", fmt.Sprintf("-%f * LEAST(%s, %d)", v2.SeenPenalty, seenSub, v2.SeenCap)},
		{"seen_penalty_24h", fmt.Sprintf("-%f * LEAST(%s, 1)", v2.SeenPenaltyRecent, seenRecentSub)},
		{"hidden_author_penalty", fmt.Sprintf("-%f * LEAST(%s, 2)", v2.HideAuthorPenalty, hiddenAuthorSub)},
	}
}

// mainScoreArgs trả placeholder theo ĐÚNG thứ tự xuất hiện trong mainScoreTerms.
// Lệch thứ tự sẽ sinh lỗi MySQL 1064 hoặc trừ điểm sai bài — test
// TestMainScoreExpr_PlaceholderCount chặn tái phát.
func mainScoreArgs(snapshot time.Time, userID string) []interface{} {
	return []interface{}{
		snapshot,               // recency
		userID, userID,         // affinity: friend
		userID, userID, userID, // affinity: đã tương tác
		snapshot, userID,       // interest: snapshot trước, rồi user
		snapshot, snapshot,     // new post boost (age xuất hiện 2 lần)
		userID, snapshot,       // seen 7 ngày
		userID, snapshot,       // seen 24h
		userID,                 // hidden author
	}
}

// FetchPersonalized chạy 3 stream (main/interests, explore user nhỏ, explore chung)
// rồi merge theo slot explore + diversity cap. Trả về trang đã merge,
// cursor đã tiến và hasMore.
func (r *PostRepository) FetchPersonalized(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, filterFollowing bool, heroRound int) ([]models.Post, PersonalCursor, bool, error) {
	w := config.DefaultFeedWeights
	v2 := config.DefaultFeedRankV2

	uid := userID
	if uid == "" {
		uid = "-"
	}

	// ── Stream MAIN: full scan có recall-bound, rank cá nhân hóa ──
	// Stream đã cạn ở trang trước (Done) thì bỏ qua, tránh đọc lại từ đầu.
	mainReq := limit + 8
	var main []models.Post
	if !cursor.MainDone {
		var err error
		main, err = r.fetchMainStream(ctx, uid, mainReq, cursor, snapshotTime, filterFollowing, w, v2)
		if err != nil {
			return nil, cursor, false, err
		}
	}

	// ── Stream explore (tắt ở tab following — user chỉ muốn followings) ──
	var small, general []models.Post
	smallReq, genReq := 0, 0
	if !filterFollowing {
		smallReq = limit/2 + 4
		genReq = limit/2 + 4
		var err error
		if !cursor.SmallDone {
			if small, err = r.fetchSmallExploreStream(ctx, uid, smallReq, cursor, snapshotTime, v2); err != nil {
				return nil, cursor, false, err
			}
		}
		if !cursor.GenDone {
			if general, err = r.fetchGeneralExploreStream(ctx, uid, genReq, cursor, snapshotTime, v2); err != nil {
				return nil, cursor, false, err
			}
		}
	}

	merged, advM, advS, advG, moreM, moreS, moreG := MergeFeedStreams(
		main, small, general, mainReq, smallReq, genReq, limit, v2.ExploreEveryN, v2.MaxPerAuthor)

	// Rerank sau merge — chỉ đổi thứ tự trong trang, không đổi tập bài đã
	// tiêu thụ, nên cursor vẫn trỏ đúng chỗ cho trang sau.
	merged = SpreadAuthors(merged, v2.SpreadWindow)
	if !filterFollowing {
		merged = EnsureExploreInTop(merged, v2.ExploreTopSlots)
	}
	if heroRound > 0 {
		merged = RotateHeroTop(merged, heroRound, v2.HeroRotationSlots)
	}

	next := PersonalCursor{}
	if moreM && advM > 0 {
		last := main[advM-1]
		next.MainScore, next.MainID, next.MainOK = last.FeedScore, last.ID, true
	} else if !moreM {
		next.MainDone = true
	}
	if moreS && advS > 0 {
		last := small[advS-1]
		next.SmallNano, next.SmallID, next.SmallOK = last.CreatedAt.UnixNano(), last.ID, true
	} else if !moreS {
		next.SmallDone = true
	}
	if moreG && advG > 0 {
		last := general[advG-1]
		next.GenScore, next.GenID, next.GenOK = last.FeedScore, last.ID, true
	} else if !moreG {
		next.GenDone = true
	}
	return merged, next, moreM || moreS || moreG, nil
}

func (r *PostRepository) fetchMainStream(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, filterFollowing bool, w config.FeedWeights, v2 config.FeedRankV2Config) ([]models.Post, error) {
	terms := mainScoreTerms(w, v2)
	scoreExpr := "(" + joinScoreTerms(terms) + ") AS feed_score"

	likesSubQuery := r.db.Table("post_reactions").Select("post_id, COUNT(*) AS likes").Group("post_id")
	commentsSubQuery := r.db.Table("comments").Select("post_id, COUNT(*) AS comments").Group("post_id")
	sharesSubQuery := r.db.Table("post_shares").Select("post_id, COUNT(*) AS shares").Group("post_id")

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(feedAuthorSelect+",\n"+scoreExpr, mainScoreArgs(snapshotTime, userID)...).
		Joins("LEFT JOIN (?) lr ON lr.post_id = posts.id", likesSubQuery).
		Joins("LEFT JOIN (?) cr ON cr.post_id = posts.id", commentsSubQuery).
		Joins("LEFT JOIN (?) sr ON sr.post_id = posts.id", sharesSubQuery)
	q = feedBaseJoins(q, &userID)
	q = applyFeedVisibility(q, userID, true)

	// Recall bound: chỉ quét bài trong N ngày gần nhất.
	q = q.Where("posts.created_at > DATE_SUB(?, INTERVAL ? DAY)", snapshotTime, v2.MainRecallDays)

	if filterFollowing {
		q = q.Where("f.follower_id IS NOT NULL")
	}
	if cursor.MainOK {
		q = q.Having("feed_score < ? OR (feed_score = ? AND posts.id < ?)",
			cursor.MainScore, cursor.MainScore, cursor.MainID)
	}
	q = q.Order("feed_score DESC, posts.id DESC").Limit(limit)

	var posts []models.Post
	if err := q.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// fetchSmallExploreStream: bài public mới của tác giả nhỏ (< ngưỡng followers),
// chưa xem hôm nay, cũ nhất trước (round-robin công bằng).
func (r *PostRepository) fetchSmallExploreStream(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, v2 config.FeedRankV2Config) ([]models.Post, error) {
	if limit < 1 {
		return nil, nil
	}
	q := r.db.WithContext(ctx).
		Table("posts").
		Select(feedAuthorSelect + ",\nUNIX_TIMESTAMP(posts.created_at) AS feed_score")
	q = feedBaseJoins(q, &userID)
	q = applyFeedVisibility(q, userID, false)
	q = q.Where("(SELECT COUNT(*) FROM follows WHERE following_id = posts.user_id) < ?", v2.SmallAuthorFollowers).
		Where("posts.created_at > DATE_SUB(?, INTERVAL ? HOUR)", snapshotTime, v2.ExploreMaxAgeHours).
		Where("posts.created_at <= ?", snapshotTime).
		Where("NOT EXISTS (SELECT 1 FROM post_views WHERE post_id = posts.id AND viewer_id = ? AND created_at > DATE_SUB(?, INTERVAL 1 DAY))", userID, snapshotTime)
	if cursor.SmallOK {
		var cursorTime time.Time
		// SmallNano lưu UnixNano của created_at.
		cursorTime = time.Unix(0, cursor.SmallNano)
		q = q.Where("(posts.created_at > ? OR (posts.created_at = ? AND posts.id > ?))", cursorTime, cursorTime, cursor.SmallID)
	}
	q = q.Order("posts.created_at ASC, posts.id ASC").Limit(limit)

	var posts []models.Post
	if err := q.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// fetchGeneralExploreStream: bài public ngoài follow-graph + bạn bè,
// sắp theo vận tốc tương tác (interactions/giờ).
func (r *PostRepository) fetchGeneralExploreStream(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, v2 config.FeedRankV2Config) ([]models.Post, error) {
	if limit < 1 {
		return nil, nil
	}
	ageH := ageHoursExpr
	velocityExpr := fmt.Sprintf(`((COALESCE(lr.likes, 0) + COALESCE(cr.comments, 0) * %f + COALESCE(sr.shares, 0) * %f) / GREATEST(%s, 0.5)) AS feed_score`,
		config.DefaultFeedWeights.CommentWeight, config.DefaultFeedWeights.ShareWeight, ageH)

	likesSubQuery := r.db.Table("post_reactions").Select("post_id, COUNT(*) AS likes").Group("post_id")
	commentsSubQuery := r.db.Table("comments").Select("post_id, COUNT(*) AS comments").Group("post_id")
	sharesSubQuery := r.db.Table("post_shares").Select("post_id, COUNT(*) AS shares").Group("post_id")

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(feedAuthorSelect+",\n"+velocityExpr, snapshotTime).
		Joins("LEFT JOIN (?) lr ON lr.post_id = posts.id", likesSubQuery).
		Joins("LEFT JOIN (?) cr ON cr.post_id = posts.id", commentsSubQuery).
		Joins("LEFT JOIN (?) sr ON sr.post_id = posts.id", sharesSubQuery)
	q = feedBaseJoins(q, &userID)
	q = applyFeedVisibility(q, userID, false)
	q = q.Where("f.follower_id IS NULL").
		Where("NOT EXISTS (SELECT 1 FROM friends WHERE ((sender_id = ? AND receiver_id = posts.user_id) OR (sender_id = posts.user_id AND receiver_id = ?)) AND status = 'accepted')", userID, userID).
		Where("posts.created_at > DATE_SUB(?, INTERVAL ? HOUR)", snapshotTime, v2.ExploreMaxAgeHours).
		Where("posts.created_at <= ?", snapshotTime).
		Where("NOT EXISTS (SELECT 1 FROM post_views WHERE post_id = posts.id AND viewer_id = ? AND created_at > DATE_SUB(?, INTERVAL 1 DAY))", userID, snapshotTime)
	if cursor.GenOK {
		q = q.Having("feed_score < ? OR (feed_score = ? AND posts.id < ?)",
			cursor.GenScore, cursor.GenScore, cursor.GenID)
	}
	q = q.Order("feed_score DESC, posts.id DESC").Limit(limit)

	var posts []models.Post
	if err := q.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

// FeedMetrics là số liệu thô cho cổng rollout rank v2:
// v2 chỉ mở rộng khi small_author_exposure tăng và CTR/report-rate không xấu đi.
type FeedMetrics struct {
	TotalViews           int64
	FeedViews            int64
	DetailViews          int64
	SmallAuthorViews     int64
	SmallAuthorsSurfaced int64
	PostReports          int64

	// Từ v2.1: đo hiệu quả xoay vòng top slot và tín hiệu âm.
	ImpressionsTotal int64
	HiddenPosts      int64
	// TopSlotVariety = trung bình số bài khác nhau từng nằm ở vị trí #1 của
	// mỗi user trong khoảng thời gian. ~1 nghĩa là top đứng im.
	TopSlotVariety float64
}

// GetFeedMetrics tổng hợp lượt xem đã dedup từ post_views (kể từ mốc since).
func (r *PostRepository) GetFeedMetrics(ctx context.Context, since time.Time, smallFollowers int) (FeedMetrics, error) {
	var m FeedMetrics
	db := r.db.WithContext(ctx)

	smallAuthorCond := "pv.created_at > ? AND (SELECT COUNT(*) FROM follows WHERE following_id = p.user_id) < ?"

	queries := []struct {
		dest       *int64
		table      string
		joins      string
		where      string
		args       []interface{}
		selectExpr string
	}{
		{&m.TotalViews, "post_views", "", "created_at > ?", []interface{}{since}, ""},
		{&m.FeedViews, "post_views", "", "source = ? AND created_at > ?", []interface{}{string(models.PostViewSourceFeed), since}, ""},
		{&m.DetailViews, "post_views", "", "source = ? AND created_at > ?", []interface{}{string(models.PostViewSourceDetail), since}, ""},
		{&m.SmallAuthorViews, "post_views pv", "JOIN posts p ON p.id = pv.post_id", smallAuthorCond, []interface{}{since, smallFollowers}, ""},
		{&m.SmallAuthorsSurfaced, "post_views pv", "JOIN posts p ON p.id = pv.post_id", smallAuthorCond, []interface{}{since, smallFollowers}, "COUNT(DISTINCT p.user_id)"},
		{&m.PostReports, "reports", "", "target_post_id IS NOT NULL AND created_at > ?", []interface{}{since}, ""},
		{&m.ImpressionsTotal, "feed_impressions", "", "created_at > ?", []interface{}{since}, ""},
		{&m.HiddenPosts, "post_hidden", "", "created_at > ?", []interface{}{since}, ""},
	}

	for _, qq := range queries {
		q := db.Table(qq.table)
		if qq.joins != "" {
			q = q.Joins(qq.joins)
		}
		if qq.selectExpr != "" {
			if err := q.Where(qq.where, qq.args...).Select(qq.selectExpr).Scan(qq.dest).Error; err != nil {
				return m, err
			}
			continue
		}
		if err := q.Where(qq.where, qq.args...).Count(qq.dest).Error; err != nil {
			return m, err
		}
	}

	// Trung bình số bài khác nhau từng giữ vị trí #1 của mỗi user. Đây là chỉ
	// số trả lời trực tiếp "top slot có xoay không": 1.0 = đứng im.
	if err := db.Raw(`SELECT COALESCE(AVG(cnt), 0) FROM (
			SELECT viewer_id, COUNT(DISTINCT post_id) AS cnt
			FROM feed_impressions
			WHERE position = 1 AND created_at > ?
			GROUP BY viewer_id
		) t`, since).Scan(&m.TopSlotVariety).Error; err != nil {
		return m, err
	}
	return m, nil
}

// MergeFeedStreams gộp 3 stream thành 1 trang: mỗi slot thứ exploreEveryN
// dành cho explore (luân phiên small/general), kèm diversity cap mỗi tác giả.
// Trả về trang merged, số item đã tiêu thụ mỗi stream và cờ còn-dữ-liệu.
func MergeFeedStreams(main, small, general []models.Post, mainReq, smallReq, genReq, limit, exploreEveryN, maxPerAuthor int) (merged []models.Post, advM, advS, advG int, moreM, moreS, moreG bool) {
	if limit < 1 {
		return nil, 0, 0, 0, false, false, false
	}
	if exploreEveryN < 2 {
		exploreEveryN = 5
	}
	if maxPerAuthor < 1 {
		maxPerAuthor = 2
	}

	authorCount := map[string]int{}
	seen := map[string]bool{}
	exploreTurn := 0
	stuckRounds := 0

	takeable := func(s []models.Post, idx int) bool {
		return idx < len(s) && authorCount[s[idx].UserID] < maxPerAuthor && !seen[s[idx].ID]
	}

	take := func(s []models.Post, idx int, source string) models.Post {
		p := s[idx]
		p.FeedSource = source
		return p
	}

	for len(merged) < limit {
		// Bỏ qua item đã lấy từ stream khác trong cùng trang (main và explore
		// có thể trùng nhau) để không lặp bài.
		for advM < len(main) && seen[main[advM].ID] {
			advM++
		}
		for advS < len(small) && seen[small[advS].ID] {
			advS++
		}
		for advG < len(general) && seen[general[advG].ID] {
			advG++
		}

		isExploreSlot := (len(merged)+1)%exploreEveryN == 0
		progress := false

		if isExploreSlot {
			// Luân phiên small/general; stream cạn thì thử stream còn lại, rồi main.
			order := []int{0, 1}
			if exploreTurn%2 == 1 {
				order = []int{1, 0}
			}
			for _, which := range order {
				if which == 0 && takeable(small, advS) {
					merged = append(merged, take(small, advS, "explore"))
					authorCount[small[advS].UserID]++
					seen[small[advS].ID] = true
					advS++
					progress = true
					break
				}
				if which == 1 && takeable(general, advG) {
					merged = append(merged, take(general, advG, "explore"))
					authorCount[general[advG].UserID]++
					seen[general[advG].ID] = true
					advG++
					progress = true
					break
				}
			}
			// Explore slot nhưng cả 2 stream đều kẹt diversity → lấp bằng main.
			if !progress && takeable(main, advM) {
				merged = append(merged, take(main, advM, "main"))
				authorCount[main[advM].UserID]++
				seen[main[advM].ID] = true
				advM++
				progress = true
			}
			if progress {
				exploreTurn++
			}
		} else if takeable(main, advM) {
			merged = append(merged, take(main, advM, "main"))
			authorCount[main[advM].UserID]++
			seen[main[advM].ID] = true
			advM++
			progress = true
		} else {
			// Main kẹt (diversity/hết) → lấp bằng explore để đủ trang.
			if takeable(small, advS) {
				merged = append(merged, take(small, advS, "explore"))
				authorCount[small[advS].UserID]++
				seen[small[advS].ID] = true
				advS++
				progress = true
			} else if takeable(general, advG) {
				merged = append(merged, take(general, advG, "explore"))
				authorCount[general[advG].UserID]++
				seen[general[advG].ID] = true
				advG++
				progress = true
			}
		}

		if !progress {
			// Bỏ qua item kẹt diversity (tiêu thụ nhưng không lấy) để tránh kẹt cứng.
			skipped := false
			if advM < len(main) {
				advM++
				skipped = true
			} else if advS < len(small) {
				advS++
				skipped = true
			} else if advG < len(general) {
				advG++
				skipped = true
			}
			if !skipped {
				break
			}
			stuckRounds++
			if stuckRounds > limit+10 {
				break
			}
		} else {
			stuckRounds = 0
		}

		if advM >= len(main) && advS >= len(small) && advG >= len(general) {
			break
		}
	}

	moreM = advM < len(main) || len(main) == mainReq
	moreS = len(small) > 0 && (advS < len(small) || (smallReq > 0 && len(small) == smallReq))
	moreG = len(general) > 0 && (advG < len(general) || (genReq > 0 && len(general) == genReq))
	return merged, advM, advS, advG, moreM, moreS, moreG
}

// authorInWindow cho biết author đã xuất hiện trong posts[start:end] chưa.
func authorInWindow(posts []models.Post, start, end int, author string) bool {
	for i := start; i < end; i++ {
		if posts[i].UserID == author {
			return true
		}
	}
	return false
}

// SpreadAuthors trải bài cùng tác giả ra khỏi cửa sổ `window` bằng cách đổi
// chỗ với ứng viên gần nhất của tác giả khác. Khác MaxPerAuthor (cắt cứng),
// cách này giữ đủ số bài nhưng không để 2 bài cùng người nằm sát nhau.
func SpreadAuthors(posts []models.Post, window int) []models.Post {
	if window < 2 || len(posts) < 3 {
		return posts
	}
	out := make([]models.Post, len(posts))
	copy(out, posts)

	for i := 1; i < len(out); i++ {
		start := i - window + 1
		if start < 0 {
			start = 0
		}
		if !authorInWindow(out, start, i, out[i].UserID) {
			continue
		}
		limit := i + window
		if limit > len(out) {
			limit = len(out)
		}
		for k := i + 1; k < limit; k++ {
			if authorInWindow(out, start, i, out[k].UserID) {
				continue
			}
			out[i], out[k] = out[k], out[i]
			break
		}
	}
	return out
}

// EnsureExploreInTop bảo đảm trong `slots` vị trí đầu có ít nhất một bài
// ngoài follow-graph, bằng cách kéo bài explore gần nhất lên slot cuối.
// Mặc định tắt (ExploreTopSlots=0) vì với dữ liệu hiện tại explore yếu hơn
// top main khá nhiều — chỉ bật khi chạy thử nghiệm discovery.
func EnsureExploreInTop(posts []models.Post, slots int) []models.Post {
	if slots < 1 || len(posts) <= slots {
		return posts
	}
	for i := 0; i < slots; i++ {
		if posts[i].FeedSource == "explore" {
			return posts
		}
	}
	idx := -1
	for i := slots; i < len(posts); i++ {
		if posts[i].FeedSource == "explore" {
			idx = i
			break
		}
	}
	if idx < 0 {
		return posts
	}
	out := make([]models.Post, len(posts))
	copy(out, posts)
	hero := out[idx]
	copy(out[slots:idx+1], out[slots-1:idx])
	out[slots-1] = hero
	return out
}

// RotateHeroTop xoay slot #1 trong `slots` bài đầu theo vòng đếm được:
// round=1 → #1, round=2 → #2, round=3 → #3, round=4 → #1...
// Tất định, không random; chỉ đổi thứ tự hiển thị, không đổi tập bài của trang.
func RotateHeroTop(posts []models.Post, round, slots int) []models.Post {
	if round <= 0 || len(posts) < 2 {
		return posts
	}
	k := slots
	if k > len(posts) {
		k = len(posts)
	}
	if k < 2 {
		return posts
	}
	idx := (round - 1) % k
	if idx == 0 {
		return posts
	}
	out := make([]models.Post, len(posts))
	copy(out, posts)
	out[0], out[idx] = out[idx], out[0]
	return out
}

// NextHeroRound tăng và trả về vòng xoay slot #1 của user.
// State nằm trong MySQL (không phải RAM) nên nhiều instance vẫn nhất quán.
func (r *PostRepository) NextHeroRound(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, nil
	}
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Exec(
		`INSERT INTO feed_hero_state (user_id, hero_round, updated_at) VALUES (?, 1, ?)
		 ON DUPLICATE KEY UPDATE hero_round = hero_round + 1, updated_at = VALUES(updated_at)`,
		userID, now).Error; err != nil {
		return 0, err
	}
	var round int
	if err := r.db.WithContext(ctx).
		Raw(`SELECT hero_round FROM feed_hero_state WHERE user_id = ?`, userID).
		Scan(&round).Error; err != nil {
		return 0, err
	}
	return round, nil
}

// LogFeedImpressions ghi log bài được trả ở trang 1 (không dedup theo ngày —
// khác post_views). Dùng để đo mức độ xoay vòng top slot.
func (r *PostRepository) LogFeedImpressions(ctx context.Context, rows []models.FeedImpression) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, 50).Error
}

// DebugScoreBreakdown tính lại điểm từng thành phần cho các bài đã trả về.
// Chạy query riêng (chỉ khi FEED_DEBUG=true) để không đụng đường nóng.
func (r *PostRepository) DebugScoreBreakdown(ctx context.Context, userID string, postIDs []string, snapshot time.Time) (map[string]models.FeedDebugScore, error) {
	out := make(map[string]models.FeedDebugScore, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	terms := mainScoreTerms(config.DefaultFeedWeights, config.DefaultFeedRankV2)
	cols := make([]string, 0, len(terms)+1)
	cols = append(cols, "posts.id")
	for _, t := range terms {
		cols = append(cols, t.Expr+" AS dbg_"+t.Name)
	}

	likesSubQuery := r.db.Table("post_reactions").Select("post_id, COUNT(*) AS likes").Group("post_id")
	commentsSubQuery := r.db.Table("comments").Select("post_id, COUNT(*) AS comments").Group("post_id")
	sharesSubQuery := r.db.Table("post_shares").Select("post_id, COUNT(*) AS shares").Group("post_id")

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(strings.Join(cols, ",\n"), mainScoreArgs(snapshot, userID)...).
		Joins("LEFT JOIN (?) lr ON lr.post_id = posts.id", likesSubQuery).
		Joins("LEFT JOIN (?) cr ON cr.post_id = posts.id", commentsSubQuery).
		Joins("LEFT JOIN (?) sr ON sr.post_id = posts.id", sharesSubQuery)
	q = feedBaseJoins(q, &userID)
	q = q.Where("posts.id IN ?", postIDs)

	var rows []struct {
		ID                  string  `gorm:"column:id"`
		Recency             float64 `gorm:"column:dbg_recency"`
		Engagement          float64 `gorm:"column:dbg_engagement"`
		Affinity            float64 `gorm:"column:dbg_affinity"`
		Interest            float64 `gorm:"column:dbg_interest"`
		NewBoost            float64 `gorm:"column:dbg_new_boost"`
		SmallBonus          float64 `gorm:"column:dbg_small_bonus"`
		SeenPenalty         float64 `gorm:"column:dbg_seen_penalty"`
		SeenPenalty24h      float64 `gorm:"column:dbg_seen_penalty_24h"`
		HiddenAuthorPenalty float64 `gorm:"column:dbg_hidden_author_penalty"`
	}
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		d := models.FeedDebugScore{
			Recency:             row.Recency,
			Engagement:          row.Engagement,
			Affinity:            row.Affinity,
			Interest:            row.Interest,
			NewBoost:            row.NewBoost,
			SmallBonus:          row.SmallBonus,
			SeenPenalty:         row.SeenPenalty,
			SeenPenalty24h:      row.SeenPenalty24h,
			HiddenAuthorPenalty: row.HiddenAuthorPenalty,
		}
		d.Final = d.Recency + d.Engagement + d.Affinity + d.Interest + d.NewBoost + d.SmallBonus +
			d.SeenPenalty + d.SeenPenalty24h + d.HiddenAuthorPenalty
		out[row.ID] = d
	}
	return out, nil
}

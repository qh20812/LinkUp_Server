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
// Stream đã cạn dùng score/nano = -1 và id = "-".
type PersonalCursor struct {
	MainScore float64
	MainID    string
	MainOK    bool

	SmallNano int64
	SmallID   string
	SmallOK   bool

	GenScore float64
	GenID    string
	GenOK    bool
}

func (c PersonalCursor) Encode(snapshot time.Time) string {
	mScore, mID := -1.0, "-"
	if c.MainOK {
		mScore, mID = c.MainScore, c.MainID
	}
	sNano, sID := int64(-1), "-"
	if c.SmallOK {
		sNano, sID = c.SmallNano, c.SmallID
	}
	gScore, gID := -1.0, "-"
	if c.GenOK {
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
	if parts[3] != "-" && mScore >= 0 {
		c.MainScore, c.MainID, c.MainOK = mScore, parts[3], true
	}
	if parts[5] != "-" && sNano >= 0 {
		c.SmallNano, c.SmallID, c.SmallOK = sNano, parts[5], true
	}
	if parts[7] != "-" && gScore >= 0 {
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
		Where("NOT EXISTS (SELECT 1 FROM blocks WHERE user_id = posts.user_id AND blocked_user_id = ?)", userID)
}

func feedBaseJoins(q *gorm.DB, userID *string) *gorm.DB {
	return q.Joins("LEFT JOIN users ON users.id = posts.user_id").
		Joins("LEFT JOIN profiles ON profiles.user_id = posts.user_id").
		Joins("LEFT JOIN follows f ON f.following_id = posts.user_id AND f.follower_id = ?", userID)
}

// FetchPersonalized chạy 3 stream (main/interests, explore user nhỏ, explore chung)
// rồi merge theo slot explore + diversity cap. Trả về trang đã merge,
// cursor đã tiến và hasMore.
func (r *PostRepository) FetchPersonalized(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, filterFollowing bool) ([]models.Post, PersonalCursor, bool, error) {
	w := config.DefaultFeedWeights
	v2 := config.DefaultFeedRankV2

	uid := userID
	if uid == "" {
		uid = "-"
	}

	// ── Stream MAIN: full scan có recall-bound, rank cá nhân hóa ──
	mainReq := limit + 8
	main, err := r.fetchMainStream(ctx, uid, mainReq, cursor, snapshotTime, filterFollowing, w, v2)
	if err != nil {
		return nil, cursor, false, err
	}

	// ── Stream explore (tắt ở tab following — user chỉ muốn followings) ──
	var small, general []models.Post
	smallReq, genReq := 0, 0
	if !filterFollowing {
		smallReq = limit/2 + 4
		genReq = limit/2 + 4
		if small, err = r.fetchSmallExploreStream(ctx, uid, smallReq, cursor, snapshotTime, v2); err != nil {
			return nil, cursor, false, err
		}
		if general, err = r.fetchGeneralExploreStream(ctx, uid, genReq, cursor, snapshotTime, v2); err != nil {
			return nil, cursor, false, err
		}
	}

	merged, advM, advS, advG, moreM, moreS, moreG := MergeFeedStreams(
		main, small, general, mainReq, smallReq, genReq, limit, v2.ExploreEveryN, v2.MaxPerAuthor)

	next := PersonalCursor{}
	if moreM && advM > 0 {
		last := main[advM-1]
		next.MainScore, next.MainID, next.MainOK = last.FeedScore, last.ID, true
	}
	if moreS && advS > 0 {
		last := small[advS-1]
		next.SmallNano, next.SmallID, next.SmallOK = last.CreatedAt.UnixNano(), last.ID, true
	}
	if moreG && advG > 0 {
		last := general[advG-1]
		next.GenScore, next.GenID, next.GenOK = last.FeedScore, last.ID, true
	}
	return merged, next, moreM || moreS || moreG, nil
}

func (r *PostRepository) fetchMainStream(ctx context.Context, userID string, limit int, cursor PersonalCursor, snapshotTime time.Time, filterFollowing bool, w config.FeedWeights, v2 config.FeedRankV2Config) ([]models.Post, error) {
	ageH := "((UNIX_TIMESTAMP(?) - UNIX_TIMESTAMP(posts.created_at)) / 3600.0)"

	interestSub := fmt.Sprintf(`(SELECT COALESCE(SUM(ui.score * EXP(-%f * ((UNIX_TIMESTAMP(NOW()) - UNIX_TIMESTAMP(ui.updated_at)) / 86400.0))), 0)
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

	newBoost := fmt.Sprintf(`CASE WHEN (%s) < %f THEN %f * (1 - (%s) / %f) ELSE 0 END`,
		fmt.Sprintf(ageH, "?"), v2.NewPostBoostHours, v2.NewPostBoost, fmt.Sprintf(ageH, "?"), v2.NewPostBoostHours)

	smallBonus := fmt.Sprintf(`CASE WHEN (SELECT COUNT(*) FROM follows WHERE following_id = posts.user_id) < %d THEN %f ELSE 0 END`,
		v2.SmallAuthorFollowers, v2.SmallAuthorBonus)

	scoreExpr := fmt.Sprintf(`(%f * EXP(-%f * %s) +
		%f * (LOG(1 + COALESCE(lr.likes, 0)) * %f + LOG(1 + COALESCE(cr.comments, 0)) * %f + LOG(1 + COALESCE(sr.shares, 0)) * %f) / 20.0 +
		%f * (%s) +
		%f * LEAST(%s, 10) / 10 +
		%s -
		%f * LEAST(%s, %d) +
		%s) AS feed_score`,
		w.Recency, w.DecayRate, fmt.Sprintf(ageH, "?"),
		w.Engagement, w.LikeWeight, w.CommentWeight, w.ShareWeight,
		w.Affinity, affinityCase,
		v2.InterestWeight, interestSub,
		newBoost,
		v2.SeenPenalty, seenSub, v2.SeenCap,
		smallBonus)

	likesSubQuery := r.db.Table("post_reactions").Select("post_id, COUNT(*) AS likes").Group("post_id")
	commentsSubQuery := r.db.Table("comments").Select("post_id, COUNT(*) AS comments").Group("post_id")
	sharesSubQuery := r.db.Table("post_shares").Select("post_id, COUNT(*) AS shares").Group("post_id")

	q := r.db.WithContext(ctx).
		Table("posts").
		Select(feedAuthorSelect+",\n"+scoreExpr,
			snapshotTime,   // recency ageH
			userID, userID, // affinity friend
			userID, userID, userID, // affinity interacted
			userID,                     // interest subquery
			snapshotTime, snapshotTime, // new boost ageH x2
			userID, snapshotTime, // seen subquery
		).
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
	ageH := "((UNIX_TIMESTAMP(?) - UNIX_TIMESTAMP(posts.created_at)) / 3600.0)"
	velocityExpr := fmt.Sprintf(`((COALESCE(lr.likes, 0) + COALESCE(cr.comments, 0) * %f + COALESCE(sr.shares, 0) * %f) / GREATEST(%s, 0.5)) AS feed_score`,
		config.DefaultFeedWeights.CommentWeight, config.DefaultFeedWeights.ShareWeight, fmt.Sprintf(ageH, "?"))

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
	exploreTurn := 0
	stuckRounds := 0

	takeable := func(s []models.Post, idx int) bool {
		return idx < len(s) && authorCount[s[idx].UserID] < maxPerAuthor
	}

	for len(merged) < limit {
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
					merged = append(merged, small[advS])
					authorCount[small[advS].UserID]++
					advS++
					progress = true
					break
				}
				if which == 1 && takeable(general, advG) {
					merged = append(merged, general[advG])
					authorCount[general[advG].UserID]++
					advG++
					progress = true
					break
				}
			}
			// Explore slot nhưng cả 2 stream đều kẹt diversity → lấp bằng main.
			if !progress && takeable(main, advM) {
				merged = append(merged, main[advM])
				authorCount[main[advM].UserID]++
				advM++
				progress = true
			}
			if progress {
				exploreTurn++
			}
		} else if takeable(main, advM) {
			merged = append(merged, main[advM])
			authorCount[main[advM].UserID]++
			advM++
			progress = true
		} else {
			// Main kẹt (diversity/hết) → lấp bằng explore để đủ trang.
			if takeable(small, advS) {
				merged = append(merged, small[advS])
				authorCount[small[advS].UserID]++
				advS++
				progress = true
			} else if takeable(general, advG) {
				merged = append(merged, general[advG])
				authorCount[general[advG].UserID]++
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

package social

import (
	"fmt"
	"math/rand"
	"time"

	"linkup/cmd/seed/internal"
	"linkup/config"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randRange(min, max int) int {
	return rng.Intn(max-min+1) + min
}

func pick[T any](items []T) T {
	return items[rng.Intn(len(items))]
}

func Run(env config.Env, state *internal.SeedState) error {
	database, err := internal.Connect(env)
	if err != nil {
		return fmt.Errorf("social: connect: %w", err)
	}
	defer database.Close()

	now := time.Now().UTC()

	postTitles := []string{
		"My journey into open source",
		"Building scalable APIs with Go",
		"Top 10 VS Code extensions for 2026",
		"How I learned to stop worrying and love the cloud",
		"Understanding distributed systems",
		"The future of AI in everyday apps",
		"Clean code principles that matter",
		"From monolith to microservices",
		"Why recursive algorithms still matter",
		"DevOps culture: more than just tools",
		"Designing user-friendly interfaces",
		"The art of code review",
		"Database optimization techniques",
		"Securing your web application",
		"Containerization best practices",
		"REST vs GraphQL: choosing the right API",
		"Introduction to event-driven architecture",
		"Performance tuning for high-traffic sites",
		"Testing strategies for modern web apps",
		"What I wish I knew as a junior dev",
		"Building real-time features with WebSockets",
		"CI/CD pipelines that actually work",
		"Managing technical debt effectively",
		"Data structures for interviews",
		"Why documentation matters",
		"Accessibility in web development",
		"Edge computing and the IoT revolution",
		"Blockchain beyond cryptocurrency",
		"Serverless architecture trade-offs",
		"Zero-trust security model explained",
	}

	postContents := []string{
		"After years of working with proprietary software, I finally made the leap to open source. The community has been incredibly welcoming and I've learned more in the past month than in the entire last year. Here are my key takeaways...",
		"In this post I'll walk through the architecture of a high-performance API built with Go. We'll cover routing, middleware, database optimization, and deployment strategies that have worked well for our team.",
		"I've curated a list of extensions that have significantly boosted my productivity. From intelligent code completion to advanced debugging tools, these plugins transform VS Code into a powerhouse.",
		"Cloud computing doesn't have to be scary. Once you understand the fundamental concepts of scalability, redundancy, and cost optimization, you'll wonder why you didn't make the switch sooner.",
		"Distributed systems are everywhere, but understanding them can be daunting. Let's break down the core concepts: consistency, partitioning, replication, and fault tolerance.",
		"AI is no longer a futuristic concept — it's here and transforming how we build applications. From recommendation engines to natural language processing, the possibilities are endless.",
		"Good code writes itself when you follow solid principles. Let's explore SRP, OCP, LSP, ISP, DIP with real-world examples that you can apply to your projects today.",
		"Our journey from a monolithic Rails app to microservices taught us valuable lessons. Not everything needs to be a microservice, but when you get it right, the benefits are enormous.",
		"Recursion isn't just an academic concept. From tree traversals to divide-and-conquer algorithms, understanding recursion will make you a better problem solver.",
		"DevOps is a cultural shift, not a toolbox. The most successful transformations happen when teams embrace collaboration, automation, and continuous improvement.",
	}

	postBodies := []string{
		"\n\nI'll cover specific examples from my own experience and provide actionable advice for anyone considering the switch. The open source ecosystem has matured enormously and there's never been a better time to get involved.",
		"\n\nWe'll start with the basics of routing and middleware setup, then dive into more advanced topics like connection pooling, query optimization, and caching strategies.",
		"\n\nEach extension on this list has been thoroughly tested and vetted. I'll include configuration tips and potential gotchas to watch out for.",
		"\n\nThis guide covers the major cloud providers and their strengths, cost optimization strategies, security best practices, and how to choose the right services for your needs.",
		"\n\nUnderstanding the CAP theorem, consensus algorithms like Raft and Paxos, and when to use synchronous vs asynchronous replication are all crucial skills.",
		"\n\nWe'll look at practical applications of AI in everyday software: smart search, personalized recommendations, automated moderation, and predictive analytics.",
		"\n\nEach principle comes with code examples showing both violations and correct implementations. You'll walk away with practical patterns you can use immediately.",
		"\n\nKey lessons: start with a well-defined bounded context, invest in observability early, and never underestimate the complexity of data consistency across services.",
		"\n\nWe'll implement several classic recursive algorithms and analyze their space/time complexity. You'll see how recursion maps naturally to certain problem domains.",
		"\n\nPractical steps to build a DevOps culture: automate everything, blameless retrospectives, shared ownership of production, and continuous learning.",
	}

	for i := 0; i < 30; i++ {
		postID := internal.UUID()
		userID := state.UserIDs[i%len(state.UserIDs)]
		title := postTitles[i%len(postTitles)]
		body := postContents[i%len(postContents)]
		extra := postBodies[i%len(postBodies)]
		views := randRange(10, 5000)

		status := "active"
		if i%15 == 7 {
			status = "hidden"
		}

		createdAt := now.Add(-time.Duration(randRange(1, 720)) * time.Hour)

		if err := internal.Exec(database,
			`INSERT INTO posts (id, user_id, title, content, views_count, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
			postID, userID, title, body+extra, views, status, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert post %d: %w", i, err)
		}
		state.PostIDs = append(state.PostIDs, postID)
	}

	commentTexts := []string{
		"Great post! Really insightful.",
		"I completely agree with your points.",
		"Thanks for sharing this, very helpful.",
		"Have you considered the alternative approach?",
		"This changed my perspective, thank you!",
		"Can you elaborate more on this topic?",
		"Saved for later reading. Excellent content!",
		"I've been using this technique and it works great.",
		"Bookmarked! This is exactly what I needed.",
		"Interesting take, but I have a different opinion.",
		"Could you share some code examples?",
		"This is underrated, should have more visibility.",
		"Well written and easy to follow.",
		"I ran into this issue last week, perfect timing!",
		"Adding this to my learning roadmap.",
	}

	for i := 0; i < 60; i++ {
		commentID := internal.UUID()
		userID := state.UserIDs[i%len(state.UserIDs)]
		postID := state.PostIDs[i%len(state.PostIDs)]
		var parentID *string
		if i >= 30 && i%3 == 0 {
			parentID = internal.Ptr(state.CommentIDs[i%len(state.CommentIDs)])
		}

		createdAt := now.Add(-time.Duration(randRange(1, 360)) * time.Hour)
		content := pick(commentTexts)
		if parentID != nil {
			content = "Reply: " + content
		}

		if err := internal.Exec(database,
			`INSERT INTO comments (id, user_id, post_id, parent_id, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
			commentID, userID, postID, parentID, content, createdAt,
		); err != nil {
			return fmt.Errorf("social: insert comment %d: %w", i, err)
		}
		state.CommentIDs = append(state.CommentIDs, commentID)
	}

	for i := 0; i < 50; i++ {
		postID := state.PostIDs[i%len(state.PostIDs)]
		hashTags := []string{"golang", "programming", "devops", "cloud", "ai", "webdev", "security", "database", "microservices", "career"}
		tagType := "hashtag"
		var commentID *string
		targetUserID := (*string)(nil)
		name := pick(hashTags)

		if i >= 25 {
			tagType = "mention"
			commentID = internal.Ptr(state.CommentIDs[i%len(state.CommentIDs)])
			targetUserID = internal.Ptr(state.UserIDs[i%len(state.UserIDs)])
			name = "@" + pick([]string{"john_doe", "jane_smith", "alice_wonder", "bob_builder", "charlie_dev"})
		}

		if err := internal.Exec(database,
			`INSERT INTO tags (id, post_id, comment_id, tag_type, target_user_id, name) VALUES (?, ?, ?, ?, ?, ?)`,
			internal.UUID(), postID, commentID, tagType, targetUserID, name,
		); err != nil {
			return fmt.Errorf("social: insert tag %d: %w", i, err)
		}
	}

	reactionSeen := map[string]bool{}
	for i := 0; i < 80; i++ {
		userID := state.UserIDs[randRange(0, len(state.UserIDs)-1)]
		postID := state.PostIDs[randRange(0, len(state.PostIDs)-1)]
		emojiID := state.EmojiIDs[randRange(0, len(state.EmojiIDs)-1)]
		key := userID + "|" + postID + "|" + emojiID
		if reactionSeen[key] {
			continue
		}
		reactionSeen[key] = true

		if err := internal.Exec(database,
			`INSERT INTO post_reactions (id, user_id, post_id, emoji_id, created_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), userID, postID, emojiID, now.Add(-time.Duration(randRange(1, 168))*time.Hour),
		); err != nil {
			return fmt.Errorf("social: insert reaction %d: %w", i, err)
		}
	}

	followSeen := map[string]bool{}
	for i := 0; i < 40; i++ {
		followerIdx := randRange(0, len(state.UserIDs)-1)
		// Loại trừ superadmin (index 0) và admin (index 1) — không ai được follow họ
		followingIdx := randRange(2, len(state.UserIDs)-1)
		if followerIdx == followingIdx {
			continue
		}
		key := state.UserIDs[followerIdx] + "|" + state.UserIDs[followingIdx]
		if followSeen[key] {
			continue
		}
		followSeen[key] = true

		if err := internal.Exec(database,
			`INSERT INTO follows (id, follower_id, following_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), state.UserIDs[followerIdx], state.UserIDs[followingIdx], now.Add(-time.Duration(randRange(1, 720))*time.Hour),
		); err != nil {
			return fmt.Errorf("social: insert follow %d: %w", i, err)
		}
	}

	friendStatuses := []string{"pending", "accepted", "accepted", "accepted", "rejected"}
	friendSeen := map[string]bool{}
	for i := 0; i < 15; i++ {
		senderIdx := randRange(0, len(state.UserIDs)-1)
		receiverIdx := randRange(0, len(state.UserIDs)-1)
		if senderIdx == receiverIdx {
			continue
		}
		key := state.UserIDs[senderIdx] + "|" + state.UserIDs[receiverIdx]
		if friendSeen[key] {
			continue
		}
		friendSeen[key] = true

		if err := internal.Exec(database,
			`INSERT INTO friends (id, sender_id, receiver_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
			internal.UUID(), state.UserIDs[senderIdx], state.UserIDs[receiverIdx], pick(friendStatuses), now.Add(-time.Duration(randRange(1, 720))*time.Hour),
		); err != nil {
			return fmt.Errorf("social: insert friend %d: %w", i, err)
		}
	}

	blockSeen := map[string]bool{}
	for i := 0; i < 5; i++ {
		userIdx := randRange(0, len(state.UserIDs)-1)
		blockedIdx := randRange(0, len(state.UserIDs)-1)
		if userIdx == blockedIdx {
			continue
		}
		key := state.UserIDs[userIdx] + "|" + state.UserIDs[blockedIdx]
		if blockSeen[key] {
			continue
		}
		blockSeen[key] = true

		if err := internal.Exec(database,
			`INSERT INTO blocks (id, user_id, blocked_user_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), state.UserIDs[userIdx], state.UserIDs[blockedIdx], now.Add(-time.Duration(randRange(1, 720))*time.Hour),
		); err != nil {
			return fmt.Errorf("social: insert block %d: %w", i, err)
		}
	}

	bookmarkSeen := map[string]bool{}
	for i := 0; i < 20; i++ {
		userID := state.UserIDs[randRange(0, len(state.UserIDs)-1)]
		postID := state.PostIDs[randRange(0, len(state.PostIDs)-1)]
		key := userID + "|" + postID
		if bookmarkSeen[key] {
			continue
		}
		bookmarkSeen[key] = true

		if err := internal.Exec(database,
			`INSERT INTO bookmarks (id, user_id, post_id, created_at) VALUES (?, ?, ?, ?)`,
			internal.UUID(), userID, postID, now.Add(-time.Duration(randRange(1, 168))*time.Hour),
		); err != nil {
			return fmt.Errorf("social: insert bookmark %d: %w", i, err)
		}
	}

	return nil
}

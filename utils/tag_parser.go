package utils

import (
	"regexp"
	"strings"
)

// ExtractHashtags quét nội dung và lấy ra danh sách các hashtag không trùng lặp
func ExtractHashtags(content string) []string {
	// \p{L}: Khớp với mọi chữ cái (Hỗ trợ tiếng Việt có dấu)
	// \p{N}: Khớp với mọi chữ số
	re := regexp.MustCompile(`#([\p{L}\p{N}_]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	uniqueTags := make(map[string]bool)
	var result []string

	for _, match := range matches {
		if len(match) > 1 {
			tag := strings.ToLower(match[1])
			if !uniqueTags[tag] {
				uniqueTags[tag] = true
				result = append(result, tag)
			}
		}
	}
	return result
}

// ExtractMentions quét nội dung và lấy ra danh sách các username được tag không trùng lặp
func ExtractMentions(content string) []string {
	re := regexp.MustCompile(`@([\p{L}\p{N}_]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	uniqueMentions := make(map[string]bool)
	var result []string

	for _, match := range matches {
		if len(match) > 1 {
			mention := strings.ToLower(match[1])
			if !uniqueMentions[mention] {
				uniqueMentions[mention] = true
				result = append(result, mention)
			}
		}
	}
	return result
}

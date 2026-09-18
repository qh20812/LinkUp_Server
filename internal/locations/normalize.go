package locations

import "strings"

// viBaseRunes maps a lowercase Vietnamese character with diacritics to its
// plain (base) Latin letter. "đ" maps to "d" (it is part of the alphabet).
var viBaseRunes = map[rune]rune{
	'à': 'a', 'á': 'a', 'ả': 'a', 'ã': 'a', 'ạ': 'a',
	'ă': 'a', 'ằ': 'a', 'ắ': 'a', 'ẳ': 'a', 'ẵ': 'a', 'ặ': 'a',
	'â': 'a', 'ầ': 'a', 'ấ': 'a', 'ẩ': 'a', 'ẫ': 'a', 'ậ': 'a',
	'è': 'e', 'é': 'e', 'ẻ': 'e', 'ẽ': 'e', 'ẹ': 'e',
	'ê': 'e', 'ề': 'e', 'ế': 'e', 'ể': 'e', 'ễ': 'e', 'ệ': 'e',
	'ì': 'i', 'í': 'i', 'ỉ': 'i', 'ĩ': 'i', 'ị': 'i',
	'ò': 'o', 'ó': 'o', 'ỏ': 'o', 'õ': 'o', 'ọ': 'o',
	'ô': 'o', 'ồ': 'o', 'ố': 'o', 'ổ': 'o', 'ỗ': 'o', 'ộ': 'o',
	'ơ': 'o', 'ờ': 'o', 'ớ': 'o', 'ở': 'o', 'ỡ': 'o', 'ợ': 'o',
	'ù': 'u', 'ú': 'u', 'ủ': 'u', 'ũ': 'u', 'ụ': 'u',
	'ư': 'u', 'ừ': 'u', 'ứ': 'u', 'ử': 'u', 'ữ': 'u', 'ự': 'u',
	'ỳ': 'y', 'ý': 'y', 'ỷ': 'y', 'ỹ': 'y', 'ỵ': 'y',
	'đ': 'd',
}

// Normalize strips Vietnamese diacritics, lowercases, collapses separators
// (space / hyphen / underscore / apostrophe / period) into a single space, and
// trims. The result is ASCII-only, so byte slicing is safe on its output.
func Normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := true
	for _, r := range s {
		if base, ok := viBaseRunes[r]; ok {
			r = base
		}
		if r == ' ' || r == '-' || r == '_' || r == '\'' || r == '.' {
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		b.WriteRune(r)
		lastSpace = false
	}
	return strings.TrimSpace(b.String())
}

// unitPrefixes are Vietnamese admin-level words (normalized) that may prefix a
// place name returned by a geocoder (e.g. "Phường Ba Đình").
var unitPrefixes = []string{
	"thanh pho", "thanhpho", "tp",
	"thi tran", "thitran", "thi xa", "thixa",
	"phuong", "xa", "huyen", "quan", "tinh", "pho",
}

// baseNormalized strips a leading admin-level prefix from a raw name and
// returns the remaining normalized core ("" if nothing remains).
func baseNormalized(raw string) string {
	n := Normalize(raw)
	for _, p := range unitPrefixes {
		if n == p {
			return ""
		}
		if strings.HasPrefix(n, p+" ") {
			return strings.TrimSpace(n[len(p):])
		}
	}
	return n
}

// matchName reports whether a raw name (optionally with an admin prefix, e.g.
// "Phường Ba Đình") matches a stored place name (vi or en).
func matchName(storedVi, storedEn, raw string) bool {
	n := baseNormalized(raw)
	if n == "" {
		return false
	}
	return Normalize(storedVi) == n || Normalize(storedEn) == n
}
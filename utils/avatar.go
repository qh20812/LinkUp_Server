package utils

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var avatarColors = []string{
	"#4F46E5",
	"#7C3AED",
	"#2563EB",
	"#059669",
	"#D97706",
	"#DC2626",
	"#DB2777",
	"#0891B2",
}

func extractInitials(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "?"
	}
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		first := []rune(parts[0])
		last := []rune(parts[len(parts)-1])
		return strings.ToUpper(string([]rune{first[0], last[0]}))
	}
	r := []rune(strings.ToUpper(name))
	if len(r) >= 2 {
		return string(r[:2])
	}
	return string(r)
}

func pickColor(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return avatarColors[h.Sum32()%uint32(len(avatarColors))]
}

func GenerateInitialsAvatar(name string) string {
	initials := extractInitials(name)
	color := pickColor(name)

	return fmt.Sprintf(`<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <circle cx="50" cy="50" r="50" fill="%s"/>
  <text x="50" y="55" text-anchor="middle" dominant-baseline="middle" fill="white" font-size="40" font-family="sans-serif" font-weight="600">%s</text>
</svg>`, color, initials)
}

func GenerateAndUploadAvatar(cloudinaryURL, name string) (string, error) {
	svg := GenerateInitialsAvatar(name)

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return "", fmt.Errorf("init cloudinary: %w", err)
	}

	publicID := "avatar-" + GenerateUUID()
	result, err := cld.Upload.Upload(context.Background(), bytes.NewBufferString(svg), uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "avatars/defaults",
		ResourceType:   "image",
		Format:         "svg",
		Transformation: "w_200,h_200,c_fill",
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar to cloudinary: %w", err)
	}

	return result.SecureURL, nil
}

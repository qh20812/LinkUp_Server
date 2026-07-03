package validations

import (
	"errors"
	"linkup/dto"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrPostWeightInvalid      = errors.New("trọng số bài viết phải từ 0 đến 100")
	ErrCommentWeightInvalid   = errors.New("trọng số bình luận phải từ 0 đến 100")
	ErrReactionWeightInvalid  = errors.New("trọng số phản hồi phải từ 0 đến 100")
	ErrEventWeightInvalid     = errors.New("trọng số sự kiện phải từ 0 đến 100")
	ErrThresholdInvalid       = errors.New("ngưỡng điểm phải lớn hơn 0")
	ErrThresholdOrderInvalid  = errors.New("ngưỡng Moderator phải lớn hơn ngưỡng Top Contributor")
	ErrHashtagRequired        = errors.New("hashtag là bắt buộc")
	ErrHashtagInvalidFormat   = errors.New("hashtag phải bắt đầu bằng #")
	ErrEndDateBeforeStart     = errors.New("ngày kết thúc phải sau ngày bắt đầu")
	ErrChallengeTitleRequired = errors.New("tên challenge là bắt buộc")
	ErrChallengeTitleTooShort = errors.New("tên challenge phải có ít nhất 5 ký tự")
	ErrChallengeTitleTooLong  = errors.New("tên challenge không được vượt quá 255 ký tự")
	ErrDescriptionTooLong     = errors.New("mô tả challenge không được vượt quá 2000 ký tự")
	ErrDateFormatInvalid      = errors.New("định dạng ngày không hợp lệ, cần dùng RFC3339")
)

type ContributionValidation struct{}

func NewContributionValidation() *ContributionValidation {
	return &ContributionValidation{}
}

func (v *ContributionValidation) ValidatePolicyInput(input dto.CreatePolicyInput) error {
	if input.PostWeight < 0 || input.PostWeight > 100 {
		return ErrPostWeightInvalid
	}
	if input.CommentWeight < 0 || input.CommentWeight > 100 {
		return ErrCommentWeightInvalid
	}
	if input.ReactionWeight < 0 || input.ReactionWeight > 100 {
		return ErrReactionWeightInvalid
	}
	if input.EventWeight < 0 || input.EventWeight > 100 {
		return ErrEventWeightInvalid
	}
	if input.TopContributorThreshold <= 0 || input.ModeratorPromotionThreshold <= 0 {
		return ErrThresholdInvalid
	}
	if input.ModeratorPromotionThreshold <= input.TopContributorThreshold {
		return ErrThresholdOrderInvalid
	}
	return nil
}

func (v *ContributionValidation) ValidateUpdatePolicyInput(input dto.UpdatePolicyInput) error {
	return v.ValidatePolicyInput(dto.CreatePolicyInput{
		PostWeight:                  input.PostWeight,
		CommentWeight:               input.CommentWeight,
		ReactionWeight:              input.ReactionWeight,
		EventWeight:                 input.EventWeight,
		TopContributorThreshold:     input.TopContributorThreshold,
		ModeratorPromotionThreshold: input.ModeratorPromotionThreshold,
		AutoPromoteEnabled:          input.AutoPromoteEnabled,
		BadgeEnabled:                input.BadgeEnabled,
	})
}

func (v *ContributionValidation) ValidateCreateChallenge(input dto.CreateChallengeInput) error {
	if err := v.ValidateChallengeTitle(input.Title); err != nil {
		return err
	}
	if err := v.ValidateChallengeDescription(input.Description); err != nil {
		return err
	}
	if err := v.ValidateHashtag(input.Hashtag); err != nil {
		return err
	}
	if err := v.ValidateChallengeDates(input.StartDate, input.EndDate); err != nil {
		return err
	}
	return nil
}

func (v *ContributionValidation) ValidateChallengeTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrChallengeTitleRequired
	}
	if utf8.RuneCountInString(title) < 5 {
		return ErrChallengeTitleTooShort
	}
	if utf8.RuneCountInString(title) > 255 {
		return ErrChallengeTitleTooLong
	}
	return nil
}

func (v *ContributionValidation) ValidateChallengeDescription(description string) error {
	if description == "" {
		return nil
	}
	if utf8.RuneCountInString(description) > 2000 {
		return ErrDescriptionTooLong
	}
	return nil
}

func (v *ContributionValidation) ValidateHashtag(hashtag string) error {
	hashtag = strings.TrimSpace(hashtag)
	if hashtag == "" {
		return ErrHashtagRequired
	}
	if !strings.HasPrefix(hashtag, "#") {
		return ErrHashtagInvalidFormat
	}
	return nil
}

func (v *ContributionValidation) ValidateChallengeDates(startDate, endDate string) error {
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(startDate))
	if err != nil {
		return ErrDateFormatInvalid
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(endDate))
	if err != nil {
		return ErrDateFormatInvalid
	}
	if !end.After(start) {
		return ErrEndDateBeforeStart
	}
	return nil
}

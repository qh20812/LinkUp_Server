package validations

import (
	"linkup/dto"
	errorsapp "linkup/errors"
	"strings"
	"time"
	"unicode/utf8"
)

type ContributionValidation struct{}

func NewContributionValidation() *ContributionValidation {
	return &ContributionValidation{}
}

func (v *ContributionValidation) NormalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return page, pageSize
}

func (v *ContributionValidation) ValidatePolicyInput(input dto.CreatePolicyInput) error {
	if input.PostWeight < 0 || input.PostWeight > 100 {
		return errorsapp.New(errorsapp.ErrCodeContribPostWeightInvalid)
	}
	if input.CommentWeight < 0 || input.CommentWeight > 100 {
		return errorsapp.New(errorsapp.ErrCodeContribCommentWeightInvalid)
	}
	if input.ReactionWeight < 0 || input.ReactionWeight > 100 {
		return errorsapp.New(errorsapp.ErrCodeContribReactionWeightInvalid)
	}
	if input.EventWeight < 0 || input.EventWeight > 100 {
		return errorsapp.New(errorsapp.ErrCodeContribEventWeightInvalid)
	}
	if input.TopContributorThreshold <= 0 || input.ModeratorPromotionThreshold <= 0 {
		return errorsapp.New(errorsapp.ErrCodeContribThresholdInvalid)
	}
	if input.ModeratorPromotionThreshold <= input.TopContributorThreshold {
		return errorsapp.New(errorsapp.ErrCodeContribThresholdOrderInvalid)
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

func (v *ContributionValidation) ValidateCreateChallenge(input dto.CreateChallengeInput) (time.Time, time.Time, error) {
	if err := v.ValidateChallengeTitle(input.Title); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if err := v.ValidateChallengeDescription(input.Description); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if err := v.ValidateHashtag(input.Hashtag); err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, end, err := v.ValidateChallengeDates(input.StartDate, input.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}

func (v *ContributionValidation) ValidateChallengeTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errorsapp.New(errorsapp.ErrCodeContribTitleRequired)
	}
	if utf8.RuneCountInString(title) < 5 {
		return errorsapp.New(errorsapp.ErrCodeContribTitleTooShort)
	}
	if utf8.RuneCountInString(title) > 255 {
		return errorsapp.New(errorsapp.ErrCodeContribTitleTooLong)
	}
	return nil
}

func (v *ContributionValidation) ValidateChallengeDescription(description string) error {
	if description == "" {
		return nil
	}
	if utf8.RuneCountInString(description) > 2000 {
		return errorsapp.New(errorsapp.ErrCodeContribDescTooLong)
	}
	return nil
}

func (v *ContributionValidation) ValidateHashtag(hashtag string) error {
	hashtag = strings.TrimSpace(hashtag)
	if hashtag == "" {
		return errorsapp.New(errorsapp.ErrCodeContribHashtagRequired)
	}
	if !strings.HasPrefix(hashtag, "#") {
		return errorsapp.New(errorsapp.ErrCodeContribHashtagInvalidFormat)
	}
	return nil
}

func (v *ContributionValidation) ValidateChallengeDates(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(startDate))
	if err != nil {
		return time.Time{}, time.Time{}, errorsapp.New(errorsapp.ErrCodeContribDateFormatInvalid)
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(endDate))
	if err != nil {
		return time.Time{}, time.Time{}, errorsapp.New(errorsapp.ErrCodeContribDateFormatInvalid)
	}
	if start.Before(time.Now()) {
		return time.Time{}, time.Time{}, errorsapp.New(errorsapp.ErrCodeContribStartDateInPast)
	}
	if !end.After(start) {
		return time.Time{}, time.Time{}, errorsapp.New(errorsapp.ErrCodeContribEndDateBeforeStart)
	}
	return start, end, nil
}

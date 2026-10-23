package contribution_test

import (
	"testing"
	"time"

	"linkup/dto"
	"linkup/validations"
)

func TestValidatePolicyInput(t *testing.T) {
	v := validations.NewContributionValidation()

	tests := []struct {
		name    string
		input   dto.CreatePolicyInput
		wantErr string
	}{
		{
			name: "valid policy",
			input: dto.CreatePolicyInput{
				PostWeight:                  10,
				CommentWeight:               5,
				ReactionWeight:              2,
				EventWeight:                 20,
				TopContributorThreshold:     2500,
				ModeratorPromotionThreshold: 5000,
			},
			wantErr: "",
		},
		{
			name: "post weight too high",
			input: dto.CreatePolicyInput{
				PostWeight:                  101,
				CommentWeight:               5,
				ReactionWeight:              2,
				EventWeight:                 20,
				TopContributorThreshold:     2500,
				ModeratorPromotionThreshold: 5000,
			},
			wantErr: "Trọng số bài viết phải từ 0 đến 100",
		},
		{
			name: "threshold order invalid",
			input: dto.CreatePolicyInput{
				PostWeight:                  10,
				CommentWeight:               5,
				ReactionWeight:              2,
				EventWeight:                 20,
				TopContributorThreshold:     5000,
				ModeratorPromotionThreshold: 2500,
			},
			wantErr: "Ngưỡng Moderator phải lớn hơn ngưỡng Top Contributor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePolicyInput(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateCreateChallenge(t *testing.T) {
	v := validations.NewContributionValidation()

	// Dates relative to now so the "must be in the future" rule never goes stale.
	start := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	end := time.Now().Add(48 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		input   dto.CreateChallengeInput
		wantErr string
	}{
		{
			name: "valid challenge",
			input: dto.CreateChallengeInput{
				Title:           "Photo Challenge",
				Description:     "Share your best photo",
				Hashtag:         "#LinkUpPhoto",
				PointsPerPost:   15,
				StartDate:       start,
				EndDate:         end,
				MaxParticipants: nil,
			},
			wantErr: "",
		},
		{
			name: "missing hashtag prefix",
			input: dto.CreateChallengeInput{
				Title:         "Photo Challenge",
				Description:   "Share your best photo",
				Hashtag:       "LinkUpPhoto",
				PointsPerPost: 15,
				StartDate:     start,
				EndDate:       end,
			},
			wantErr: "Hashtag phải bắt đầu bằng #",
		},
		{
			name: "end before start",
			input: dto.CreateChallengeInput{
				Title:         "Photo Challenge",
				Description:   "Share your best photo",
				Hashtag:       "#LinkUpPhoto",
				PointsPerPost: 15,
				StartDate:     time.Now().Add(48 * time.Hour).Format(time.RFC3339),
				EndDate:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			},
			wantErr: "Ngày kết thúc phải sau ngày bắt đầu",
		},
		{
			name: "points per post out of range",
			input: dto.CreateChallengeInput{
				Title:         "Photo Challenge",
				Description:   "Share your best photo",
				Hashtag:       "#LinkUpPhoto",
				PointsPerPost: 0,
				StartDate:     start,
				EndDate:       end,
			},
			wantErr: "Điểm mỗi bài phải từ 1 đến 1000",
		},
		{
			name: "max participants invalid",
			input: dto.CreateChallengeInput{
				Title:           "Photo Challenge",
				Description:     "Share your best photo",
				Hashtag:         "#LinkUpPhoto",
				PointsPerPost:   15,
				StartDate:       start,
				EndDate:         end,
				MaxParticipants: intPtr(0),
			},
			wantErr: "Số người tham gia tối đa phải từ 1 đến 100000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := v.ValidateCreateChallenge(tt.input)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// Update được phép giữ start_date trong quá khứ (admin sửa challenge đang diễn ra),
// còn Create thì không.
func TestValidateUpdateChallenge_AllowsPastStartDate(t *testing.T) {
	v := validations.NewContributionValidation()
	past := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	end := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	input := dto.CreateChallengeInput{
		Title:         "Ongoing Challenge",
		Hashtag:       "#Ongoing",
		PointsPerPost: 10,
		StartDate:     past,
		EndDate:       end,
	}

	start, endTime, err := v.ValidateUpdateChallenge(input)
	if err != nil {
		t.Fatalf("update should allow past start date, got %v", err)
	}
	if !endTime.After(start) {
		t.Fatalf("end = %v, must be after start = %v", endTime, start)
	}

	if _, _, err := v.ValidateCreateChallenge(input); err == nil {
		t.Fatal("create must reject a start date in the past")
	}
}

func TestValidateChallengeStatus(t *testing.T) {
	v := validations.NewContributionValidation()

	for _, status := range []string{"active", "completed", "cancelled"} {
		if err := v.ValidateChallengeStatus(status); err != nil {
			t.Errorf("status %q: unexpected error %v", status, err)
		}
	}
	for _, status := range []string{"", "done", "ACTIVE ", "deleted"} {
		if err := v.ValidateChallengeStatus(status); err == nil {
			t.Errorf("status %q: expected error, got nil", status)
		}
	}
}

func intPtr(v int) *int { return &v }

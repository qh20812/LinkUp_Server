package contribution_test

import (
	"testing"

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
				StartDate:       "2026-09-01T00:00:00Z",
				EndDate:         "2026-09-07T00:00:00Z",
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
				StartDate:     "2026-09-01T00:00:00Z",
				EndDate:       "2026-09-07T00:00:00Z",
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
				StartDate:     "2026-09-10T00:00:00Z",
				EndDate:       "2026-09-03T00:00:00Z",
			},
			wantErr: "Ngày kết thúc phải sau ngày bắt đầu",
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

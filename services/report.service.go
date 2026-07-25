package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"

	"gorm.io/gorm"
)

type ReportService struct {
	reportRepo *repository.ReportRepository
	authRepo   *repository.AuthRepository
	postRepo   *repository.PostRepository
	validation *validations.ReportValidation
}

func NewReportService(reportRepo *repository.ReportRepository, authRepo *repository.AuthRepository, postRepo *repository.PostRepository, validation *validations.ReportValidation) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		authRepo:   authRepo,
		postRepo:   postRepo,
		validation: validation,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, reporterID string, input dto.CreateReportInput) (dto.CreateReportResponse, error) {
	if err := s.validation.ValidateCreateReport(input.TargetType, input.TargetID, input.ReportType, input.ReasonDetail); err != nil {
		return dto.CreateReportResponse{}, err
	}

	existing, err := s.reportRepo.FindPendingByReporterAndTarget(ctx, reporterID, input.TargetType, input.TargetID)
	if err != nil {
		return dto.CreateReportResponse{}, fmt.Errorf("check existing report: %w", err)
	}
	if existing != nil {
		return dto.CreateReportResponse{}, errors.New("bạn đã có báo cáo đang chờ xử lý cho đối tượng này, vui lòng chỉnh sửa báo cáo thay vì tạo mới")
	}

	switch input.TargetType {
	case "user":
		if reporterID == input.TargetID {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo chính mình")
		}
		isAdmin, err := s.authRepo.HasRole(ctx, input.TargetID, models.RoleAdmin)
		if err != nil {
			return dto.CreateReportResponse{}, fmt.Errorf("check target role: %w", err)
		}
		isSuperAdmin, err := s.authRepo.HasRole(ctx, input.TargetID, models.RoleSuperAdmin)
		if err != nil {
			return dto.CreateReportResponse{}, fmt.Errorf("check target role: %w", err)
		}
		if isAdmin || isSuperAdmin {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo quản trị viên hoặc siêu quản trị viên")
		}
		targetUser, err := s.authRepo.FindByID(ctx, input.TargetID)
		if err != nil {
			return dto.CreateReportResponse{}, fmt.Errorf("check target user: %w", err)
		}
		if targetUser.IsBanned() {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo người dùng đã bị cấm")
		}
	case "post":
		post, err := s.postRepo.FindByID(ctx, input.TargetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.CreateReportResponse{}, errors.New("không tìm thấy bài viết hoặc bài viết không hoạt động")
			}
			return dto.CreateReportResponse{}, fmt.Errorf("check post: %w", err)
		}
		if post.Status == models.PostStatusHidden {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo bài viết đã bị ẩn")
		}
		if post.UserID == reporterID {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo bài viết của chính mình")
		}
	case "comment":
		comment, err := s.postRepo.FindCommentByID(ctx, input.TargetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.CreateReportResponse{}, errors.New("không tìm thấy bình luận hoặc bình luận không hoạt động")
			}
			return dto.CreateReportResponse{}, fmt.Errorf("check comment: %w", err)
		}
		if comment.Status == models.CommentStatusHidden {
			return dto.CreateReportResponse{}, errors.New("không thể báo cáo bình luận đã bị ẩn")
		}
	}

	now := time.Now().UTC()

	report := models.Report{
		ID:              utils.GenerateUUID(),
		ReporterID:      reporterID,
		ReportType:      input.ReportType,
		ViolationRuleID: input.ViolationRuleID,
		ReasonDetail:    input.ReasonDetail,
		Status:          models.ReportStatusPending,
		CreatedAt:       now,
	}

	switch input.TargetType {
	case "user":
		report.TargetUserID = &input.TargetID
	case "post":
		report.TargetPostID = &input.TargetID
	case "comment":
		report.TargetCommentID = &input.TargetID
	}

	if err := s.reportRepo.Create(ctx, &report); err != nil {
		return dto.CreateReportResponse{}, fmt.Errorf("create report: %w", err)
	}

	return dto.CreateReportResponse{
		Message: "Báo cáo đã được gửi thành công",
	}, nil
}

func (s *ReportService) UpdateReport(ctx context.Context, reporterID, reportID string, input dto.UpdateReportInput) (dto.UpdateReportResponse, error) {
	if err := s.validation.ValidateUpdateReport(input.ReportType, input.ReasonDetail); err != nil {
		return dto.UpdateReportResponse{}, err
	}

	report, err := s.reportRepo.FindByID(ctx, reportID)
	if err != nil {
		return dto.UpdateReportResponse{}, err
	}

	if report.ReporterID != reporterID {
		return dto.UpdateReportResponse{}, errors.New("bạn không có quyền chỉnh sửa báo cáo này")
	}

	if report.Status != models.ReportStatusPending {
		return dto.UpdateReportResponse{}, errors.New("chỉ có thể chỉnh sửa báo cáo đang chờ xử lý")
	}

	if err := s.reportRepo.Update(ctx, reportID, input.ReportType, input.ReasonDetail, input.ViolationRuleID); err != nil {
		return dto.UpdateReportResponse{}, fmt.Errorf("update report: %w", err)
	}

	return dto.UpdateReportResponse{
		Message: "Báo cáo đã được cập nhật thành công",
	}, nil
}

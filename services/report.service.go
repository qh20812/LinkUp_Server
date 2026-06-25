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

	switch input.TargetType {
	case "user":
		isAdmin, err := s.authRepo.HasRole(ctx, input.TargetID, models.RoleAdmin)
		if err != nil {
			return dto.CreateReportResponse{}, fmt.Errorf("check target role: %w", err)
		}
		isSuperAdmin, err := s.authRepo.HasRole(ctx, input.TargetID, models.RoleSuperAdmin)
		if err != nil {
			return dto.CreateReportResponse{}, fmt.Errorf("check target role: %w", err)
		}
		if isAdmin || isSuperAdmin {
			return dto.CreateReportResponse{}, errors.New("cannot report admin or super admin")
		}
	case "post":
		_, err := s.postRepo.FindByID(ctx, input.TargetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.CreateReportResponse{}, errors.New("post not found or not active")
			}
			return dto.CreateReportResponse{}, fmt.Errorf("check post: %w", err)
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

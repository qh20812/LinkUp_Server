package services

import (
	"context"
	"fmt"
	"time"

	"linkup/dto"
	"linkup/models"
	"linkup/repository"
	"linkup/utils"
	"linkup/validations"
)

type ReportService struct {
	reportRepo *repository.ReportRepository
	validation *validations.ReportValidation
}

func NewReportService(reportRepo *repository.ReportRepository, validation *validations.ReportValidation) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		validation: validation,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, reporterID string, input dto.CreateReportInput) (dto.CreateReportResponse, error) {
	if err := s.validation.ValidateCreateReport(input.TargetType, input.TargetID, input.ReportType, input.ReasonDetail); err != nil {
		return dto.CreateReportResponse{}, err
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

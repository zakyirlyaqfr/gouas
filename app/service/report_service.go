package service

import "gouas/app/repository"

type ReportService interface {
	GetStatistics() (map[string]interface{}, error)
}

type reportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo}
}

func (s *reportService) GetStatistics() (map[string]interface{}, error) {
	return s.repo.GetAchievementStats()
}
package service

import (
	"gouas/app/model"
	"gouas/app/repository"
)

type ReportService interface {
	GetDashboardStatistics() (*model.DashboardStats, error)
}

type reportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

func (s *reportService) GetDashboardStatistics() (*model.DashboardStats, error) {
	// 1. Ambil data status dari Postgres
	statusStats, err := s.repo.CountByStatus()
	if err != nil {
		return nil, err
	}

	// 2. Ambil data tipe dari Mongo
	typeStats, err := s.repo.CountByType()
	if err != nil {
		return nil, err
	}

	// 3. Hitung Total
	total := 0
	for _, count := range statusStats {
		total += count
	}

	// 4. Return Data Gabungan
	return &model.DashboardStats{
		TotalAchievements: total,
		ByStatus:          statusStats,
		ByType:            typeStats,
	}, nil
}
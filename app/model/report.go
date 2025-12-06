package model

type DashboardStats struct {
	TotalAchievements int            `json:"total_achievements"`
	ByStatus          map[string]int `json:"by_status"`
	ByType            map[string]int `json:"by_type"`
}
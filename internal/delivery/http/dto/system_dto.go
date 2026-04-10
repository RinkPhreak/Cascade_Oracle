package dto

// BreakGlassRequest model for pausing the system or campaign securely
type BreakGlassRequest struct {
	Reason   string `json:"reason,omitempty"`
	Password string `json:"password" validate:"required"`
}

// SystemMetricsResponse строго совпадает с фронтенд-типом SystemMetrics
type SystemMetricsResponse struct {
	CascadeMemoryUsageRatio float64 `json:"cascade_memory_usage_ratio"`
	ActiveTgAccounts        int     `json:"active_tg_accounts"`
	TotalTgAccounts         int     `json:"total_tg_accounts"`
	QueueDepth              int     `json:"queue_depth"`
	SystemStatus            string  `json:"system_status"` // 'OPERATIONAL', 'DEGRADED', 'HALTED'
}

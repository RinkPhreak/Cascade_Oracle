package dto

// BreakGlassRequest model for pausing the system or campaign securely
type BreakGlassRequest struct {
	Reason   string `json:"reason,omitempty"`
	Password string `json:"password" validate:"required"`
}

type SystemMetricsResponse struct {
	IsHalted bool   `json:"is_halted"`
	Reason   string `json:"reason,omitempty"`
}

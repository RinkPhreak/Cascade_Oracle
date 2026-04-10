package dto

type RegisterAccountRequest struct {
	Phone       string `json:"phone" validate:"required"`
	Channel     string `json:"channel" validate:"required"`
	ProxyID     string `json:"proxy_id,omitempty"`
	Credentials string `json:"credentials,omitempty"`
}

type AccountResponse struct {
	ID              string  `json:"id"`
	Phone           string  `json:"phone"`
	Channel         string  `json:"channel"`
	ProxyID         string  `json:"proxy_id,omitempty"`
	Status          string  `json:"status"`
	DailyCheckCount int     `json:"daily_check_count"`
	DailySendCount  int     `json:"daily_send_count"`
	CreatedAt       string  `json:"created_at"`
}

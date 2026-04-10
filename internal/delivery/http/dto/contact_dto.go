package dto

// AnonymiseResponse confirms deletion and hash results
type AnonymiseResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ContactID string `json:"contact_id"`
}

type LeadResponse struct {
	ID        string `json:"id"`
	Phone     string `json:"phone"`
	Name      string `json:"name"`
	Message   string `json:"message"`
	Channel   string `json:"channel"`
	RepliedAt string `json:"replied_at"`
}

type ContactTraceResponse struct {
	ID           string  `json:"id"`
	Channel      string  `json:"channel"`
	Status       string  `json:"status"`
	ErrorCode    *string `json:"error_code,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	Attempt      int     `json:"attempt"`
	Timestamp    string  `json:"timestamp"`
}

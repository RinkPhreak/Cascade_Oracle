package dto

// AnonymiseResponse confirms deletion and hash results
type AnonymiseResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ContactID string `json:"contact_id"`
}

package dto

type CreateProxyRequest struct {
	Address string `json:"address" validate:"required"`
}

type ProxyResponse struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

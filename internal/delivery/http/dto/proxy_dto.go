package dto

type CreateProxyRequest struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProxyResponse struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

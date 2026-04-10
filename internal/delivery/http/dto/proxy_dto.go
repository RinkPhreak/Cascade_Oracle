package dto

type CreateProxyRequest struct {
	ID       string `json:"id"`
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProxyResponse struct {
	ID        string `json:"id"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

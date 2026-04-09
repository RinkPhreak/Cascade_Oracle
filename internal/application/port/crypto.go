package port

// CryptoService abstracts AES-256-GCM encryption and password hashing logic.
type CryptoService interface {
	Encrypt(plain string) (string, error)
	Decrypt(cipher string) (string, error)
	HashPhone(phone string) string
	HashPassword(password string) (string, error)
	ComparePassword(hash, password string) error
}

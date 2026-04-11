package domain

type TelethonSession struct {
	DCID          int
	ServerAddress string
	Port          int
	AuthKey       []byte
	TakeoutID     int
}

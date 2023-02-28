package server

type Config struct {
	Host         string `env:""`
	Port         uint   `env:""`
	TCPKeepAlive bool   `env:""`
	TCPNoDelay   bool   `env:""`
}

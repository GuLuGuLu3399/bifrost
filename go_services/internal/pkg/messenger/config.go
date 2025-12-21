package messenger

type Config struct {
	Addr       string `yaml:"addr"`
	ClientName string `yaml:"client_name"`
}

func DefaultConfig() *Config {
	return &Config{
		Addr:       "nats://127.0.0.1:4222",
		ClientName: "bifrost_service",
	}
}

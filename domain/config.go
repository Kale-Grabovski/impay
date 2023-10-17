package domain

const EnvPrefix = "IMPAY"

type Config struct {
	LogLevel   string `yaml:"logLevel"`
	WalletPort string `yaml:"walletPort"`
	StatsPort  string `yaml:"statsPort"`
	Kafka      struct {
		Host string `yaml:"host"`
	} `yaml:"kafka"`
}

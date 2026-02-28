package config

type Config struct {
	MetricsPort            int    `yaml:"metricsPort"`
	PollIntervalSecs       int    `yaml:"pollIntervalSecs"`
	RestartVerifyDelaySecs int    `yaml:"restartVerifyDelaySecs"`
	LogLevel               string `yaml:"logLevel"`
	DiscordWebhook         string `yaml:"discordWebhook"`
}

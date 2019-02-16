package config

type RateLimitConfig struct {
	Limit     int
	Remaining int
	Reset     string
}

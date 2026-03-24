package types

type ServerConfig struct {
	Port int 
	Environment string
	DSN string
	AppVersion string
	Limiter struct {
		RPS     float64
		Burst   int
		Enabled bool
	}
	CORS struct {
		TrustedOrigins []string
	}
}
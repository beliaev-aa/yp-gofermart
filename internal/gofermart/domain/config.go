package domain

// Config - описывает конфигурацию приложения
type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8080"`
	JWTSecret            string `env:"JWT_SECRET" envDefault:"your-256-bit-secret-key"`
}

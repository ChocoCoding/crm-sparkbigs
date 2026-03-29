package domain

// JWTConfig agrupa los parámetros de configuración para la generación y validación de JWT.
type JWTConfig struct {
	Secret              string
	AccessTokenDuration  int // minutos
	RefreshTokenDuration int // días
}

// AppConfig agrupa toda la configuración de la aplicación leída desde variables de entorno.
// Se construye en main.go y se pasa por inyección de dependencias.
type AppConfig struct {
	JWT           JWTConfig
	EncryptionKey string
	Port          string
	BaseURL       string
	CORSOrigins   string
	MySQLDSN      string
}

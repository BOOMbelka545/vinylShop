package config

type Properties struct {
	DBLog          string   `env:"DB_LOG" env-default:"postgres"`
	DBPas	       string   `env:"DB_PAS" env-default:"PCNXD3FE"`
	DBHost	       string   `env:"DB_HOST" env-default:"localhost"`
	DBName         string   `env:"DB_NAME" env-default:"vinylShop"`
	DBPort	       string   `env:"DB_PORT" env-default:"5432"`
	Port           string   `env:"MY_APP_PORT" env-default:"8080"`
	Host           string   `env:"HOST" env-default:"localhost"`
	JwtTokenSecret string   `env:"JWT_TOKEN_SECRET" env-default:"PASO4KA"`
}


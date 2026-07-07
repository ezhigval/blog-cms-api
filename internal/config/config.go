package config

import (
	"time"

	"github.com/ezhigval/go-toolkit/config"
)

type Config struct {
	Port         string `env:"PORT" envDefault:"8090"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat    string `env:"LOG_FORMAT" envDefault:"json"`
	DatabaseURL  string `env:"DATABASE_URL,required"`
	JWTSecret    string `env:"JWT_SECRET,required"`
	AccessTTL    time.Duration `env:"ACCESS_TTL" envDefault:"24h"`
	BcryptCost   int    `env:"BCRYPT_COST" envDefault:"12"`
	UploadDir    string `env:"UPLOAD_DIR" envDefault:"./uploads"`
	MaxUploadMB  int64  `env:"MAX_UPLOAD_MB" envDefault:"5"`
	CORSOrigins  []string `env:"CORS_ORIGINS" envDefault:"http://localhost:3001"`
	PublicURL    string `env:"PUBLIC_URL" envDefault:"http://localhost:8090"`
}

func MustLoad() Config {
	return config.MustLoad[Config]()
}

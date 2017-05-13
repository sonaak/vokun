package app


import "github.com/evilwire/go-env"


type DbConfig struct {
	Host string `env:"HOST"`
	Port string `env:"PORT"`
	User string `env:"USER"`
	Password string `env:"PASSWORD"`
}


type Config struct {
	Db DbConfig `env:"DB_"`
}


func NewConfig(env goenv.EnvReader) (*Config, error) {
	marshaler := goenv.DefaultEnvMarshaler{ env }
	config := Config {}
	err := marshaler.Unmarshal(&config)
	return &config, err
}

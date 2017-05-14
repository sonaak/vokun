package app


import "github.com/evilwire/go-env"


type DbConfig struct {
	Host string `env:"HOST",json:"host"`
	Port string `env:"PORT",json:"port"`
	User string `env:"USER",json:"user"`
	Password string `env:"PASSWORD"`
}


type Config struct {
	Db DbConfig `env:"DB_",json:"db"`
	Version string `env:"VERSION"`
}


func NewConfig(env goenv.EnvReader) (*Config, error) {
	marshaler := goenv.DefaultEnvMarshaler{ env }
	config := Config {}
	err := marshaler.Unmarshal(&config)
	return &config, err
}

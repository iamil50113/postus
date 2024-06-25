package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type Config struct {
	Env            string `yaml:"env" env-default:"local"`
	PostusConfig   `yaml:"postus"`
	HTTPServer     `yaml:"http_server"`
	PostgresConfig `yaml:"postgresql"`
}

type PostusConfig struct {
	CommentLenLimit         int  `yaml:"comment_length_limit"`
	PaginationCommentsLimit int  `yaml:"pagination_of_comments_limit"`
	UseInMemory             bool `yaml:"use_in_memory"`
}

type HTTPServer struct {
	Port string `yaml:"port" env-default:":8080"`
}

type PostgresConfig struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	DBName string `yaml:"database_name"`
	User   string `yaml:"username"`
	Pass   string `yaml:"password"`
}

func MustLoad() *Config {
	var configPath string

	flag.StringVar(&configPath, "config", "config/local.yaml", "path to config file")
	flag.Parse()

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

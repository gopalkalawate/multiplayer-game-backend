package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct{
	Address string `yaml:"address" env-required:"true"`
}

type Config struct{
	Env string `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer	`yaml:"http_server"`
}

func MustLoad() *Config{
	var configPath string;
	configPath = os.Getenv("CONFIG_PATH") // used cleanenv package

	if configPath == ""{
		flags := flag.String("config","","path to configuration file")
		flag.Parse()
		configPath = *flags

		if configPath==""{
			log.Fatal("Config path not set")
		}
	}

	{
		_,err :=  os.Stat(configPath)
	
		if err!=nil || os.IsNotExist(err){
		log.Fatal("config file doesn't exist: ",configPath)
		}
	}

	var cfg Config
	err := cleanenv.ReadConfig(configPath,&cfg)

	if err != nil{
		log.Fatalf("cannot read config file: %s",err.Error())
	}

	return &cfg
}
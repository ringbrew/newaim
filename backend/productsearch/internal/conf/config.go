package conf

import (
	"github.com/ringbrew/gsv/config"
)

type Config struct {
	Env           string        `yaml:"environment"`
	Debug         bool          `yaml:"debug"`
	Host          string        `yaml:"host"`
	Port          int           `yaml:"port"`
	Redis         Redis         `yaml:"redis"`
	Miluvs        Miluvs        `yaml:"miluvs"`
	OpenAI        OpenAI        `yaml:"openAI"`
	ElasticSearch ElasticSearch `yaml:"elasticSearch"`
	ForceRebuild  bool          `yaml:"forceRebuild"`
}

type Mysql struct {
	UserName string `yaml:"user_name"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Trace struct {
	Type     string `yaml:"type"`
	Endpoint string `yaml:"endpoint"`
}

type Miluvs struct {
	Endpoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type OpenAI struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
}

type ElasticSearch struct {
	Address  []string `yaml:"address"`
	UserName string   `yaml:"userName"`
	Password string   `yaml:"password"`
}

func Load(path string) (Config, error) {
	var result Config
	loader := config.NewLoader(config.LoaderTypeYml, path)
	if err := loader.Load(&result); err != nil {
		return Config{}, err
	}

	return result, nil
}

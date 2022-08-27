package config

import (
	"log"
	"os"

	"github.com/inlets/connect/bucket"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Buckets bucket.Buckets `yaml:"buckets"`
}

func Init(configPath string) (c *Config) {
	f, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("cannot read %s: %s", configPath, err.Error())

		return c
	}

	err = yaml.Unmarshal(f, &c)

	if err != nil {
		log.Printf("cannot decode %s: %s", configPath, err.Error())

		return c
	}

	for k, v := range c.Buckets {
		c.Buckets[k] = v.Init()
	}

	for k, v := range c.Buckets {
		//nolint:gomnd
		log.Printf("%s: duration: %s, limit: %gMB", k, v.LimitDuration, float64(v.Quota)/1024/1024)
	}

	return c
}

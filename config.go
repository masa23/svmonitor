package svmonitor

import (
	"io/ioutil"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Graphite       configGraphite `yaml:"Graphite"`
	ReportInterval time.Duration  `yaml:"ReportInterval"`
	Metric         configMetric   `yaml:"Metric"`
}

type configGraphite struct {
	Prefix     string `yaml:"Prefix"`
	Host       string `yaml:"Host"`
	Port       int    `yaml:"Port"`
	SendBuffer int    `yaml:"SendBuffer"`
}

type configMetric struct {
	CPU     configCPU     `yaml:"CPU"`
	Network configNetwork `yaml:"Network"`
}

type configCPU struct {
	Enable   bool `yaml:"Enable"`
	AllCores bool `yaml:"AllCores"`
}

type configNetwork struct {
	Enable     bool     `yaml:"Enable"`
	Interfaces []string `yaml:"Interfaces"`
}

func ConfigLoad(configFile string) (Config, error) {
	var conf Config
	fd, err := os.Open(configFile)
	if err != nil {
		return conf, err
	}
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		return conf, err
	}
	return conf, nil
}

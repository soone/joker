package config

import "github.com/spf13/viper"

type Cfg struct {
	Env      string `json:"env" mapstructure:"env"`
	LogLevel int    `json:"log_level" mapstructure:"log_level"`
	LogFile  string `json:"log_file" mapstructure:"log_file"`

	Services []SrvCfg `json:"services" mapstructure:"services"`
}

func NewCfg() *Cfg {
	cfg := &Cfg{}

	viper.Unmarshal(cfg)

	return cfg
}

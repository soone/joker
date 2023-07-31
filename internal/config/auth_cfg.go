package config

type AuthCfg struct {
	Enable bool   `json:"enable" mapstructure:"enable"`
	User   string `json:"user" mapstructure:"user"`
	Pass   string `json:"pass" mapstructure:"pass"`
}

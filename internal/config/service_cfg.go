package config

type SrvCfg struct {
	Name         string  `json:"name" mapstructure:"name"`
	Type         string  `json:"type" mapstructure:"type"`
	Enable       bool    `json:"enable" mapstructure:"enable"`
	ListenAddr   string  `json:"listen_addr" mapstructure:"listen_addr"`
	OutgoingAddr string  `json:"outgoing_addr" mapstructure:"outgoing_addr"`
	Auth         AuthCfg `json:"auth" mapstructure:"auth"`
}

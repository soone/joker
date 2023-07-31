package service

import (
	"io"
	"os"

	"github.com/soone/joker/internal/config"
	"github.com/soone/joker/internal/service/socks5"
	"github.com/soone/vegapunk/clog"
	"github.com/soone/vegapunk/initialize"
)

var cfg *config.Cfg

func JokerRun() {
	cfg = config.NewCfg()

	if len(cfg.LogFile) > 0 {
		file, _ := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		multi := io.MultiWriter(os.Stdout, file)
		clog.Logx = clog.New(multi, int64(cfg.LogLevel), cfg.Env, 0)
	}

	for _, sCfg := range cfg.Services {
		initialize.WG2Exec(func(args ...any) {
			sCfg := args[0].(config.SrvCfg)
			s := socks5.NewSocks5(initialize.GetContext(), sCfg)
			err := s.Run()
			if err != nil {
				clog.Logx.Errorf("Socks5 service run error: %v, sCfg: %v", err, sCfg)
			}
		}, sCfg)
	}
}

func GetIfaces(exclude []string) {

}

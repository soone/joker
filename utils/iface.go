package utils

import (
	"net"

	"github.com/soone/vegapunk/clog"
)

func GetIfaces() {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			panic(err)
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				clog.Logx.Debugf("%v %v", i.Name, v)
			case *net.IPAddr:
				clog.Logx.Debugf("%v %v", i.Name, v)
			}
		}
	}

}

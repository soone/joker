package socks5

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/soone/joker/internal/config"
	"github.com/soone/vegapunk/clog"
	"github.com/soone/vegapunk/initialize"
)

type Socks5Service struct {
	ctx    context.Context
	srvCfg config.SrvCfg
	server net.Listener
	mu     *sync.Mutex
	status config.SERVERSTATUS
}

func (s *Socks5Service) Run() error {

	var err error
	s.server, err = net.Listen("tcp", s.srvCfg.ListenAddr)
	if err != nil {
		return err
	}

	s.setStatus(config.RUNNING)

	initialize.WG2Exec(func(args ...any) {
		for {
			client, err := s.server.Accept()
			if err != nil {
				if s.status == config.STOPPED {
					return
				}

				clog.Logx.Errorf("Accept error: %v", err)
				continue
			}

			initialize.WG2Exec(s.handler, client)
		}
	})

	<-s.ctx.Done()
	s.Stop()

	return nil
}

func (s *Socks5Service) handler(args ...any) {

	client := args[0].(net.Conn)
	defer client.Close()

	if err := s.auth(client); err != nil {
		clog.Logx.Errorf("auth error: %v", err)
		return
	}

	if err := s.connectAndForward(client); err != nil {
		clog.Logx.Errorf("connectAndForward error: %v", err)
		return
	}
}

func (s *Socks5Service) Stop() error {
	if s.server != nil {
		s.setStatus(config.STOPPED)
		return s.server.Close()
	}

	return nil
}

func (s *Socks5Service) setStatus(status config.SERVERSTATUS) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
}

func (s *Socks5Service) connectAndForward(client net.Conn) error {
	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		return err
	}

	if n < 10 {
		return fmt.Errorf("[0x95]invalid socks5 protocol")
	}

	cmd := buf[1]
	if cmd != 0x01 {
		return fmt.Errorf("[0x100]invalid socks5 protocol")
	}

	atyp := buf[3]

	var dstAddr string
	switch atyp {
	case 0x01:
		if n < 4+net.IPv4len {
			return fmt.Errorf("[0x109]invalid socks5 protocol")
		}

		dstAddr = net.IP(buf[4 : 4+net.IPv4len]).String()
	case 0x03:
		if n < 5 {
			return fmt.Errorf("[0x115]invalid socks5 protocol")
		}

		addrLen := int(buf[4])
		if n < 5+addrLen+2 {
			return fmt.Errorf("[0x120]invalid socks5 protocol")
		}
		dstAddr = string(buf[5 : 5+addrLen])
	case 0x04:
		if n < 4+net.IPv6len {
			return fmt.Errorf("[0x125]invalid socks5 protocol")
		}
		dstAddr = net.IP(buf[4 : 4+net.IPv6len]).String()
	default:
		return fmt.Errorf("[0x129]invalid socks5 protocol")
	}

	dstPort := int(buf[n-2])<<8 + int(buf[n-1])
	// dstPort := binary.BigEndian.Uint16(buf[:2])
	clog.Logx.Debugf("dstAddr: %s, dstPort: %d", dstAddr, dstPort)
	var dstConn net.Conn
	if s.srvCfg.OutgoingAddr != "" {
		dstConn, err = net.DialTCP("tcp", &net.TCPAddr{IP: net.ParseIP(s.srvCfg.OutgoingAddr)}, &net.TCPAddr{IP: net.ParseIP(dstAddr), Port: dstPort})
	} else {
		dstConn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", dstAddr, dstPort))
	}

	if err != nil {
		return err
	}

	defer dstConn.Close()

	client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	forwardFunc := func(args ...any) {
		src := args[0].(net.Conn)
		dest := args[1].(net.Conn)

		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}

	initialize.WG2Exec(forwardFunc, dstConn, client)
	// initialize.WG2Exec(forwardFunc, client, dstConn)
	forwardFunc(client, dstConn)

	return nil
}

func (s *Socks5Service) auth(client net.Conn) error {
	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		return err
	}

	if n < 3 {
		return fmt.Errorf("invalid socks5 protocol")
	}
	ver := buf[0]

	if s.srvCfg.Auth.Enable {
		nMethod := int(buf[1])
		methods := buf[2 : 2+nMethod]

		var hasAuth bool
		for _, m := range methods {
			if m == 0x02 {
				hasAuth = true
				break
			}
		}

		if !hasAuth {
			client.Write([]byte{ver, 0xff})
			return fmt.Errorf("client does not support username/password authentication")
		}

		client.Write([]byte{ver, 0x02})

		n, err = client.Read(buf)
		if err != nil {
			return err
		}

		if n < 4 {
			return fmt.Errorf("invalid socks5 protocol")
		}

		usernameLen := int(buf[1])
		username := string(buf[2 : 2+usernameLen])
		passwordLen := int(buf[2+usernameLen])
		password := string(buf[3+usernameLen : 3+usernameLen+passwordLen])

		if s.srvCfg.Auth.User != username || s.srvCfg.Auth.Pass != password {
			return fmt.Errorf("invalid username/password")
		}

		client.Write([]byte{ver, 0x00})

	} else {
		client.Write([]byte{ver, 0x00})
	}

	return nil
}

func NewSocks5(ctx context.Context, srvCfg config.SrvCfg) *Socks5Service {
	return &Socks5Service{
		ctx:    ctx,
		srvCfg: srvCfg,
		mu:     &sync.Mutex{},
	}
}

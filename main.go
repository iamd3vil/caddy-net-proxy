package caddynet

import (
	"io"
	"log"
	"net"

	"github.com/caddyserver/caddy/v2"
)

func init() {
	err := caddy.RegisterModule(Proxy{})
	if err != nil {
		log.Fatal(err)
	}
}

// Proxy implements Caddy module interface for TCP/UDP proxying
type Proxy struct {
	Source       string `json:"source,omitempty"`
	Destination  string `json:"destination,omitempty"`
	laddr, raddr *net.TCPAddr
}

// CaddyModule returns the Caddy module information.
func (Proxy) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		Name: "proxy",
		New: func() caddy.Module {
			return new(Proxy)
		},
	}
}

func (p *Proxy) Start() error {
	laddr, err := net.ResolveTCPAddr("tcp", p.Source)
	if err != nil {
		return err
	}

	raddr, err := net.ResolveTCPAddr("tcp", p.Destination)
	if err != nil {
		return err
	}

	p.laddr = laddr
	p.raddr = raddr

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatalf("Failed to open local port to listen: %s", err)
	}

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Failed to accept connection '%s'", err)
			continue
		}

		defer conn.Close()

		rconn, err := net.DialTCP("tcp", nil, raddr)
		if err != nil {
			conn.Close()
			continue
		}
		defer rconn.Close()

		closeChan := make(chan int, 2)

		go func() {
			io.Copy(rconn, conn)
			closeChan <- 1
		}()
		go func() {
			io.Copy(conn, rconn)
			closeChan <- 1
		}()
		<-closeChan
		rconn.Close()
		conn.Close()
	}
}

func (p *Proxy) Stop() error {
	return nil
}

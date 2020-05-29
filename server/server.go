package server

import (
	"errors"
	"github.com/lishengye/sfs"
	"github.com/lishengye/sfs/log"
	"net"
)

type Server struct {
	Config      Config
	ClientCount int16
	Listener    *net.TCPListener
}

func NewServer(config Config) *Server {
	return &Server{
		Config:      config,
		ClientCount: 0,
	}
}

func (server *Server) init() error {
	return nil
}

func (server *Server) Run() error {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: int(server.Config.Port),
		IP:   net.IPv4(0, 0, 0, 0),
	})
	if err != nil {
		log.Error("Listening on %v: err: %s", server.Config.Port, err.Error())
		return errors.New("Server listening error")
	}
	log.Info("Server start listening")
	server.Listener = ln
	failed := 0
	for {
		conn, err := server.Listener.Accept()
		if err != nil {
			log.Error("Aceep error: %s, failed: %d", err.Error(), failed)
			if failed++; failed > 5 {
				break
			}
			continue
		}
		log.Info("Handling client: %s", conn.RemoteAddr().String())
		clientHandler := &ClientHandler{
			connection: sfs.NewConnection(conn),
			Config:     server.Config,
		}
		go func() {
			defer log.Info("Client close: %s", conn.RemoteAddr().String())
			clientHandler.Handle()
		}()
	}
	return nil
}

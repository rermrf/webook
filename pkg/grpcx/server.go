package grpcx

import (
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	*grpc.Server
	Addr string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", s.Addr)
	log.Println("server worked on：", s.Addr)
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

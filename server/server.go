/*
Package server define how to serve RPC services.

Usage:

		type Upper struct {}

		func (Upper) Handle(rw protocol.ResponseWriter, req *protocol.Request) {
			text := req.Params["text"].(string)
			rw.Set(strings.ToUpper(text))
		}

		serv := server.New(jsonrpc.New(http.New()))
		serv.Register("Upper", Upper{})

		_ = serv.Serve()

*/
package server

import (
	"errors"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/mouadino/go-nano/discovery"
	"github.com/mouadino/go-nano/handler"
	"github.com/mouadino/go-nano/handler/middleware"
	"github.com/mouadino/go-nano/protocol"
	"github.com/mouadino/go-nano/transport"
)

// Server represents an RPC server.
type Server struct {
	proto protocol.Protocol
	mux   handlersMux
	metas map[string]map[string]interface{}
}

// New create a Server.
func New(proto protocol.Protocol) *Server {
	return &Server{
		proto: proto,
		mux:   handlersMux{make(map[string]handler.Handler)},
		metas: make(map[string]map[string]interface{}),
	}
}

// Register given handler under name.
func (s *Server) Register(name string, svc interface{}, ms ...handler.Middleware) error {
	hdlr := middleware.Chain(handler.New(svc), ms...)
	return s.mux.Add(name, hdlr)
}

// Register given handler under name with given metadata.
func (s *Server) RegisterWithMetadata(name string, svc interface{}, meta map[string]interface{}, ms ...handler.Middleware) error {
	err := s.Register(name, svc, ms...)
	if err != nil {
		return err
	}
	s.metas[name] = meta
	return nil
}

// Serve listens on transport addr (if there is any) and then
// start handling requests from transport.
func (s *Server) Serve() {
	s.listen()
	go s.loop()
	wait()
}

// ServeAndAnnounce start by serving server than announce it in the given Announcer.
func (s *Server) ServeAndAnnounce(an discovery.Announcer) error {
	s.listen()
	err := s.announce(an)
	if err != nil {
		return err
	}
	go s.loop()
	wait()
	return nil
}

func (s *Server) listen() {
	trans := s.proto.Transport()
	trans.Listen()
}

func (s *Server) announce(an discovery.Announcer) error {
	addr, ok := s.proto.Transport().(transport.Addresser)
	if !ok {
		return errors.New("can only announce transport of type transport.Addresser")
	}

	for _, name := range s.mux.Names() {
		instance := discovery.NewInstance(addr.Addr(), s.metas[name])
		err := an.Announce(name, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) loop() {
	for {
		resp, req, err := s.proto.Receive()
		if err != nil {
			log.Errorf("transport receive failed: %s", err)
			continue
		}
		if err != nil {
			log.Errorf("code failed to decode: %s", err)
			continue
		}
		go s.mux.Handle(resp, req)
	}
}

func wait() {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		log.Print("Received SIGTERM, exiting ...")
	}
}

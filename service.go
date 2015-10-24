package nano

import (
	"log"

	"github.com/mouadino/go-nano/handler"
	"github.com/mouadino/go-nano/protocol"
	"github.com/mouadino/go-nano/reflection"
	"github.com/mouadino/go-nano/transport"
)

// TODO: Cli flags

type Service struct {
	transport transport.Transport
	protocol  protocol.Protocol
	handler   handler.Handler
	svc       interface{}
}

func (s *Service) ListenAndServe() {
	if s, ok := s.svc.(Startable); ok {
		err := s.NanoStart()
		if err != nil {
			log.Fatalf("Service failed to start: %s", err)
		}
		defer s.NanoStop()
	}
	// TODO: CTRL-C Catch.
	// TODO: goroutine Pool.
	// FIXME: Harcoded listen.
	go s.transport.Listen("127.0.0.1:8080")
	for {
		resp, req := s.protocol.ReceiveRequest()
		log.Printf("%s -> %s\n", req, resp)
		go s.handler.Handle(resp, req)
	}
}

func Default(service interface{}) *Service {
	trans := transport.NewHTTPTransport()
	return Custom(
		service,
		trans,
		protocol.NewJSONRPCProtocol(trans),
	)
}

func Custom(svc interface{}, trans transport.Transport, proto protocol.Protocol) *Service {
	return &Service{
		svc:       svc,
		transport: trans,
		protocol:  proto,
		handler:   reflection.FromStruct(svc),
	}
}
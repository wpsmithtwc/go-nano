package jsonrpc

import (
	"bytes"

	"github.com/mouadino/go-nano/header"
	"github.com/mouadino/go-nano/protocol"
	"github.com/mouadino/go-nano/serializer"
	"github.com/mouadino/go-nano/transport"

	log "github.com/Sirupsen/logrus"
)

type RequestBody struct {
	Version string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	Id      string                 `json:"id"`
}

func (b *RequestBody) ToRequest() *protocol.Request {
	return &protocol.Request{
		Method: b.Method,
		Params: b.Params,
	}
}

type JSONRPCProtocol struct {
	trans  transport.Transport
	serial serializer.Serializer
}

func Serializer(serial serializer.Serializer) func(*JSONRPCProtocol) {
	return func(p *JSONRPCProtocol) {
		p.serial = serial
	}
}

func New(trans transport.Transport, options ...func(*JSONRPCProtocol)) *JSONRPCProtocol {
	proto := &JSONRPCProtocol{
		trans:  trans,
		serial: serializer.JSONSerializer{},
	}

	for _, opt := range options {
		opt(proto)
	}
	return proto
}

func (proto *JSONRPCProtocol) Transport() transport.Transport {
	return proto.trans
}

func (proto *JSONRPCProtocol) Send(endpoint string, r *protocol.Request) (*protocol.Response, error) {
	log.WithFields(log.Fields{
		"endpoint": endpoint,
		"method":   r.Method,
	}).Info("sending")
	reqBody, err := proto.getBody(r)
	if err != nil {
		return nil, err
	}
	b, err := proto.trans.Send(endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	respBody := ResponseBody{}
	err = proto.serial.Decode(b, &respBody)
	if err != nil {
		return nil, err
	}
	return respBody.ToResponse(), nil
}

func (proto *JSONRPCProtocol) getBody(r *protocol.Request) ([]byte, error) {
	body := RequestBody{
		Version: "2.0",
		Method:  r.Method,
		Params:  r.Params,
		Id:      "0", // TODO: gouuid !?
	}
	return proto.serial.Encode(body)
}

func (proto *JSONRPCProtocol) Receive() (protocol.ResponseWriter, *protocol.Request, error) {
	b := <-proto.trans.Receive()
	body := RequestBody{}
	err := proto.serial.Decode(b.Body.([]byte), &body)
	if err != nil {
		return nil, nil, err
	}
	rw := &JSONRPCResponseWriter{
		b.Resp,
		proto,
		header.Header{},
	}
	return rw, body.ToRequest(), nil
}

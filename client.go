package nano

import (
	"github.com/mouadino/go-nano/protocol"
	"github.com/mouadino/go-nano/transport"
)

type remoteClient struct {
	endpoint string
	proto    protocol.Protocol
}

func Client(endpoint string) *remoteClient {
	return &remoteClient{
		endpoint: endpoint,
		proto:    protocol.NewJSONRPCProtocol(transport.NewHTTPTransport()),
	}
}

func (client *remoteClient) Call(method string, params map[string]interface{}) (interface{}, error) {
	req := protocol.Request{
		Method: method,
		Params: params,
	}
	resp, err := client.proto.SendRequest(client.endpoint, &req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

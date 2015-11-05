package protocol

import (
	"fmt"

	"github.com/mouadino/go-nano/serializer"
	"github.com/mouadino/go-nano/transport"
)

type RequestBody struct {
	Version string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	Id      string                 `json:"id"`
}

func (b *RequestBody) ToRequest() *Request {
	return &Request{
		Method: b.Method,
		Params: b.Params,
	}
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (b *ErrorBody) Error() error {
	// TODO: This should be the reverse of FromNanoError.
	return fmt.Errorf("%s: %s", b.Code, b.Message)
}

func FromNanoError(err error) *ErrorBody {
	// TODO: Set http status,
	// TODO: Set Data.
	switch {
	case err == UnknownMethod:
		return &ErrorBody{
			Code:    "-32601",
			Message: err.Error(),
			Data:    "",
		}
	case err == ParamsError:
		return &ErrorBody{
			Code:    "-32602",
			Message: err.Error(),
			Data:    "",
		}
	case err == InternalError:
		return &ErrorBody{
			Code:    "-32603",
			Message: err.Error(),
			Data:    "",
		}
	default:
		return &ErrorBody{
			Code:    "-32000",
			Message: "Server error",
			Data:    err.Error(),
		}
	}
}

type ResponseBody struct {
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *ErrorBody  `json:"error"`
	Id      string      `json:"id"`
}

// TODO: Use me !
func (b *ResponseBody) ToRespone() {
}

type JSONRPCResponseWriter struct {
	transport.ResponseWriter
	p *JSONRPCProtocol
}

func (w *JSONRPCResponseWriter) Write(data interface{}) error {
	body := ResponseBody{
		Version: "2.0",
		Result:  data,
		Id:      "0",
	}
	return w.writeToTransport(&body)
}

func (w *JSONRPCResponseWriter) WriteError(err error) error {
	body := ResponseBody{
		Version: "2.0",
		Result:  nil,
		Error:   FromNanoError(err),
		Id:      "0",
	}
	return w.writeToTransport(&body)
}

func (w *JSONRPCResponseWriter) writeToTransport(body *ResponseBody) error {
	b, err := w.p.serializer.Encode(body)
	if err != nil {
		return err
	}
	err = w.ResponseWriter.Write(b)
	if err != nil {
		return err
	}
	return nil
}

type JSONRPCProtocol struct {
	transport  transport.Transport
	serializer serializer.Serializer
}

// TODO: DI for serializer !
func NewJSONRPCProtocol(t transport.Transport) *JSONRPCProtocol {
	return &JSONRPCProtocol{
		transport:  t,
		serializer: serializer.JSONSerializer{},
	}
}

// TODO: should we return Response !?
func (p *JSONRPCProtocol) SendRequest(endpoint string, r *Request) (interface{}, error) {
	b, err := p.getBody(r)
	if err != nil {
		return nil, err
	}
	resp, err := p.transport.Send(endpoint, b)
	if err != nil {
		return nil, err
	}
	data, err := resp.Read()
	if err != nil {
		return nil, err
	}
	respBody := ResponseBody{}
	err = p.serializer.Decode(data, &respBody)
	if err != nil {
		return nil, err
	}
	if respBody.Error != nil {
		return nil, respBody.Error.Error()
	}
	return respBody.Result, nil
}

func (p *JSONRPCProtocol) getBody(r *Request) ([]byte, error) {
	body := RequestBody{
		Version: "2.0",
		Method:  r.Method,
		Params:  r.Params,
		Id:      "0", // TODO: gouuid !?
	}
	return p.serializer.Encode(body)
}

func (p *JSONRPCProtocol) ReceiveRequest() (transport.ResponseWriter, *Request) {
	b := <-p.transport.Receive()
	body := RequestBody{}
	err := p.serializer.Decode(b.Body, &body)
	if err != nil {
		return nil, nil
	}
	return &JSONRPCResponseWriter{
		b.Resp,
		p,
	}, body.ToRequest()
}

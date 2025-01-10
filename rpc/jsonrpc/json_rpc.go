package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"
)

const jsonrpcVersion = "2.0"

//func NewRequest(method string, params ...interface{}) *RPCRequest{
//  request := &RPCRequest{
//    JSONRPC: jsonrpcVersion,
//    Id:      newID(),
//    Method:  method,
//    Params:  Params(params...),
//  }
// return request
//}

//type RPCClient interface{
//  Call(ctx context.Context, method string, params ...interface{}) (*RPCResponse, error)
//}

type RPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Id      any         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Id      any             `json:"id"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCResponses []*RPCResponse

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var spewConf = spew.ConfigState{
	Indent:                " ",
	DisableMethods:        true,
	DisablePointerMethods: true,
	SortKeys:              true,
}

func (e *RPCError) Error() string {
	return spewConf.Sdump(e)
}

type HTTPError struct {
	Code int
	err  error
}

func (e *HTTPError) Error() string {
	return e.err.Error()
}

// This is an abstraction over the http client
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
	CloseIdleConnections()
}

type RPCClientOpts struct {
	HTTPClient    HTTPClient
	CustomHeaders map[string]string
}

type rpcClient struct {
	endpoint      string
	httpClient    HTTPClient
	customHeaders map[string]string
}

func NewClient(endpoint string) *rpcClient {
	return NewClientWithOpts(endpoint, nil)
}

func NewClientWithOpts(endpoint string, opts *RPCClientOpts) *rpcClient {
	rpcClient := &rpcClient{
		endpoint:      endpoint,
		httpClient:    &http.Client{},
		customHeaders: make(map[string]string),
	}

	if opts == nil {
		return rpcClient
	}

	if opts.HTTPClient != nil {
		rpcClient.httpClient = opts.HTTPClient
	}

	if opts.CustomHeaders != nil {
		for k, v := range opts.CustomHeaders {
			rpcClient.customHeaders[k] = v
		}
	}
	return rpcClient
}

func (client *rpcClient) newRequest(ctx context.Context, reqBody interface{}) (*http.Request, error) {
	// reqBytes -> slice of bytes reprenesting reqBody struct in JSON format
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", client.endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return request, err
	}

	request.Header.Set("Content-Type", "application/json") // request -> struct, Header map[string], Set is (key, val) method
	request.Header.Set("Accept", "application/json")

	for k, v := range client.customHeaders {
		request.Header.Set(k, v)
	}

	return request, nil
}

func (client *rpcClient) makeCallWithCallbackOnHTTPResponse(ctx context.Context, RPCRequest *RPCRequest,
	callback func(*http.Request, *http.Response) error) error {
	if RPCRequest != nil && RPCRequest.Id == nil {
		RPCRequest.Id = newID()
	}

	httpRequest, err := client.newRequest(ctx, RPCRequest) // <-- format http request
	if err != nil {
		if httpRequest != nil {
			return fmt.Errorf("rpc call %v() on %v: %w", RPCRequest.Method, httpRequest.URL.String(), err)
		}
		return fmt.Errorf("rpc call %v(): %w", RPCRequest.Method, err)
	}
	httpResponse, err := client.httpClient.Do(httpRequest) // <-- make formated http request
	//fmt.Println(httpResponse.Header["X-Ratelimit-Method-Remaining"])
	if err != nil {
		return fmt.Errorf("rpc call %v(): %w", httpRequest.Method, err)
	}
	defer httpResponse.Body.Close()

	return callback(httpRequest, httpResponse)
}

func (client *rpcClient) makeCall(ctx context.Context, RPCRequest *RPCRequest) (*RPCResponse, error) {
	var finalRpcResponse *RPCResponse
	err := client.makeCallWithCallbackOnHTTPResponse(
		ctx,
		RPCRequest,
		func(httpRequest *http.Request, httpResponse *http.Response) error { // <- defined function as an argument (to test for errors) I
			decoder := json.NewDecoder(httpResponse.Body) // creates a new insance of json decoder based on the body response
			decoder.DisallowUnknownFields()
			decoder.UseNumber() // unmarshals any floats into interfaces
			err := decoder.Decode(&finalRpcResponse)

			if err != nil {
				if httpResponse.StatusCode >= 400 {
					return &HTTPError{
						Code: httpResponse.StatusCode,
						err:  fmt.Errorf("rpc call %v() on %v status code %v: couldn't decode body to rpc resposne %w", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err),
					}
				}
				return fmt.Errorf("rpc call %v() on %v status code %v: couldn't decode body to rpc resposne %w", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err)
			}

			//rpc body is empty
			if finalRpcResponse == nil {
				if httpResponse.StatusCode >= 400 {
					return &HTTPError{
						Code: httpResponse.StatusCode,
						err:  fmt.Errorf("rpc call %v() on %v status code: %v. rpc response missing: %w", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err),
					}
				}
				return fmt.Errorf("rpc call %v() on %v status code: %v. rpc response missing: %w", httpRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return finalRpcResponse, nil
}

func (client *rpcClient) CallForInfo(ctx context.Context, out interface{}, method string, params []interface{}) error {
	// Crafting the request
	request := &RPCRequest{
		JSONRPC: jsonrpcVersion,
		Method:  method,
	}
	if params != nil {
		request.Params = params
	}

	rpcResponse, err := client.makeCall(ctx, request)
	if err != nil {
		return err
	}
	if rpcResponse.Error != nil {
		return rpcResponse.Error
	}

	return rpcResponse.GetObject(out)
}

func (client *rpcClient) Call(ctx context.Context, method string, params ...interface{}) (*RPCResponse, error) {
	request := &RPCRequest{
		JSONRPC: jsonrpcVersion,
		Id:      newID(),
		Method:  method,
		Params:  Params(params...),
	}
	return client.makeCall(ctx, request)
}

// Structuring params into a slice
func Params(params ...interface{}) interface{} {
	var finalParams interface{}
	if params != nil {
		switch len(params) {
		case 0:
		case 1:
			if params[0] != nil {
				var typeOf reflect.Type

				// Traversing to the underling type of a pointer (by derefrencing the pointers)
				for typeOf = reflect.TypeOf(params[0]); typeOf != nil && typeOf.Kind() == reflect.Ptr; typeOf.Elem() {
				}

				if typeOf != nil {
					switch typeOf.Kind() {
					case reflect.Struct, reflect.Array, reflect.Slice, reflect.Interface, reflect.Map:
						finalParams = params[0]
					default:
						finalParams = params
					}
				}
			} else {
				finalParams = params
			}
		default:
			finalParams = params
		}
	}
	return finalParams
}

var useIntegerID = false
var integerID = new(atomic.Uint64)

func newID() any {
	if useIntegerID {
		return integerID.Add(1) // gives a unique id
	} else {
		return 1 // fixed id
	}
}

/*
Converts an RPC response to whatever type you passsed as an argument
toType must be a pointer so that unmarshaling can modify its pointer
*/
func (RPCResponse *RPCResponse) GetObject(toType interface{}) error {
	if RPCResponse == nil {
		return errors.New("rpc response is nil")
	}
	rv := reflect.ValueOf(toType)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer got a value instead: %v", reflect.TypeOf(toType))
	}
	if RPCResponse.Result == nil {
		RPCResponse.Result = []byte(`null`)
	}

	return json.Unmarshal(RPCResponse.Result, toType)
}

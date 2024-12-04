package rpc
import(
  "context"
  "time"
  "net/http"
  "net"
  jsonrpc "go_projects/rpc/jsonrpc"
)

type JSONRPCClient interface{
  Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error)
  CallForInfo(ctx context.Context, out interface{}, method string, params []interface{}) error
}

type Client struct{
  rpcURL    string
  rpcClient JSONRPCClient // -> abstraction over rpcClient from jsonrpc
}

var (
  defaultTimeout =  time.Minute * 5
)

func (c *Client) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error){
  return c.rpcClient.Call(ctx, method, params)
}

// returns a new http client from the provided config 
func newHTTP() *http.Client{
  tr := newHTTPTransport()
  return &http.Client{
    Timeout: defaultTimeout,
    Transport: tr,
  }
}

func newHTTPTransport() *http.Transport{
  return &http.Transport{
    IdleConnTimeout: defaultTimeout,
    MaxConnsPerHost: 9,
    MaxIdleConnsPerHost: 9,
    Proxy: http.ProxyFromEnvironment,
    DialContext: (&net.Dialer{
      Timeout: time.Minute * 5 ,
      KeepAlive: time.Second * 180,
      DualStack: true, //enables ipv4, ipv6
    }).DialContext, 
    ForceAttemptHTTP2: true,
    TLSHandshakeTimeout: time.Second * 10,
  }
}

func New(rpcEndpoint string) *Client{
  opts := &jsonrpc.RPCClientOpts{
    HTTPClient: newHTTP(),
//  CustomHeaders: omitted
  }
  rpc_client := jsonrpc.NewClientWithOpts(rpcEndpoint, opts) // receives a pointer to an jsonrpc.rpcClient
  return &Client{rpcClient: rpc_client,} // creates a new Solana rpc client with the provided rpc client
}


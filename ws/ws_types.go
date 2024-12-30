package ws

import (
	stdjson "encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type request struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
}

type response struct {
	Version string              `json:"jsonrpc"`
	Params  *params             `json:"params"`
	Error   *stdjson.RawMessage `json:"error"`
}

type params struct {
	Result       *stdjson.RawMessage `json:"result"`
	Subscription int                 `json:"subscription"`
}

type Options struct {
	HttpHeader       http.Header
	HandshakeTimeout time.Duration
	ShortID          bool // some RPC do not support int63/uint64 id, so need to enable it to rand a int31/uint32 id
}

var DefaultHandshakeTimeout = 45 * time.Second

func newRequest(params []interface{}, method string, conf map[string]interface{}, shortID bool) *request {
	if params != nil && conf != nil {
		params = append(params, conf)
	}
	var ID uint64
	if !shortID {
		ID = uint64(rand.Int63())
	} else {
		ID = uint64(rand.Int31())
	}
  // This is Solana's payload for a subscription
	return &request{
		Version: "2.0",
		Method:  method,
		Params:  params,
		ID:      ID,
	}
}

func (r *request) encode() ([]byte, error) {
	data, err := stdjson.Marshal(r) 
	if err != nil {
		return nil, fmt.Errorf("encode request: json marshal: %w", err)
	}
	return data, nil
}

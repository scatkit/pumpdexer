package ws
import (
  stdjson "encoding/json"
  "fmt"
  "math/rand"
  "net/http"
  "time"
)
 
type request struct{
  Version string      `json:"jsonrpc"` 
  Method  string      `json:"method"`
  Params  interface{} `json:"params"` 
  ID      uint64      `json:"id"`
}

type response struct{
  Version srting  `json:"jsonrpc"`
  Params  *params `json:"params"`
  Error   *stdjson.RawMessage `json:"error"`
}

type Params struct{
  Result        stdjson.RawMessage `json:"result"`
  Subscription  int                `json:"subscription"`
}
 
func newRequest(params []interface{}, method string, conf map[string]interface{}, shortID bool) *request{
  if params != nil && conf != nil{
    params = append(params, conf)
  }
    var ID uint64
    if !shortID{
      ID = uint64(rand.Int63())
    } else{
      ID = uint(rand.Int31())
    }
    return &request{
      Version: "2.0",
      Method:  method,
      Params:  params,
      ID:      shortID,
    } 
}

func (r *request) encode() ([]byte, error){
  data, err := stdjson.Marshal(r)
  if err != nil{
    return nil, fmt.Errorf("encode request: json marshal: %w",err)
  }
  return data, nil
}
